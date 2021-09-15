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
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	"github.com/google/uuid"
	kcore_v1 "k8s.io/api/core/v1"
	kmeta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha1 "quortex.io/kubestitute/api/v1alpha1"
	"quortex.io/kubestitute/metrics"
	"quortex.io/kubestitute/utils/clusterautoscaler"
)

type PriorityExpanderReconcilerConfiguration struct {
	ClusterAutoscalerStatusNamespace string
	ClusterAutoscalerStatusName      string
	ClusterAutoscalerPEConfigMapName string
	PriorityExpanderNamespace        string
	PriorityExpanderName             string
}

type PriorityExpanderReconciler struct {
	client.Client
	Configuration PriorityExpanderReconcilerConfiguration
	Scheme        *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=priorityexpanders,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=priorityexpanders/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=priorityexpanders/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PriorityExpander object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *PriorityExpanderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx, "priorityexpander", req.NamespacedName, "reconciliationID", uuid.New().String())

	log.V(1).Info("PriorityExpander reconciliation started")
	defer log.V(1).Info("PriorityExpander reconciliation done")

	var pexp corev1alpha1.PriorityExpander
	if err := r.Get(ctx, req.NamespacedName, &pexp); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		err = client.IgnoreNotFound(err)
		if err != nil {
			log.Error(err, "Unable to fetch PriorityExpander")
		}

		return ctrl.Result{}, err
	}

	// Skip if PriorityExpander object is not the one and only allowed.
	if pexp.ObjectMeta.Name != r.Configuration.PriorityExpanderName ||
		pexp.ObjectMeta.Namespace != r.Configuration.PriorityExpanderNamespace {
		return ctrl.Result{}, nil
	}

	// Fetch clusterautoscaler status configmap.
	var cm kcore_v1.ConfigMap
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: r.Configuration.ClusterAutoscalerStatusNamespace,
		Name:      r.Configuration.ClusterAutoscalerStatusName,
	}, &cm); err != nil {
		log.Error(
			err,
			"Unable to fetch ClusterAutoscaler status configmap",
			"namespace", r.Configuration.ClusterAutoscalerStatusNamespace,
			"name", r.Configuration.ClusterAutoscalerStatusName,
		)

		return ctrl.Result{}, err
	}
	// Get human readable status from configmap...
	readableStatus, ok := cm.Data["status"]
	if !ok {
		err := fmt.Errorf("invalid autoscaler status configmap")
		log.Error(
			err,
			"unable to parse ClusterAutoscaler status",
			"namespace", r.Configuration.ClusterAutoscalerStatusNamespace,
			"name", r.Configuration.ClusterAutoscalerStatusName,
		)
		return ctrl.Result{}, err
	}

	// ... and parse it.
	status := clusterautoscaler.ParseReadableString(readableStatus)

	var oroot = map[string]map[string]int32{}
	for _, node := range status.NodeGroups {
		oroot[node.Name] = make(map[string]int32)
		oroot[node.Name]["CloudProviderTarget"] = node.Health.CloudProviderTarget
		oroot[node.Name]["Ready"] = node.Health.Ready
		oroot[node.Name]["Unready"] = node.Health.Unready
		oroot[node.Name]["NotStarted"] = node.Health.NotStarted
		oroot[node.Name]["LongNotStarted"] = node.Health.LongNotStarted
		oroot[node.Name]["Registered"] = node.Health.Registered
		oroot[node.Name]["LongUnregistered"] = node.Health.LongUnregistered
		oroot[node.Name]["MinSize"] = node.Health.MinSize
		oroot[node.Name]["MaxSize"] = node.Health.MaxSize
	}

	// Create new PriorityExpander template and parse it
	t, err := template.New("template").Funcs(sprig.TxtFuncMap()).Parse(pexp.Spec.Template)
	if err != nil {
		log.Error(
			err,
			"Error parsing PriorityExpander template. Check your syntax and/or rtfm.",
		)
		return r.endReconciliation(ctx, pexp, controllerutil.OperationResultNone, err)
	}

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, oroot); err != nil {
		log.Error(
			err,
			"Error parsing generating template output.",
		)
		return r.endReconciliation(ctx, pexp, controllerutil.OperationResultNone, err)
	}

	// parsed content: fmt.Println(buf.String())
	// Create the new ConfigMap object
	pecm := kcore_v1.ConfigMap{
		ObjectMeta: kmeta_v1.ObjectMeta{
			Name:      r.Configuration.ClusterAutoscalerPEConfigMapName,
			Namespace: r.Configuration.ClusterAutoscalerStatusNamespace,
		},
	}

	op, err := ctrl.CreateOrUpdate(ctx, r.Client, &pecm, func() error {

		pecm.Data = map[string]string{
			"priorities": buf.String(),
		}

		return nil
	})

	if err != nil {
		log.Error(
			err,
			"Unable to reconcile ClusterAutoscaler priority expander configmap",
			"namespace", r.Configuration.ClusterAutoscalerStatusNamespace,
			"name", r.Configuration.ClusterAutoscalerPEConfigMapName,
		)
		return r.endReconciliation(ctx, pexp, op, err)
	} else {
		log.Info("ConfigMap successfully reconciled", "operation", op)
	}

	return r.endReconciliation(ctx, pexp, op, nil)
}

func (r *PriorityExpanderReconciler) endReconciliation(
	ctx context.Context,
	pexp corev1alpha1.PriorityExpander,
	op controllerutil.OperationResult,
	error error,
) (ctrl.Result, error) {
	// update LastSuccessfulUpdate only if ConfigMap is actually modified.
	if error != nil {
		// Set to 1 if template error.
		metrics.PriorityExpanderTemplateError.Set(1)
		pexp.Status.State = corev1alpha1.PriorityExpanderStateFailure
	} else if op == controllerutil.OperationResultNone {
		pexp.Status.State = corev1alpha1.PriorityExpanderStateSuccess
		metrics.PriorityExpanderTemplateError.Set(0)
	} else {
		now := kmeta_v1.Now()
		pexp.Status.LastSuccessfulUpdate = &now
		pexp.Status.State = corev1alpha1.PriorityExpanderStateSuccess
		metrics.PriorityExpanderTemplateError.Set(0)
	}

	return ctrl.Result{}, r.Status().Update(ctx, &pexp)
}

// SetupWithManager sets up the controller with the Manager.
func (r *PriorityExpanderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.PriorityExpander{}).
		Watches(
			&source.Kind{Type: &kcore_v1.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(func(_ client.Object) []reconcile.Request {
				return []reconcile.Request{
					{
						NamespacedName: types.NamespacedName{
							Namespace: r.Configuration.PriorityExpanderNamespace,
							Name:      r.Configuration.PriorityExpanderName,
						},
					},
				}
			}),
			r.clusterAutoscalerStatusConfigmapPredicates(),
		).
		Complete(r)
}

// reconciliationPredicates returns predicates for the controller reconciliation configuration.
func (r *PriorityExpanderReconciler) clusterAutoscalerStatusConfigmapPredicates() builder.Predicates {
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

// shouldReconcileConfigmap returns if given ConfigMap is the clusterautoscaler status
// Configmap and should be reconciled by the controller.
func (r *PriorityExpanderReconciler) shouldReconcileConfigmap(obj *kcore_v1.ConfigMap) bool {
	// We should only consider reconciliation for clusterautoscaler status
	// configmap.
	return obj.Namespace == r.Configuration.ClusterAutoscalerStatusNamespace &&
		obj.Name == r.Configuration.ClusterAutoscalerStatusName
}
