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

package registry

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	pkgmetav1 "github.com/yndd/ndd-core/apis/pkg/meta/v1"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/nddo-grpc/ndd"
	rclient "github.com/yndd/nddo-grpc/resource/client"
	"github.com/yndd/nddo-grpc/resource/resourcepb"
	nddov1 "github.com/yndd/nddo-runtime/apis/common/v1"
	"github.com/yndd/nddo-runtime/pkg/odns"
	orgv1alpha1 "github.com/yndd/nddr-org-registry/apis/org/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	nddNamespace     = "ndd-system"
	defaultNamespace = "default"
)

type RegisterKind string

const (
	RegisterKindIpam RegisterKind = "ipam"
	RegisterKindAs   RegisterKind = "as"
	RegisterKindNi   RegisterKind = "ni"
	RegisterKindVlan RegisterKind = "vlan"
	RegisterKindEsi  RegisterKind = "esi"
	RegisterKindRt   RegisterKind = "rt"
)

func (r RegisterKind) String() string {
	switch r {
	case RegisterKindIpam:
		return "ipam"
	case RegisterKindAs:
		return "as"
	case RegisterKindNi:
		return "ni"
	case RegisterKindVlan:
		return "vlan"
	case RegisterKindEsi:
		return "epg"
	case RegisterKindRt:
		return "rt"
	}
	return "unknown"
}

type registry struct {
	log logging.Logger
	// kubernetes
	client client.Client
}

func New(opts ...Option) Registry {
	s := &registry{}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *registry) WithLogger(log logging.Logger) {
	s.log = log
}

func (s *registry) WithClient(c client.Client) {
	s.client = c
}

/*
func (r *registry) GetRegisterName(organizationName string, deploymentName string) []string {
	registerName := make([]string, 0)
	if deploymentName == "" {
		registerName = append(registerName, organizationName)

	}
	registerName = append(registerName, organizationName)
	registerName = append(registerName, deploymentName)
	return registerName
}
*/

func (r *registry) GetRegister(ctx context.Context, namespace string, odns *odns.Odns) (map[string]string, error) {
	// critical registers are ipam and as right now since they server dynamic
	// grpc services
	criticalRegisters := []string{
		RegisterKindIpam.String(),
		RegisterKindAs.String(),
		RegisterKindNi.String(),
	}

	fullOdaName, odaKind := odns.GetFullOdaName()

	var registers map[string]string
	switch odaKind {
	case nddov1.OdaKindDeployment:
		dep := &orgv1alpha1.Deployment{}
		if err := r.client.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      fullOdaName,
		}, dep); err != nil {
			return nil, err
		}

		registers = dep.GetStateRegister()
	default:
		org := &orgv1alpha1.Organization{}
		if err := r.client.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      fullOdaName,
		}, org); err != nil {
			return nil, err
		}
		registers = org.GetStateRegister()
	}
	for _, register := range criticalRegisters {
		if _, ok := registers[register]; !ok {
			return nil, fmt.Errorf("critical register %s not found in registry", register)
		}
	}
	return registers, nil
}

func (r *registry) GetAddressAllocationStrategy(ctx context.Context, namespace string, oda *nddov1.OdaInfo) (*nddov1.AddressAllocationStrategy, error) {
	switch {
	case oda.GetDeployment() != "":
		dep := &orgv1alpha1.Deployment{}
		if err := r.client.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      oda.GetDeployment(),
		}, dep); err != nil {
			return nil, err
		}
		return dep.GetStateAddressAllocationStrategy(), nil

	default:
		org := &orgv1alpha1.Organization{}
		if err := r.client.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      oda.GetOrganization(),
		}, org); err != nil {
			return nil, err
		}
		return org.GetStateAddressAllocationStrategy(), nil
	}
}

func (r *registry) GetRegistryClient(ctx context.Context, registerName string) (resourcepb.ResourceClient, error) {
	registers := map[string]string{
		RegisterKindIpam.String(): "nddr-ipam-registry",
		RegisterKindAs.String():   "nddr-as-registry",
		RegisterKindNi.String():   "nddr-ni-registry",
	}

	if _, ok := registers[registerName]; !ok {
		return nil, fmt.Errorf("wrong register request, name not found: %s", registerName)
	}
	registerMatch := registers[registerName]

	pods := &corev1.PodList{}
	opts := []client.ListOption{
		client.InNamespace(nddNamespace),
	}
	if err := r.client.List(ctx, pods, opts...); err != nil {
		return nil, err
	}

	var podname string
	found := false
	for _, pod := range pods.Items {
		podName := pod.GetName()
		//r.log.Debug("pod", "podname", podName)
		if strings.Contains(podName, registerMatch) {
			podname = podName
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("no pod that matches %s, %s", registerName, registerMatch)
	}

	return getResourceClient(ctx, getGrpcServerName(podname))

}

func getGrpcServerName(podName string) string {
	var newName string
	for i, s := range strings.Split(podName, "-") {
		if i == 0 {
			newName = s
		} else if i <= (len(strings.Split(podName, "-")) - 3) {
			newName += "-" + s
		}
	}
	return pkgmetav1.PrefixGnmiService + "-" + newName + "." + pkgmetav1.NamespaceLocalK8sDNS + strconv.Itoa((pkgmetav1.GnmiServerPort))
}

func getResourceClient(ctx context.Context, grpcserver string) (resourcepb.ResourceClient, error) {
	cfg := &ndd.Config{
		Address:  grpcserver,
		Username: "admin",
		Password: "admin",
		//Timeout:    10 * time.Second,
		SkipVerify: true,
		Insecure:   true,
		TLSCA:      "", //TODO TLS
		TLSCert:    "",
		TLSKey:     "",
	}
	return rclient.NewClient(ctx, cfg)
}
