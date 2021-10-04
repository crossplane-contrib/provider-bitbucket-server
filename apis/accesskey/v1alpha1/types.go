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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane-contrib/provider-bitbucket-server/internal/clients/bitbucket"
)

// AccessKeyParameters are the configurable fields of a AccessKey.
type AccessKeyParameters struct {
	// The project key is the short name for the project for a
	// repository. Typically the key for a project called "Foo Bar"
	// would be "FB".
	// +immutable
	ProjectKey string `json:"projectKey"`

	// The repoName is the name of the git repository.
	// +immutable
	RepoName string `json:"repoName"`

	PublicKey PublicKey `json:"publicKey"`
}

// +immutable does not make the CRD immutable
// https://discuss.kubernetes.io/t/immutable-crd/10068
// https://github.com/kubernetes/kubernetes/issues/65973
// https://crossplane.slack.com/archives/C01718T2476/p1615201920017800?thread_ts=1615199267.016100&cid=C01718T2476

// PublicKey contains the information about the public key. Only the permission field is mutable.
type PublicKey struct {
	// Label
	Label string `json:"label"`

	// The ssh-key with access to the git repo. Leave empty to get a ssh-privatekey in the connection details
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern=((ssh|ecdsa)-[a-z0-9-]+ .*|)
	Key string `json:"key,omitempty"`

	// +kubebuilder:validation:Enum=REPO_READ;REPO_WRITE
	Permission string `json:"permission"`
}

// AccessKeyObservation are the observable fields of an AccessKey.
type AccessKeyObservation struct {
	// +kubebuilder:validation:Optional
	ID int `json:"id,omitempty"`
	// +kubebuilder:validation:Optional
	Key *PublicKey `json:"publicKey,omitempty"`
}

// An AccessKeySpec defines the desired state of an AccessKey.
type AccessKeySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       AccessKeyParameters `json:"forProvider"`
}

// An AccessKeyStatus represents the observed state of an AccessKey.
type AccessKeyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          AccessKeyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An AccessKey is an SSH key with read or write access to a bitbucket git repo.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type AccessKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccessKeySpec   `json:"spec"`
	Status AccessKeyStatus `json:"status,omitempty"`
}

func (a AccessKey) Repo() bitbucket.Repo {
	return bitbucket.Repo{
		ProjectKey: a.Spec.ForProvider.ProjectKey,
		Repo:       a.Spec.ForProvider.RepoName,
	}
}

func (a AccessKey) AccessKey() bitbucket.AccessKey {
	return bitbucket.AccessKey{
		Key:        a.Spec.ForProvider.PublicKey.Key,
		Label:      a.Spec.ForProvider.PublicKey.Label,
		Permission: a.Spec.ForProvider.PublicKey.Permission,
	}
}

// +kubebuilder:object:root=true

// AccessKeyList contains a list of AccessKey
type AccessKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessKey `json:"items"`
}
