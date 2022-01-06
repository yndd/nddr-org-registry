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

package deployment

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/yndd/ndd-runtime/pkg/event"
	"github.com/yndd/ndd-runtime/pkg/logging"
	nddov1 "github.com/yndd/nddo-runtime/apis/common/v1"
	"github.com/yndd/nddo-runtime/pkg/reconciler/managed"
	"github.com/yndd/nddo-runtime/pkg/resource"
	orgv1alpha1 "github.com/yndd/nddr-org-registry/apis/org/v1alpha1"
	"github.com/yndd/nddr-org-registry/internal/handler"
	"github.com/yndd/nddr-org-registry/internal/shared"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	// timers
	reconcileTimeout = 1 * time.Minute
	shortWait        = 5 * time.Second
	veryShortWait    = 1 * time.Second
	// errors
	errUnexpectedResource = "unexpected deployment object"
	errGetK8sResource     = "cannot get deployment resource"
)

// Setup adds a controller that reconciles infra.
func Setup(mgr ctrl.Manager, o controller.Options, nddcopts *shared.NddControllerOptions) error {
	name := "nddo/" + strings.ToLower(orgv1alpha1.DeploymentGroupKind)
	depfn := func() orgv1alpha1.Dp { return &orgv1alpha1.Deployment{} }
	deplfn := func() orgv1alpha1.DpList { return &orgv1alpha1.DeploymentList{} }
	orglfn := func() orgv1alpha1.OrgList { return &orgv1alpha1.OrganizationList{} }

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(orgv1alpha1.DeploymentGroupVersionKind),
		managed.WithLogger(nddcopts.Logger.WithValues("controller", name)),
		managed.WithApplication(&application{
			client: resource.ClientApplicator{
				Client:     mgr.GetClient(),
				Applicator: resource.NewAPIPatchingApplicator(mgr.GetClient()),
			},
			log:        nddcopts.Logger.WithValues("applogic", name),
			newDep:     depfn,
			newOrgList: orglfn,
			handler:    nddcopts.Handler,
		}),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)

	orgHandler := &EnqueueRequestForAllOrganizations{
		client:     mgr.GetClient(),
		log:        nddcopts.Logger,
		ctx:        context.Background(),
		newDepList: deplfn,
		handler:    nddcopts.Handler,
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&orgv1alpha1.Deployment{}).
		Owns(&orgv1alpha1.Deployment{}).
		WithEventFilter(resource.IgnoreUpdateWithoutGenerationChangePredicate()).
		Watches(&source.Kind{Type: &orgv1alpha1.Organization{}}, orgHandler).
		Complete(r)

}

type application struct {
	client resource.ClientApplicator
	log    logging.Logger

	newDep     func() orgv1alpha1.Dp
	newOrgList func() orgv1alpha1.OrgList

	handler handler.Handler
}

func getCrName(cr orgv1alpha1.Dp) string {
	return strings.Join([]string{cr.GetNamespace(), cr.GetName()}, ".")
}

func (r *application) Initialize(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*orgv1alpha1.Deployment)
	if !ok {
		return errors.New(errUnexpectedResource)
	}

	if err := cr.InitializeResource(); err != nil {
		r.log.Debug("Cannot initialize", "error", err)
		return err
	}

	return nil
}

func (r *application) Update(ctx context.Context, mg resource.Managed) (map[string]string, error) {
	cr, ok := mg.(*orgv1alpha1.Deployment)
	if !ok {
		return nil, errors.New(errUnexpectedResource)
	}

	return r.handleAppLogic(ctx, cr)
}

func (r *application) FinalUpdate(ctx context.Context, mg resource.Managed) {
}

func (r *application) Timeout(ctx context.Context, mg resource.Managed) time.Duration {
	/*
		cr, _ := mg.(*orgv1alpha1.Deployment)
		crName := getCrName(cr)
		speedy := r.handler.GetSpeedy(crName)
		if speedy <= 2 {
			r.handler.IncrementSpeedy(crName)
			r.log.Debug("Speedy incr", "number", r.handler.GetSpeedy(crName))
			switch speedy {
			case 0:
				return veryShortWait
			case 1, 2:
				return shortWait
			}

		}
	*/
	return reconcileTimeout
}

func (r *application) Delete(ctx context.Context, mg resource.Managed) (bool, error) {
	_, ok := mg.(*orgv1alpha1.Deployment)
	if !ok {
		return true, errors.New(errUnexpectedResource)
	}
	//if err := r.handler.DeleteDeploymentNamespace(ctx, cr); err != nil {
	//	return true, err
	//}
	return true, nil
}

func (r *application) FinalDelete(ctx context.Context, mg resource.Managed) {
	cr, ok := mg.(*orgv1alpha1.Deployment)
	if !ok {
		return
	}
	crName := getCrName(cr)
	r.handler.Delete(crName)
}

func (r *application) handleAppLogic(ctx context.Context, cr orgv1alpha1.Dp) (map[string]string, error) {
	log := r.log.WithValues("function", "handleAppLogic", "crname", cr.GetName())
	log.Debug("handleAppLogic")

	// initialize speedy
	crName := getCrName(cr)
	r.handler.Init(crName)

	orgs := r.newOrgList()
	if err := r.client.List(ctx, orgs); err != nil {
		return nil, err
	}

	orgfound := false
	orgRegister := make(map[string]string)
	var orgAddressAllocationStrategy *nddov1.AddressAllocationStrategy
	for _, org := range orgs.GetOrganizations() {
		log.Debug("org matches", "orgname", org.GetName(), "depNamespace", cr.GetNamespace())
		if org.GetOrganizationName() == cr.GetOrganizationName() {
			orgfound = true
			orgRegister = org.GetRegister()
			orgAddressAllocationStrategy = org.GetAddressAllocationStrategy()
			break
		}
	}
	if !orgfound {
		cr.SetStatus("down")
		cr.SetReason("organization not found")
		cr.SetStateRegister(make(map[string]string))
		return nil, errors.New("organization not found")
	}

	//if err := r.handler.CreateDeploymentNamespace(ctx, cr); err != nil {
	//	return make(map[string]string), err
	//}

	if cr.GetAdminState() == "disable" {
		cr.SetStatus("down")
		cr.SetReason("admin state disabled")
		cr.SetStateRegister(make(map[string]string))
	} else {
		cr.SetStatus("up")
		cr.SetReason("")
		depRegister := getDeploymentRegister(orgRegister, cr.GetRegister())
		cr.SetStateRegister(depRegister)
		aas := getDeploymentAddresssAllocationStrategy(orgAddressAllocationStrategy, cr.GetAddressAllocationStrategy())
		cr.SetStateAddressAllocationStrategy(aas)
	}
	return make(map[string]string), nil
}

func getDeploymentRegister(orgRegister, depRegister map[string]string) map[string]string {
	for orgKind, orgName := range orgRegister {
		if _, ok := depRegister[orgKind]; !ok {
			depRegister[orgKind] = orgName
		}
	}
	return depRegister
}

func getDeploymentAddresssAllocationStrategy(orgass, depaas *nddov1.AddressAllocationStrategy) *nddov1.AddressAllocationStrategy {
	if depaas != nil {
		return depaas
	}
	return orgass
}
