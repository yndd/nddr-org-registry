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
	"github.com/yndd/app-runtime/pkg/odns"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ OrgList = &OrganizationList{}

// +k8s:deepcopy-gen=false
type OrgList interface {
	client.ObjectList

	GetOrganizations() []Org
}

func (x *OrganizationList) GetOrganizations() []Org {
	xs := make([]Org, len(x.Items))
	for i, r := range x.Items {
		r := r // Pin range variable so we can take its address.
		xs[i] = &r
	}
	return xs
}

var _ Org = &Organization{}

// +k8s:deepcopy-gen=false
type Org interface {
	resource.Object
	resource.Conditioned

	GetCondition(ct nddv1.ConditionKind) nddv1.Condition
	SetConditions(c ...nddv1.Condition)

	SetHealthConditions(c nddv1.HealthConditionedStatus)

	GetDeletionPolicy() nddv1.DeletionPolicy
	SetDeletionPolicy(p nddv1.DeletionPolicy)
	GetDeploymentPolicy() nddv1.DeploymentPolicy
	SetDeploymentPolicy(p nddv1.DeploymentPolicy)

	GetTargetReference() *nddv1.Reference
	SetTargetReference(p *nddv1.Reference)

	GetRootPaths() []string
	SetRootPaths(rootPaths []string)

	GetOrganizationName() string
	GetDescription() string
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
func (x *Organization) GetCondition(ct nddv1.ConditionKind) nddv1.Condition {
	return x.Status.GetCondition(ct)
}

// SetConditions of the Network Node.
func (x *Organization) SetConditions(c ...nddv1.Condition) {
	x.Status.SetConditions(c...)
}

func (x *Organization) SetHealthConditions(c nddv1.HealthConditionedStatus) {
	x.Status.Health = c
}

func (x *Organization) GetDeletionPolicy() nddv1.DeletionPolicy {
	return x.Spec.Lifecycle.DeletionPolicy
}

func (x *Organization) SetDeletionPolicy(c nddv1.DeletionPolicy) {
	x.Spec.Lifecycle.DeletionPolicy = c
}

func (x *Organization) GetDeploymentPolicy() nddv1.DeploymentPolicy {
	return x.Spec.Lifecycle.DeploymentPolicy
}

func (x *Organization) SetDeploymentPolicy(c nddv1.DeploymentPolicy) {
	x.Spec.Lifecycle.DeploymentPolicy = c
}

func (x *Organization) GetTargetReference() *nddv1.Reference {
	return x.Spec.TargetReference
}

func (x *Organization) SetTargetReference(p *nddv1.Reference) {
	x.Spec.TargetReference = p
}

func (x *Organization) GetRootPaths() []string {
	return x.Status.RootPaths
}

func (x *Organization) SetRootPaths(rootPaths []string) {
	x.Status.RootPaths = rootPaths
}

func (x *Organization) GetOrganizationName() string {
	return odns.Name2Odns(x.GetName()).GetOrganization()
}

func (x *Organization) GetDescription() string {
	if reflect.ValueOf(x.Spec.Properties.Description).IsZero() {
		return ""
	}
	return *x.Spec.Properties.Description
}

func (x *Organization) GetRegister() map[string]string {
	s := make(map[string]string)
	if reflect.ValueOf(x.Spec.Properties.Register).IsZero() {
		return s
	}
	for _, register := range x.Spec.Properties.Register {
		for kind, name := range register.GetRegister() {
			s[kind] = name
		}
	}
	return s
}

func (x *Organization) GetAddressAllocationStrategy() *nddov1.AddressAllocationStrategy {
	if reflect.ValueOf(x.Spec.Properties.AddressAllocationStrategy).IsZero() {
		return &nddov1.AddressAllocationStrategy{}
	}
	return x.Spec.Properties.AddressAllocationStrategy
}

func (x *Organization) InitializeResource() error {
	if x.Status.Organization != nil {
		// resource was already initialiazed
		// copy the spec, but not the state
		return nil
	}

	x.Status.Organization = &NddrOrganization{
		Register:                  make([]*nddov1.Register, 0),
		AddressAllocationStrategy: &nddov1.AddressAllocationStrategy{},
		State: &NddrOrgDeploymentState{
			Status: utils.StringPtr(""),
			Reason: utils.StringPtr(""),
		},
	}
	return nil
}

func (x *Organization) SetStatus(s string) {
	x.Status.Organization.State.Status = &s
}

func (x *Organization) SetReason(s string) {
	x.Status.Organization.State.Reason = &s
}

func (x *Organization) GetStatus() string {
	if x.Status.Organization != nil && x.Status.Organization.State != nil && x.Status.Organization.State.Status != nil {
		return *x.Status.Organization.State.Status
	}
	return "unknown"
}

func (x *Organization) GetStateRegister() map[string]string {
	r := make(map[string]string)
	if x.Status.Organization != nil && x.Status.Organization.State != nil && x.Status.Organization.State.Status != nil {
		for _, register := range x.Status.Organization.Register {
			for kind, name := range register.GetRegister() {
				r[kind] = name
			}
		}
	}
	return r
}

func (x *Organization) SetStateRegister(r map[string]string) {
	x.Status.Organization.Register = make([]*nddov1.Register, 0, len(r))
	for kind, name := range r {
		x.Status.Organization.Register = append(x.Status.Organization.Register, &nddov1.Register{
			Kind: utils.StringPtr(kind),
			Name: utils.StringPtr(name),
		})
	}
}

func (x *Organization) GetStateAddressAllocationStrategy() *nddov1.AddressAllocationStrategy {
	if x.Status.Organization != nil {
		return x.Status.Organization.AddressAllocationStrategy
	}
	return &nddov1.AddressAllocationStrategy{}
}

func (x *Organization) SetStateAddressAllocationStrategy(a *nddov1.AddressAllocationStrategy) {
	x.Status.Organization.AddressAllocationStrategy = a
}
