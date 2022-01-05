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

	//ndddvrv1 "github.com/yndd/ndd-core/apis/dvr/v1"
	"github.com/yndd/ndd-runtime/pkg/logging"
	orgv1alpha1 "github.com/yndd/nddr-org-registry/apis/org/v1alpha1"
	"github.com/yndd/nddr-org-registry/internal/handler"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type adder interface {
	Add(item interface{})
}

type EnqueueRequestForAllOrganizations struct {
	client client.Client
	log    logging.Logger
	ctx    context.Context

	handler handler.Handler

	newDepList func() orgv1alpha1.DpList
}

// Create enqueues a request for all infrastructures which pertains to the topology.
func (e *EnqueueRequestForAllOrganizations) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	e.add(evt.Object, q)
}

// Create enqueues a request for all infrastructures which pertains to the topology.
func (e *EnqueueRequestForAllOrganizations) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	e.add(evt.ObjectOld, q)
	e.add(evt.ObjectNew, q)
}

// Create enqueues a request for all infrastructures which pertains to the topology.
func (e *EnqueueRequestForAllOrganizations) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	e.add(evt.Object, q)
}

// Create enqueues a request for all infrastructures which pertains to the topology.
func (e *EnqueueRequestForAllOrganizations) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	e.add(evt.Object, q)
}

func (e *EnqueueRequestForAllOrganizations) add(obj runtime.Object, queue adder) {
	dd, ok := obj.(*orgv1alpha1.Organization)
	if !ok {
		return
	}
	log := e.log.WithValues("function", "watch org", "name", dd.GetName(), "namespace", dd.GetNamespace())
	log.Debug("handleEvent")

	d := e.newDepList()
	if err := e.client.List(e.ctx, d); err != nil {
		return
	}

	for _, dep := range d.GetDeployments() {
		// only enqueue if the organization name match
		log.Debug("handleEvent", "depl namespace", dep.GetNamespace(), "depl name", dep.GetName(), "depl org", dep.GetOrganizationName())
		if dep.GetOrganizationName() == dd.GetName() {

			crName := getCrName(dep)
			e.handler.ResetSpeedy(crName)

			queue.Add(reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: dep.GetNamespace(),
				Name:      dep.GetName()}})
		}
	}
}
