/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webhook

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	apisv1alpha1 "github.com/crossplane-contrib/provider-bitbucket-server/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-bitbucket-server/apis/webhook/v1alpha1"
	"github.com/crossplane-contrib/provider-bitbucket-server/internal/clients"
	"github.com/crossplane-contrib/provider-bitbucket-server/internal/clients/bitbucket"
)

const (
	errNotWebhook   = "managed resource is not a Webhook custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"

	errNewClient = "cannot create new Service"

	errGetFailed    = "cannot get webhook from bitbucket API"
	errDeleteFailed = "cannot delete webhook from bitbucket API"
	errCreateFailed = "cannot create webhook with bitbucket API"
	errUpdateFailed = "cannot update webhook with bitbucket API"
)

// Setup adds a controller that reconciles Webhook managed resources.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.WebhookGroupKind)

	o := controller.Options{
		RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.WebhookGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			log:          l,
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: clients.NewWebhookClient}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.Webhook{}).
		Complete(r)
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        resource.Tracker
	log          logging.Logger
	newServiceFn func(clients.Config) bitbucket.WebhookClientAPI
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Webhook)
	if !ok {
		return nil, errors.New(errNotWebhook)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	cd := pc.Spec.Credentials
	data, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	svc := c.newServiceFn(clients.Config{
		BaseURL: pc.Spec.BaseURL,
		Token:   string(data),
	})

	return &external{service: svc, log: c.log, pwgen: pwgen}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	service bitbucket.WebhookClientAPI
	log     logging.Logger
	pwgen   func() (string, error)
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Webhook)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotWebhook)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{}, nil
	}

	externalName := meta.GetExternalName(cr)
	id, err := strconv.Atoi(externalName)
	if err != nil {
		return managed.ExternalObservation{}, nil // not exists
	}

	hook, err := c.service.GetWebhook(ctx, cr.Repo(), id)
	if err != nil {
		if errors.Is(err, bitbucket.ErrNotFound) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	ignoreEventOrder := cmp.Transformer("Sort", func(webhook bitbucket.Webhook) bitbucket.Webhook {
		webhook.Events = append([]string(nil), webhook.Events...) // Copy input to avoid mutating it

		sort.Strings(webhook.Events)
		return webhook
	})

	ignoreID := cmpopts.IgnoreFields(bitbucket.Webhook{}, "ID")

	diff := cmp.Diff(cr.Webhook(), hook, ignoreEventOrder, ignoreID)

	upToDate := diff == ""
	if !upToDate {
		c.log.Debug("Not up to date", "diff", diff)
	}

	return managed.ExternalObservation{
		// Return false when the external resource does not exist. This lets
		// the managed resource reconciler know that it needs to call Create to
		// (re)create the resource, or that it has successfully been deleted.
		ResourceExists: true,

		// Return false when the external resource exists, but it not up to date
		// with the desired managed resource state. This lets the managed
		// resource reconciler know that it needs to call Update.
		ResourceUpToDate: upToDate,

		// Return any details that may be required to connect to the external
		// resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func pwgen() (string, error) {
	b := make([]byte, 20)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Webhook)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotWebhook)
	}

	cr.Status.SetConditions(xpv1.Creating())

	hook := cr.Webhook()
	if hook.Configuration.Secret == "" {
		secret, err := c.pwgen()
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, "could not generate random password")
		}

		hook.Configuration.Secret = secret
	}

	key, err := c.service.CreateWebhook(ctx, cr.Repo(), hook)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, fmt.Sprint(key.ID))
	cr.Status.SetConditions(xpv1.Available())

	//	cr.Status.AtProvider.ID = key.ID TODO do we want this?

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{
			"secret": []byte(hook.Configuration.Secret),
		},
		ExternalNameAssigned: true,
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Webhook)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotWebhook)
	}

	id, _ := strconv.Atoi(meta.GetExternalName(cr))
	if _, err := c.service.UpdateWebhook(ctx, cr.Repo(), id, cr.Webhook()); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Webhook)
	if !ok {
		return errors.New(errNotWebhook)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	id, _ := strconv.Atoi(meta.GetExternalName(cr)) // TODO err
	if err := c.service.DeleteWebhook(ctx, cr.Repo(), id); err != nil {
		return errors.Wrap(err, errDeleteFailed)
	}

	return nil
}
