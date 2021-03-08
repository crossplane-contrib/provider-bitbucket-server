/*
Copyright 2020 The Crossplane Authors.

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
)

// AccessKeyParameters are the configurable fields of a AccessKey.
type AccessKeyParameters struct {
	ConfigurableField string `json:"configurableField"`
}

// AccessKeyObservation are the observable fields of a AccessKey.
type AccessKeyObservation struct {
	ObservableField string `json:"observableField,omitempty"`
}

// A AccessKeySpec defines the desired state of a AccessKey.
type AccessKeySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       AccessKeyParameters `json:"forProvider"`
}

// A AccessKeyStatus represents the observed state of a AccessKey.
type AccessKeyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          AccessKeyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A AccessKey is an example API type
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type AccessKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccessKeySpec   `json:"spec"`
	Status AccessKeyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AccessKeyList contains a list of AccessKey
type AccessKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessKey `json:"items"`
}
