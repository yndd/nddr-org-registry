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

const (
	// DeploymentFinalizer is the name of the finalizer added to
	// Deployment to block delete operations until the physical node can be
	// deprovisioned.
	DeploymentFinalizer string = "Deployment.org.nddr.yndd.io"
)

type NddrOrgDeployment struct {
	Register                  []*nddov1.Register                `json:"register,omitempty"`
	AddressAllocationStrategy *nddov1.AddressAllocationStrategy `json:"address-allocation-strategy,omitempty"`
	State                     *NddrOrgDeploymentState           `json:"state,omitempty"`
}

type NddrOrgDeploymentState struct {
	Reason *string `json:"reason,omitempty"`
	Status *string `json:"status,omitempty"`
}

// Deployment struct
type OrgDeployment struct {
	// +kubebuilder:validation:Enum=`disable`;`enable`
	// +kubebuilder:default:="enable"
	AdminState *string `json:"admin-state,omitempty"`
	// kubebuilder:validation:MinLength=1
	// kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="[A-Za-z0-9 !@#$^&()|+=`~.,'/_:;?-]*"
	Description *string `json:"description,omitempty"`
	Region      *string `json:"region,omitempty"`
	// +kubebuilder:validation:Enum=`dc`;`wan`
	// +kubebuilder:default:="dc"
	Kind                      *string                           `json:"kind,omitempty"`
	Register                  []*nddov1.Register                `json:"register,omitempty"`
	AddressAllocationStrategy *nddov1.AddressAllocationStrategy `json:"address-allocation-strategy,omitempty"`
}

// A DeploymentSpec defines the desired state of a Deployment.
type DeploymentSpec struct {
	//nddv1.ResourceSpec `json:",inline"`
	Deployment *OrgDeployment `json:"deployment,omitempty"`
}

// A DeploymentStatus represents the observed state of a Deployment.
type DeploymentStatus struct {
	nddv1.ConditionedStatus `json:",inline"`
	Deployment              *NddrOrgDeployment `json:"deployment,omitempty"`
}

// +kubebuilder:object:root=true

// Deployment is the Schema for the Deployment API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SYNC",type="string",JSONPath=".status.conditions[?(@.kind=='Synced')].status"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.conditions[?(@.kind=='Ready')].status"
// +kubebuilder:printcolumn:name="IPAM",type="string",JSONPath=".status.deployment.register[?(@.kind=='ipam')].name"
// +kubebuilder:printcolumn:name="NI",type="string",JSONPath=".status.deployment.register[?(@.kind=='network-instance')].name"
// +kubebuilder:printcolumn:name="AS",type="string",JSONPath=".status.deployment.register[?(@.kind=='as')].name"
// +kubebuilder:printcolumn:name="EPG",type="string",JSONPath=".status.deployment.register[?(@.kind=='endpoint-group')].name"
// +kubebuilder:printcolumn:name="VLAN",type="string",JSONPath=".status.deployment.register[?(@.kind=='vlan')].name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type Deployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeploymentSpec   `json:"spec,omitempty"`
	Status DeploymentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DeploymentList contains a list of Deployments
type DeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Deployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Deployment{}, &DeploymentList{})
}

// Deployment type metadata.
var (
	DeploymentKindKind         = reflect.TypeOf(Deployment{}).Name()
	DeploymentGroupKind        = schema.GroupKind{Group: Group, Kind: DeploymentKindKind}.String()
	DeploymentKindAPIVersion   = DeploymentKindKind + "." + GroupVersion.String()
	DeploymentGroupVersionKind = GroupVersion.WithKind(DeploymentKindKind)
)
