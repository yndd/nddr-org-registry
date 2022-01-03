/*
Copyright 2021 NDD.

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
	"reflect"

	nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"
	nddov1 "github.com/yndd/nddo-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type NddrOrganization struct {
	Register                  []*nddov1.Register                `json:"register,omitempty"`
	AddressAllocationStrategy *nddov1.AddressAllocationStrategy `json:"address-allocation-strategy,omitempty"`
	State                     *NddrOrgDeploymentState           `json:"state,omitempty"`
}

type NddrOrganizationState struct {
	Reason *string `json:"reason,omitempty"`
	Status *string `json:"status,omitempty"`
}

// Organization struct
type OrgOrganization struct {
	// kubebuilder:validation:MinLength=1
	// kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[A-Za-z0-9 !@#$^&()|+=`~.,'/_:;?-]*"
	Description               *string                           `json:"description,omitempty"`
	Register                  []*nddov1.Register                `json:"register,omitempty"`
	AddressAllocationStrategy *nddov1.AddressAllocationStrategy `json:"address-allocation-strategy,omitempty"`
}

// A OrganizationSpec defines the desired state of a Organization.
type OrganizationSpec struct {
	//nddv1.ResourceSpec `json:",inline"`
	Organization *OrgOrganization `json:"organization,omitempty"`
}

// A OrganizationStatus represents the observed state of a Organization.
type OrganizationStatus struct {
	nddv1.ConditionedStatus `json:",inline"`
	Organization            *NddrOrganization `json:"organization,omitempty"`
}

// +kubebuilder:object:root=true

// Organization is the Schema for the Organization API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SYNC",type="string",JSONPath=".status.conditions[?(@.kind=='Synced')].status"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.conditions[?(@.kind=='Ready')].status"
// +kubebuilder:printcolumn:name="IPAM",type="string",JSONPath=".status.organization.register[?(@.kind=='ipam')].name"
// +kubebuilder:printcolumn:name="NI",type="string",JSONPath=".status.organization.register[?(@.kind=='network-instance')].name"
// +kubebuilder:printcolumn:name="AS",type="string",JSONPath=".status.organization.register[?(@.kind=='as')].name"
// +kubebuilder:printcolumn:name="EPG",type="string",JSONPath=".status.organization.register[?(@.kind=='endpoint-group')].name"
// +kubebuilder:printcolumn:name="VLAN",type="string",JSONPath=".status.organization.register[?(@.kind=='vlan')].name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type Organization struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrganizationSpec   `json:"spec,omitempty"`
	Status OrganizationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OrganizationList contains a list of Organizations
type OrganizationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Organization `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Organization{}, &OrganizationList{})
}

// Organization type metadata.
var (
	OrganizationKindKind         = reflect.TypeOf(Organization{}).Name()
	OrganizationGroupKind        = schema.GroupKind{Group: Group, Kind: OrganizationKindKind}.String()
	OrganizationKindAPIVersion   = OrganizationKindKind + "." + GroupVersion.String()
	OrganizationGroupVersionKind = GroupVersion.WithKind(OrganizationKindKind)
)
