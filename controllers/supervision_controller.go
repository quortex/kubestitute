/*
Copyright 2021.

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

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha1 "quortex.io/kubestitute/api/v1alpha1"
	"quortex.io/kubestitute/utils/supervisor"
)

// SupervisionReconciler reconciles a Supervision object
type SupervisionReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Supervisor supervisor.Supervisor
}

//+kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=schedulers,verbs=get;list;watch
//+kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=instances,verbs=get;list;watch

func (r *SupervisionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx, "supervision", req.NamespacedName, "reconciliationID", uuid.New().String())

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
	for _, instance := range instances.Items {
		asgs = append(asgs, instance.Spec.ASG)
	}
	for _, scheduler := range schedulers.Items {
		asgs = append(asgs, scheduler.Spec.ASGTarget, scheduler.Spec.ASGFallback)
	}

	// Give ASG names to supervisor.
	r.Supervisor.SetASGs(asgs)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SupervisionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Instance{}).
		Watches(&source.Kind{Type: &corev1alpha1.Scheduler{}}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
