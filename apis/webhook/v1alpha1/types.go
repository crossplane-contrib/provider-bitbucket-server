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

	"github.com/crossplane-contrib/provider-bitbucket-server/internal/clients/bitbucket"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

/*
https://docs.atlassian.com/bitbucket-server/rest/7.10.0/bitbucket-rest.html

    "active": true
    }
*/

// WebhookParameters are the configurable fields of a Webhook.
type WebhookParameters struct {
	// The project key is the short name for the project for a
	// repository. Typically the key for a project called "Foo Bar"
	// would be "FB".
	// +immutable
	ProjectKey string `json:"projectKey"`

	// The repoName is the name of the git repository.
	// +immutable
	RepoName string `json:"repoName"`

	Webhook BitbucketWebhook `json:"webhook"`
}

// BitbucketWebhook provide a way to configure Bitbucket Server to make requests
// to your server (or another external service) whenever certain events occur in Bitbucket
type BitbucketWebhook struct {
	Name string `json:"name"`

	Configuration BitbucketWebhookConfiguration `json:"configuration"`

	Events []Event `json:"events"`

	URL string `json:"url"`

	// active bool
}

// TODO: Look up all options

// Event describes a bitbucket server event type
// +kubebuilder:validation:Enum="repo:refs_changed";"repo:modified"
type Event string

// BitbucketWebhookConfiguration configures settings for a webhook configuration
type BitbucketWebhookConfiguration struct {
	// Webhook secret. Leave empty to get a secret in the connection details
	// +kubebuilder:validation:Optional
	Secret string `json:"secret"`
	// TODO: ref as an option
	// TODO: Generate as an option, output connection secret
}

// WebhookObservation are the observable fields of an Webhook.
type WebhookObservation struct {
	// Consider stats here?
	ID int `json:"id,omitempty"`
}

// An WebhookSpec defines the desired state of an Webhook.
type WebhookSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       WebhookParameters `json:"forProvider"`
}

// An WebhookStatus represents the observed state of an Webhook.
type WebhookStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          WebhookObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An Webhook is an SSH key with read or write access to a bitbucket git repo.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type Webhook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebhookSpec   `json:"spec"`
	Status WebhookStatus `json:"status,omitempty"`
}

// Repo returns current repository for webhook
// TODO: Move
func (a Webhook) Repo() bitbucket.Repo {
	return bitbucket.Repo{
		ProjectKey: a.Spec.ForProvider.ProjectKey,
		Repo:       a.Spec.ForProvider.RepoName,
	}
}

// Webhook returns the bitbucket rest client of the object
// TODO: Move
func (a Webhook) Webhook() bitbucket.Webhook {
	events := make([]string, 0, len(a.Spec.ForProvider.Webhook.Events))
	for _, ev := range a.Spec.ForProvider.Webhook.Events {
		events = append(events, string(ev))
	}
	return bitbucket.Webhook{
		// ID: get from CR? meta.GetExternalName?

		Name:          a.Spec.ForProvider.Webhook.Name,
		Configuration: a.Spec.ForProvider.Webhook.Configuration,
		Events:        events,
		URL:           a.Spec.ForProvider.Webhook.URL,
	}
}

// +kubebuilder:object:root=true

// WebhookList contains a list of Webhook
type WebhookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Webhook `json:"items"`
}
