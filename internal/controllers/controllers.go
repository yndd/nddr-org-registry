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

package controllers

import (
	"github.com/yndd/nddr-org-registry/internal/controllers/deployment"
	"github.com/yndd/nddr-org-registry/internal/controllers/organization"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/yndd/nddr-org-registry/internal/shared"
)

// Setup package controllers.
func Setup(mgr ctrl.Manager, option controller.Options, nddcopts *shared.NddControllerOptions) error {
	for _, setup := range []func(ctrl.Manager, controller.Options, *shared.NddControllerOptions) error{
		organization.Setup,
		deployment.Setup,
	} {
		if err := setup(mgr, option, nddcopts); err != nil {
			return err
		}
	}

	return nil
}
