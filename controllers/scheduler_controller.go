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
	kcore_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha1 "quortex.io/kubestitute/api/v1alpha1"
)

// SchedulerReconcilerConfiguration wraps configuration for the SchedulerReconciler.
type SchedulerReconcilerConfiguration struct {
	ClusterAutoscalerStatusNamespace string
	ClusterAutoscalerStatusName      string
}

// SchedulerReconciler reconciles a Scheduler object
type SchedulerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Conf   SchedulerReconcilerConfiguration
}

// +kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=schedulers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=schedulers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch

// Reconcile reconciles the requested state with the current state.
func (r *SchedulerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("scheduler", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

// AllSchedulersMapper is a Mapper reconciling all Schedulers.
type AllSchedulersMapper struct {
	cli client.Client
	log logr.Logger
}

// Map function reconciling all Schedulers for each event.
func (m *AllSchedulersMapper) Map(obj handler.MapObject) []reconcile.Request {
	ctx := context.Background()
	log := m.log.WithName("schedulermapper")

	// List Schedulers
	schedulers := &corev1alpha1.SchedulerList{}
	if err := m.cli.List(ctx, schedulers); err != nil {
		log.Error(err, "Unable to list Schedulers")
		return []reconcile.Request{}
	}

	// We reconcile each Scheduler atm.
	res := make([]reconcile.Request, len(schedulers.Items))
	for i, e := range schedulers.Items {
		res[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: e.Namespace,
				Name:      e.Name,
			},
		}
	}
	return res
}

// SetupWithManager instantiates and returns the SchedulerReconciler controller.
func (r *SchedulerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Scheduler{}).
		Watches(
			&source.Kind{Type: &kcore_v1.ConfigMap{}},
			&handler.EnqueueRequestsFromMapFunc{ToRequests: &AllSchedulersMapper{cli: mgr.GetClient(), log: mgr.GetLogger()}},
			r.clusterAutoscalerStatusConfigmapPredicates(),
		).
		Complete(r)
}

// reconciliationPredicates returns predicates for the controller reconciliation configuration.
func (r *SchedulerReconciler) clusterAutoscalerStatusConfigmapPredicates() builder.Predicates {
	builder.WithPredicates()
	return builder.WithPredicates(predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return r.shouldReconcileConfigmap(e.Object.(*kcore_v1.ConfigMap))
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return r.shouldReconcileConfigmap(e.Object.(*kcore_v1.ConfigMap))
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return r.shouldReconcileConfigmap(e.ObjectNew.(*kcore_v1.ConfigMap)) || r.shouldReconcileConfigmap(e.ObjectOld.(*kcore_v1.ConfigMap))
		},
	})
}

// shouldReconcileConfigmap returns if given ConfigMap is teh clusterautoscaler status
// Configmap and should be reconciled by the controller.
func (r *SchedulerReconciler) shouldReconcileConfigmap(obj *kcore_v1.ConfigMap) bool {
	// We should only consider reconciliation for clusterautoscaler status
	// configmap.
	return obj.Namespace == r.Conf.ClusterAutoscalerStatusNamespace &&
		obj.Name == r.Conf.ClusterAutoscalerStatusName
}
