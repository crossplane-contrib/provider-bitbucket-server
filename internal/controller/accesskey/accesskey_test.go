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
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/crossplane/provider-bitbucket-server/apis/accesskey/v1alpha1"
	"github.com/crossplane/provider-bitbucket-server/internal/clients/bitbucket"
	"github.com/crossplane/provider-bitbucket-server/internal/clients/bitbucket/fake"
)

type resourceModifier func(*v1alpha1.AccessKey)

func withConditions(c ...xpv1.Condition) resourceModifier {
	return func(r *v1alpha1.AccessKey) { r.Status.ConditionedStatus.Conditions = c }
}

func withExternalName(id int) resourceModifier {
	return func(r *v1alpha1.AccessKey) { meta.SetExternalName(r, fmt.Sprint(id)) }
}

func withPermission(permission string) resourceModifier {
	return func(r *v1alpha1.AccessKey) { r.Spec.ForProvider.PublicKey.Permission = permission }
}

const (
	namespace = "cool-namespace"
	key1      = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDKW79iJEhqKPa6ZxeRDTh3i7h6ms4e1ABmHKfZkbyhOeC1ycMQAtteqi42oYFMscMODYqEgjgiOwi75Ol+rint7iZdXzkPDbqzHDOW4XNPzKNiqh2mOQY60n6nk8EiIIs71ff6RryxEYA2x2r3snm257o/vr4OE2F6VMmK4Io8K3TTGqsZKp8SePHnx40s8dusAtZWn7UUFedkLLHCUYAMk8gtSKcTA/ntjNdHTcIxVO5WbkZoCHPLMPc29Vz5MYq096qZ35idgCa3bSK/VSZpsNQUJEwwc04k1G9LA2z+sjD22hg79SZtY4P7knV1vvlXf5uZs+0myK9Qiwvfu3IXFWXYVr6q73VshdyM25N4C7wID4KqZTmHVLM/oQGw8jvWnWbzVwuvv+wVB1h8SBryxJsJwylCsRw8gLzpc/t0TluXQWSk2zWHHeETw83Mm0tT60mcaipCgTkbWYO+IP1OTxwsJzZtdgrrEO/Wwwk7AXRPNhiOAS5XFgZrRpj3HWU= user@example.com"
	key2      = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCpBjjwXykLFApECNzgHUOX+EhgFuFWUE/o4AQItHvuZUxqcp/ajxNXzK8Av2OyrWfJ9qvHYCpC/bOLSJfEOw5yF816t/m86TAQArEB7BhQj2mfVvFHtpg9n5f1STxu3hzWKrM0r3/R/9G/8YwFp2+6PvIvrpxmtkWuO1TEhuqRAVwdHmZ/l+8bsuQrXpaQhZ0gTTMFOMPgqkiZ5tBz4n0ocZdSI3LpsG2QuA4QYCxECcIZLzvMzqmV69+ReGJXHhX+yHwOdmtt+dvb5en0nLzbaQlYB37tGBfiaM31qXgiTd5h8tLWlgjLvnfUEOD03J887tl8OBjHLG+pa1CgBwrtKuqJirUdUhelRAfy/zkhMfFzOrPLRYu2VcKPhGV+oI8tog/ydwX62ouSN+yIxICkGf31gDVisIHILJXP2qfv8Vm7gWETfTkh9Nyrx/NbJwTuP0p2SIs94Oywwl8UpT4ytlW+BHhS6L4gUNErZKpFBnjkmCoc+h1IilJfTHmLsSc= user@example.com"
	label     = "user@example.com"

	connectionSecretName = "cool-connection-secret"
)

func instance(rm ...resourceModifier) *v1alpha1.AccessKey {
	r := &v1alpha1.AccessKey{
		Spec: v1alpha1.AccessKeySpec{
			ResourceSpec: xpv1.ResourceSpec{
				WriteConnectionSecretToReference: &xpv1.SecretReference{
					Namespace: namespace,
					Name:      connectionSecretName,
				},
			},
			ForProvider: v1alpha1.AccessKeyParameters{
				ProjectKey: "proj",
				RepoName:   "repo",
				PublicKey: v1alpha1.PublicKey{
					Label:      label,
					Key:        key1,
					Permission: bitbucket.PermissionRepoRead,
				},
			},
		},
	}

	for _, m := range rm {
		m(r)
	}

	return r
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type args struct {
		cr *v1alpha1.AccessKey
		r  bitbucket.KeyClientAPI
	}
	type want struct {
		cr  *v1alpha1.AccessKey
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
				r: &fake.MockKeyClient{
					MockGetAccessKey: func(_ context.Context, repo bitbucket.Repo, id int) (result bitbucket.AccessKey, err error) {
						return bitbucket.AccessKey{
							Key:        key1,
							Label:      label,
							ID:         id,
							Permission: bitbucket.PermissionRepoRead,
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
				r: &fake.MockKeyClient{
					MockGetAccessKey: func(_ context.Context, repo bitbucket.Repo, id int) (result bitbucket.AccessKey, err error) {
						return bitbucket.AccessKey{
							Key:        key1,
							Label:      label,
							ID:         id,
							Permission: bitbucket.PermissionRepoWrite,
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
				r: &fake.MockKeyClient{
					MockGetAccessKey: func(_ context.Context, repo bitbucket.Repo, id int) (result bitbucket.AccessKey, err error) {
						return bitbucket.AccessKey{}, errorBoom
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
				r: &fake.MockKeyClient{
					MockGetAccessKey: func(_ context.Context, repo bitbucket.Repo, id int) (result bitbucket.AccessKey, err error) {
						return bitbucket.AccessKey{}, bitbucket.ErrNotFound
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
		cr *v1alpha1.AccessKey
		r  bitbucket.KeyClientAPI
	}
	type want struct {
		cr  *v1alpha1.AccessKey
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
				cr: instance(withExternalName(99), withPermission(bitbucket.PermissionRepoWrite)),
				r: &fake.MockKeyClient{
					MockUpdateAccessKeyPermission: func(_ context.Context, repo bitbucket.Repo, id int, permission string) error {
						if id != 99 {
							t.Errorf("unexpected id: %v", id)
						}
						return nil
					},
				},
			},
			want: want{
				cr: instance(withExternalName(99), withPermission(bitbucket.PermissionRepoWrite)),
				o: managed.ExternalUpdate{
					ConnectionDetails: managed.ConnectionDetails{},
				},
			},
		},
		"Failed": {
			args: args{
				cr: instance(withExternalName(99)),
				r: &fake.MockKeyClient{
					MockUpdateAccessKeyPermission: func(_ context.Context, repo bitbucket.Repo, id int, permission string) error {
						if id != 99 {
							t.Errorf("unexpected id: %v", id)
						}
						return errorBoom
					},
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
		cr *v1alpha1.AccessKey
		r  bitbucket.KeyClientAPI
	}
	type want struct {
		cr  *v1alpha1.AccessKey
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
				r: &fake.MockKeyClient{
					MockDeleteAccessKey: func(_ context.Context, repo bitbucket.Repo, id int) error {
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
				r: &fake.MockKeyClient{
					MockDeleteAccessKey: func(_ context.Context, repo bitbucket.Repo, id int) error {
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
