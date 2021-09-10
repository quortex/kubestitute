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

	"github.com/google/uuid"
	kcore_v1 "k8s.io/api/core/v1"
	kmeta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "quortex.io/kubestitute/api/v1alpha1"
	"quortex.io/kubestitute/utils/clusterautoscaler"
)

type PriorityExpanderReconcilerConfiguration struct {
	ClusterAutoscalerStatusNamespace string
	ClusterAutoscalerStatusName      string
	ClusterAutoscalerPEConfigMapName string
}

type PriorityExpanderReconciler struct {
	client.Client
	Configuration PriorityExpanderReconcilerConfiguration
	Scheme        *runtime.Scheme
}

type PriorityExpanderConfigMap struct {
	ready, unready, notStarted, longNotStarted, registered, longUnregistered, cloudProviderTarget, minSize, maxSize int32
}

//+kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=priorityexpanders,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=priorityexpanders/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=priorityexpanders/finalizers,verbs=update

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
	log := ctrllog.FromContext(ctx, "pexp", req.NamespacedName, "reconciliationID", uuid.New().String())

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
		err := fmt.Errorf("Invalid configmap: no status")
		log.Error(
			err,
			"unable to parse ClusterAutoscaler status",
			"namespace", r.Configuration.ClusterAutoscalerStatusNamespace,
			"name", r.Configuration.ClusterAutoscalerStatusName,
		)
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
	t := template.Must(template.New("template").Parse(pexp.Spec.Template))
	buf := new(bytes.Buffer)
	_ = t.Execute(buf, status)

	// Create the new ConfigMap object
	pecm := kcore_v1.ConfigMap{
		TypeMeta: kmeta_v1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: kmeta_v1.ObjectMeta{
			Name:      r.Configuration.ClusterAutoscalerPEConfigMapName,
			Namespace: r.Configuration.ClusterAutoscalerStatusNamespace,
		},
		Data: map[string]string{
			"priorities": buf.String(),
		},
	}

	// Update it inplace. The ConfigMap _must_ exist.
	if err := r.Update(ctx, &pecm); err != nil {
		log.Error(
			err,
			"Unable to update ClusterAutoscaler priority expander configmap",
			"namespace", r.Configuration.ClusterAutoscalerStatusNamespace,
			"name", r.Configuration.ClusterAutoscalerStatusName,
		)

		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PriorityExpanderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.PriorityExpander{}).
		Complete(r)
}
