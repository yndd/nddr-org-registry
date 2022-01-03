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
	"github.com/yndd/ndd-runtime/pkg/resource"
	"github.com/yndd/ndd-runtime/pkg/utils"
	nddov1 "github.com/yndd/nddo-runtime/apis/common/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ DpList = &DeploymentList{}

// +k8s:deepcopy-gen=false
type DpList interface {
	client.ObjectList

	GetDeployments() []Dp
}

func (x *DeploymentList) GetDeployments() []Dp {
	xs := make([]Dp, len(x.Items))
	for i, r := range x.Items {
		r := r // Pin range variable so we can take its address.
		xs[i] = &r
	}
	return xs
}

var _ Dp = &Deployment{}

// +k8s:deepcopy-gen=false
type Dp interface {
	resource.Object
	resource.Conditioned

	GetOrganizationName() string
	GetDeploymentName() string
	GetAdminState() string
	GetDescription() string
	GetKind() string
	GetRegion() string
	GetRegister() map[string]string
	GetAddressAllocationStrategy() *nddov1.AddressAllocationStrategy
	InitializeResource() error

	SetStatus(string)
	SetReason(string)
	GetStatus() string
	GetStateRegister() map[string]string
	SetStateRegister(map[string]string)
	GetStateAddressAllocationStrategy() *nddov1.AddressAllocationStrategy
	SetStateAddressAllocationStrategy(*nddov1.AddressAllocationStrategy)
}

// GetCondition of this Network Node.
func (x *Deployment) GetCondition(ct nddv1.ConditionKind) nddv1.Condition {
	return x.Status.GetCondition(ct)
}

// SetConditions of the Network Node.
func (x *Deployment) SetConditions(c ...nddv1.Condition) {
	x.Status.SetConditions(c...)
}

func (x *Deployment) GetOrganizationName() string {
	return x.GetNamespace()
}

func (x *Deployment) GetDeploymentName() string {
	return x.GetName()
}

func (x *Deployment) GetAdminState() string {
	if reflect.ValueOf(x.Spec.Deployment.AdminState).IsZero() {
		return ""
	}
	return *x.Spec.Deployment.AdminState
}

func (x *Deployment) GetDescription() string {
	if reflect.ValueOf(x.Spec.Deployment.Description).IsZero() {
		return ""
	}
	return *x.Spec.Deployment.Description
}

func (x *Deployment) GetKind() string {
	if reflect.ValueOf(x.Spec.Deployment.Kind).IsZero() {
		return ""
	}
	return *x.Spec.Deployment.Kind
}

func (x *Deployment) GetRegion() string {
	if reflect.ValueOf(x.Spec.Deployment.Region).IsZero() {
		return ""
	}
	return *x.Spec.Deployment.Region
}

func (x *Deployment) GetRegister() map[string]string {
	s := make(map[string]string)
	if reflect.ValueOf(x.Spec.Deployment.Register).IsZero() {
		return s
	}
	for _, register := range x.Spec.Deployment.Register {
		for kind, name := range register.GetRegister() {
			s[kind] = name
		}
	}
	return s
}

func (x *Deployment) GetAddressAllocationStrategy() *nddov1.AddressAllocationStrategy {
	if reflect.ValueOf(x.Spec.Deployment.AddressAllocationStrategy).IsZero() {
		return &nddov1.AddressAllocationStrategy{}
	}
	return x.Spec.Deployment.AddressAllocationStrategy
}

func (x *Deployment) InitializeResource() error {
	if x.Status.Deployment != nil {
		// resource was already initialiazed
		// copy the spec, but not the state
		return nil
	}

	x.Status.Deployment = &NddrOrgDeployment{
		Register:                  make([]*nddov1.Register, 0),
		AddressAllocationStrategy: &nddov1.AddressAllocationStrategy{},
		State: &NddrOrgDeploymentState{
			Status: utils.StringPtr(""),
			Reason: utils.StringPtr(""),
		},
	}
	return nil
}

func (x *Deployment) SetStatus(s string) {
	x.Status.Deployment.State.Status = &s
}

func (x *Deployment) SetReason(s string) {
	x.Status.Deployment.State.Reason = &s
}

func (x *Deployment) GetStatus() string {
	if x.Status.Deployment != nil && x.Status.Deployment.State != nil && x.Status.Deployment.State.Status != nil {
		return *x.Status.Deployment.State.Status
	}
	return "unknown"
}

func (x *Deployment) GetStateRegister() map[string]string {
	r := make(map[string]string)
	if x.Status.Deployment != nil && x.Status.Deployment.State != nil && x.Status.Deployment.State.Status != nil {
		for _, register := range x.Status.Deployment.Register {
			for kind, name := range register.GetRegister() {
				r[kind] = name
			}
		}
	}
	return r
}

func (x *Deployment) SetStateRegister(r map[string]string) {
	x.Status.Deployment.Register = make([]*nddov1.Register, 0, len(r))
	for kind, name := range r {
		x.Status.Deployment.Register = append(x.Status.Deployment.Register, &nddov1.Register{
			Kind: utils.StringPtr(kind),
			Name: utils.StringPtr(name),
		})
	}
}

func (x *Deployment) GetStateAddressAllocationStrategy() *nddov1.AddressAllocationStrategy {
	if x.Status.Deployment != nil {
		return x.Status.Deployment.AddressAllocationStrategy
	}
	return &nddov1.AddressAllocationStrategy{}
}

func (x *Deployment) SetStateAddressAllocationStrategy(a *nddov1.AddressAllocationStrategy) {
	x.Status.Deployment.AddressAllocationStrategy = a
}
