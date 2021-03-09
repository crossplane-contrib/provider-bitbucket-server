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
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/crossplane/provider-bitbucket-server/apis/webhook/v1alpha1"
	"github.com/crossplane/provider-bitbucket-server/internal/clients/bitbucket"
	"github.com/crossplane/provider-bitbucket-server/internal/clients/bitbucket/fake"
)

type resourceModifier func(*v1alpha1.Webhook)

func withConditions(c ...xpv1.Condition) resourceModifier {
	return func(r *v1alpha1.Webhook) { r.Status.ConditionedStatus.Conditions = c }
}

func withExternalName(id int) resourceModifier {
	return func(r *v1alpha1.Webhook) { meta.SetExternalName(r, fmt.Sprint(id)) }
}

const (
	namespace = "cool-namespace"

	connectionSecretName = "cool-connection-secret"
)

func instance(rm ...resourceModifier) *v1alpha1.Webhook {
	r := &v1alpha1.Webhook{
		Spec: v1alpha1.WebhookSpec{
			ResourceSpec: xpv1.ResourceSpec{
				WriteConnectionSecretToReference: &xpv1.SecretReference{
					Namespace: namespace,
					Name:      connectionSecretName,
				},
			},
			ForProvider: v1alpha1.WebhookParameters{
				ProjectKey: "proj",
				RepoName:   "repo",
				Webhook: v1alpha1.BitbucketWebhook{
					Name: "name",
					Configuration: v1alpha1.BitbucketWebhookConfiguration{
						Secret: "123",
					},
					Events: []string{
						"repo:refs_changed",
						"repo:modified",
					},
					URL: "https://example.com",
				},
			},
		},
	}
	// active bool

	for _, m := range rm {
		m(r)
	}

	return r
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type args struct {
		cr *v1alpha1.Webhook
		r  bitbucket.WebhookClientAPI
	}
	type want struct {
		cr  *v1alpha1.Webhook
		o   managed.ExternalObservation
		err error
	}

	errorBoom := errors.New("error")

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr: instance(withExternalName(99)),
				r: &fake.MockWebhookClient{
					MockGetWebhook: func(_ context.Context, repo bitbucket.Repo, id int) (result bitbucket.Webhook, err error) {
						return bitbucket.Webhook{
							// TODO
						}, nil
					},
				},
			},
			want: want{
				cr: instance(withExternalName(99)),
				o: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: managed.ConnectionDetails{
						/*						xpv1.ResourceCredentialsSecretEndpointKey: []byte(hostName),*/
					},
				},
			},
		},
		// TODO: What about immutable field changed?
		"NotUpToDate": {
			args: args{
				cr: instance(withExternalName(99)),
				r: &fake.MockWebhookClient{
					MockGetWebhook: func(_ context.Context, repo bitbucket.Repo, id int) (result bitbucket.Webhook, err error) {
						return bitbucket.Webhook{
							// TODO
						}, nil
					},
				},
			},
			want: want{
				cr: instance(withExternalName(99)),
				o: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: managed.ConnectionDetails{},
				},
			},
		},
		"NoExternalName": {
			args: args{
				cr: instance(),
			},
			want: want{
				cr: instance(),
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"GetFailed": {
			args: args{
				cr: instance(withExternalName(99)),
				r: &fake.MockWebhookClient{
					MockGetWebhook: func(_ context.Context, repo bitbucket.Repo, id int) (result bitbucket.Webhook, err error) {
						return bitbucket.Webhook{}, errorBoom
					},
				},
			},
			want: want{
				cr:  instance(withExternalName(99)),
				err: errors.Wrap(errorBoom, errGetFailed),
			},
		},
		"NotFound": {
			args: args{
				cr: instance(withExternalName(99)),
				r: &fake.MockWebhookClient{
					MockGetWebhook: func(_ context.Context, repo bitbucket.Repo, id int) (result bitbucket.Webhook, err error) {
						return bitbucket.Webhook{}, bitbucket.ErrNotFound
					},
				},
			},
			want: want{
				cr: instance(withExternalName(99)),
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{
				service: tc.r,
			}
			o, err := e.Observe(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.cr, tc.args.cr); diff != "" {
				t.Errorf("Observe(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Observe(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.o, o); diff != "" {
				t.Errorf("Observe(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		cr *v1alpha1.Webhook
		r  bitbucket.WebhookClientAPI
	}
	type want struct {
		cr  *v1alpha1.Webhook
		o   managed.ExternalUpdate
		err error
	}

	errorBoom := errors.New("error")

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr: instance(withExternalName(99)),
				r:  &fake.MockWebhookClient{
					// TODO
				},
			},
			want: want{
				cr: instance(withExternalName(99)),
				o: managed.ExternalUpdate{
					ConnectionDetails: managed.ConnectionDetails{},
				},
			},
		},
		"Failed": {
			args: args{
				cr: instance(withExternalName(99)),
				r:  &fake.MockWebhookClient{
					// TODO
				},
			},
			want: want{
				cr:  instance(withExternalName(99)),
				o:   managed.ExternalUpdate{},
				err: errors.Wrap(errorBoom, errUpdateFailed),
			},
		},

		/*		"NoExternalName": {
					args: args{
						cr: instance(),
					},
					want: want{
						cr: instance(),
						o: managed.ExternalObservation{
							ResourceExists: false,
						},
					},
				},
			},*/
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{
				service: tc.r,
			}
			o, err := e.Update(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.cr, tc.args.cr); diff != "" {
				t.Errorf("Update(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Update(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.o, o); diff != "" {
				t.Errorf("Update(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		cr *v1alpha1.Webhook
		r  bitbucket.WebhookClientAPI
	}
	type want struct {
		cr  *v1alpha1.Webhook
		err error
	}

	errorBoom := errors.New("error")

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr: instance(withExternalName(99)),
				r: &fake.MockWebhookClient{
					MockDeleteWebhook: func(_ context.Context, repo bitbucket.Repo, id int) error {
						if id != 99 {
							t.Errorf("unexpected id: %v", id)
						}
						return nil
					},
				},
			},
			want: want{
				cr: instance(withExternalName(99), withConditions(xpv1.Deleting())), // TODO clear external name?
			},
		},
		/*		"NoExternalName": {
				args: args{
					cr: instance(),
				},
				want: want{
					cr: instance(),
					o: managed.ExternalObservation{
						ResourceExists: false,
					},
				},
			},*/
		"DeleteFailed": {
			args: args{
				cr: instance(withExternalName(99)),
				r: &fake.MockWebhookClient{
					MockDeleteWebhook: func(_ context.Context, repo bitbucket.Repo, id int) error {
						if id != 99 {
							t.Errorf("unexpected id: %v", id)
						}
						return errorBoom
					},
				},
			},
			want: want{
				cr:  instance(withExternalName(99)),
				err: errors.Wrap(errorBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{
				service: tc.r,
			}
			err := e.Delete(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.cr, tc.args.cr); diff != "" {
				t.Errorf("Delete(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(...): -want, +got\n%s", diff)
			}
		})
	}
}
