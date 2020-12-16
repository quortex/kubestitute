/*


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
	"context"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/runtime"
	corev1alpha1 "quortex.io/kubestitute/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"quortex.io/kubestitute/utils/supervisor"
)

// SupervisionReconciler reconciles a Supervision object
type SupervisionReconciler struct {
	client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	Supervisor supervisor.Supervisor
}

// +kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=schedulers,verbs=get;list;watch
// +kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=instances,verbs=get;list;watch

// Reconcile reconciles the requested state with the current state.
func (r *SupervisionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("supervision", req.NamespacedName, "reconciliationID", uuid.New().String())

	log.V(1).Info("Supervision reconciliation started")
	defer log.V(1).Info("Supervision reconciliation done")

	// Get current existing Instances
	log.V(1).Info("Listing Instances")
	instances := &corev1alpha1.InstanceList{}
	if err := r.List(ctx, instances); err != nil {
		log.Error(err, "Failed to list Instances")
		return ctrl.Result{}, err
	}

	// Get current existing Schedulers
	log.V(1).Info("Listing Schedulers")
	schedulers := &corev1alpha1.SchedulerList{}
	if err := r.List(ctx, schedulers); err != nil {
		log.Error(err, "Failed to list Schedulers")
		return ctrl.Result{}, err
	}

	// Compute all ASG names
	asgs := []string{}
	for _, e := range instances.Items {
		asgs = append(asgs, e.Spec.ASG)
	}
	for _, e := range schedulers.Items {
		asgs = append(asgs, e.Spec.ASGTarget, e.Spec.ASGFallback)
	}

	// Give ASG names to supervisor.
	r.Supervisor.SetASGs(asgs)

	return ctrl.Result{}, nil
}

// SetupWithManager instantiates and returns the SupervisionReconciler controller.
func (r *SupervisionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Instance{}).
		Watches(&source.Kind{Type: &corev1alpha1.Scheduler{}}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
