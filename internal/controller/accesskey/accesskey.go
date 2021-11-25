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

package accesskey

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
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

	"github.com/crossplane-contrib/provider-bitbucket-server/apis/accesskey/v1alpha1"
	apisv1alpha1 "github.com/crossplane-contrib/provider-bitbucket-server/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-bitbucket-server/internal/clients"
	"github.com/crossplane-contrib/provider-bitbucket-server/internal/clients/bitbucket"
	"github.com/crossplane-contrib/provider-bitbucket-server/internal/controller/config"
)

const (
	errNotAccessKey = "managed resource is not a AccessKey custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"

	errGetFailed    = "cannot get access key from bitbucket API"
	errDeleteFailed = "cannot delete access key from bitbucket API"
	errCreateFailed = "cannot create access key with bitbucket API"
	errUpdateFailed = "cannot update access permission key with bitbucket API"
)

// Setup adds a controller that reconciles AccessKey managed resources.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.AccessKeyGroupKind)

	o := controller.Options{
		RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.AccessKeyGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: clients.NewAccessKeyClient}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.AccessKey{}).
		Complete(r)
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(clients.Config) bitbucket.KeyClientAPI
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.AccessKey)
	if !ok {
		return nil, errors.New(errNotAccessKey)
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
		BaseURL:   pc.Spec.BaseURL,
		Token:     string(data),
		TLSConfig: config.NewTLSConfig(*pc),
	})

	return &external{service: svc, keygen: keygen}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	service bitbucket.KeyClientAPI
	keygen  func() (string, []byte, error)
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.AccessKey)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotAccessKey)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{}, nil
	}

	externalName := meta.GetExternalName(cr)
	id, err := strconv.Atoi(externalName)
	if err != nil {
		return managed.ExternalObservation{}, nil // nolint // This is ok as it does not exists
	}

	key, err := c.service.GetAccessKey(ctx, cr.Repo(), id)
	if err != nil {
		if errors.Is(err, bitbucket.ErrNotFound) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	cr.Status.SetConditions(xpv1.Available())

	cr.Status.AtProvider.ID = key.ID
	cr.Status.AtProvider.Key = &v1alpha1.PublicKey{
		Key:        key.Key,
		Label:      key.Label,
		Permission: key.Permission,
	}

	return managed.ExternalObservation{
		// Return false when the external resource does not exist. This lets
		// the managed resource reconciler know that it needs to call Create to
		// (re)create the resource, or that it has successfully been deleted.
		ResourceExists: true,

		// Return false when the external resource exists, but it not up to date
		// with the desired managed resource state. This lets the managed
		// resource reconciler know that it needs to call Update.
		ResourceUpToDate: key.Permission == cr.Spec.ForProvider.PublicKey.Permission,

		// Return any details that may be required to connect to the external
		// resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) create(ctx context.Context, cr *v1alpha1.AccessKey) error {
	key, err := c.service.CreateAccessKey(ctx, cr.Repo(), cr.AccessKey())
	if err != nil {
		return err
	}

	meta.SetExternalName(cr, fmt.Sprint(key.ID))
	cr.Status.SetConditions(xpv1.Available())
	cr.Status.AtProvider.ID = key.ID
	cr.Status.AtProvider.Key = &v1alpha1.PublicKey{
		Key:        key.Key,
		Label:      key.Label,
		Permission: key.Permission,
	}
	return nil
}

func keygen() (string, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", nil, err
	}
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	var private bytes.Buffer
	if err := pem.Encode(&private, privateKeyPEM); err != nil {
		return "", nil, err
	}
	// generate public key
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", nil, err
	}
	public := ssh.MarshalAuthorizedKey(pub)
	return string(public), private.Bytes(), nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.AccessKey)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotAccessKey)
	}

	cr.Status.SetConditions(xpv1.Creating())
	conndetails := managed.ConnectionDetails{}

	if cr.Spec.ForProvider.PublicKey.Key == "" {
		var err error
		var privateKey []byte
		cr.Spec.ForProvider.PublicKey.Key, privateKey, err = c.keygen()
		if err != nil {
			return managed.ExternalCreation{}, err
		}
		conndetails["ssh-privatekey"] = privateKey
	}
	if err := c.create(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails:    conndetails,
		ExternalNameAssigned: true,
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.AccessKey)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotAccessKey)
	}

	id, _ := strconv.Atoi(meta.GetExternalName(cr))
	if err := c.service.UpdateAccessKeyPermission(ctx, cr.Repo(), id, cr.Spec.ForProvider.PublicKey.Permission); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.AccessKey)
	if !ok {
		return errors.New(errNotAccessKey)
	}

	id, _ := strconv.Atoi(meta.GetExternalName(cr)) // TODO err
	if err := c.service.DeleteAccessKey(ctx, cr.Repo(), id); err != nil {
		return errors.Wrap(err, errDeleteFailed)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	return nil
}
