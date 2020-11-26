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
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	kcore_v1 "k8s.io/api/core/v1"
	kmeta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha1 "quortex.io/kubestitute/api/v1alpha1"
	"quortex.io/kubestitute/utils/clusterautoscaler"
)

const (
	annScaleUp   = "scaleup-policies"
	annScaleDown = "scaledown-policies"
	lblScheduler = "kubestitute.quortex.io/scheduler"
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

// matchedPolicy describe a policy with the last time it matched.
// It is used to store policy matchs state in scheduler's annotations.
type matchedPolicy struct {
	Policy corev1alpha1.SchedulerPolicy `json:"policy,omitempty"`
	Match  time.Time                    `json:"match,omitempty"`
}

// +kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=schedulers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=schedulers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=instances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch

// Reconcile reconciles the requested state with the current state.
func (r *SchedulerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("scheduler", req.NamespacedName)

	var scheduler corev1alpha1.Scheduler
	if err := r.Get(ctx, req.NamespacedName, &scheduler); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		err = client.IgnoreNotFound(err)
		if err != nil {
			log.Error(err, "unable to fetch Scheduler")
		}

		return ctrl.Result{}, err
	}

	// Check if Trigger is valid (only ClusterAutoscaler trigger supported atm).
	if scheduler.Spec.Trigger != corev1alpha1.SchedulerTriggerClusterAutoscaler {
		err := fmt.Errorf("invalid Trigger: %s", scheduler.Spec.Trigger)
		log.Error(err, "unable to fetch Scheduler", "namespace", scheduler.Namespace, "name", scheduler.Name)
		return ctrl.Result{}, err
	}

	// Fetch clusterautoscaler status configmap.
	var cm kcore_v1.ConfigMap
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: r.Conf.ClusterAutoscalerStatusNamespace,
		Name:      r.Conf.ClusterAutoscalerStatusName,
	}, &cm); err != nil {
		log.Error(
			err,
			"unable to fetch ClusterAutoscaler status configmap",
			"namespace", r.Conf.ClusterAutoscalerStatusNamespace,
			"name", r.Conf.ClusterAutoscalerStatusName,
		)

		return ctrl.Result{}, err
	}

	// Get human readable status from configmap...
	readableStatus, ok := cm.Data["status"]
	if !ok {
		err := fmt.Errorf("invalid configmap: no status")
		log.Error(
			err,
			"unable to parse ClusterAutoscaler status",
			"namespace", r.Conf.ClusterAutoscalerStatusNamespace,
			"name", r.Conf.ClusterAutoscalerStatusName,
		)
	}

	// ... and parse it.
	status := clusterautoscaler.ParseReadableString(readableStatus)
	targetStatus := clusterautoscaler.GetNodeGroupWithName(status.NodeGroups, scheduler.Spec.ASGTarget)
	if targetStatus == nil {
		err := fmt.Errorf("node group not in cluster autoscaler status: %s", scheduler.Spec.ASGTarget)
		log.Error(err, "invalid NodeGroup", "namespace", scheduler.Namespace, "name", scheduler.Name)
	}

	// Get last ScaleUp policies that matched.
	var lastScaleUpPolicies []matchedPolicy
	if ann, ok := scheduler.Annotations[annScaleUp]; ok {
		if err := json.Unmarshal([]byte(ann), &lastScaleUpPolicies); err != nil {
			log.Error(err, fmt.Sprintf("unable to unmarshal %s annotation", annScaleUp), "scheduler", scheduler.Name)
			return ctrl.Result{}, err
		}
	}

	// This will be the reference time for all reconciliation.
	now := time.Now()

	var newScaleUpPolicies []matchedPolicy
	var replicas int32
	for _, e := range scheduler.Spec.ScaleUpRules.Policies {
		// Here, we control that NodeGroup match desired Scheduler policy.
		match := matchPolicy(*targetStatus, e.SchedulerPolicy)
		if !match {
			log.Info("scaleUp policy did not match: comparison did not match")
			continue
		}

		// Try to get that policy in last matched policies...
		policy := getMatchedPolicy(lastScaleUpPolicies, e.SchedulerPolicy)
		if policy == nil {
			// ... or initialize it as matched now.
			policy = &matchedPolicy{Policy: e.SchedulerPolicy, Match: now}
		}

		// Store policy as a matched policy.
		newScaleUpPolicies = append(newScaleUpPolicies, *policy)

		// Check periodSeconds and set desired replica count accordingly.
		if now.Sub(policy.Match) >= time.Duration(e.PeriodSeconds)*time.Second {
			r := nodeGroupReplicas(*targetStatus, e.Replicas)
			// The desired replicas is the highest replicas of a matching policy.
			if replicas < r {
				replicas = r
			}
		}
	}

	// Get last ScaleDown policies that matched.
	var lastScaleDownPolicies []matchedPolicy
	if ann, ok := scheduler.Annotations[annScaleDown]; ok {
		if err := json.Unmarshal([]byte(ann), &lastScaleDownPolicies); err != nil {
			log.Error(err, fmt.Sprintf("unable to unmarshal %s annotation", annScaleDown), "scheduler", scheduler.Name)
			return ctrl.Result{}, err
		}
	}

	var newScaleDownPolicies []matchedPolicy
	var down int32
	for _, e := range scheduler.Spec.ScaleDownRules.Policies {
		// Here, we control that NodeGroup match desired Scheduler policy.
		match := matchPolicy(*targetStatus, e)
		if !match {
			log.Info("scaleDown policy did not match: comparison did not match")
			continue
		}

		// Try to get that policy in last matched policies...
		policy := getMatchedPolicy(lastScaleDownPolicies, e)
		if policy == nil {
			// ... or initialize it as matched now.
			policy = &matchedPolicy{Policy: e, Match: now}
		}

		// Store policy as a matched policy.
		newScaleDownPolicies = append(newScaleDownPolicies, *policy)

		// Check periodSeconds and set scale down count accordingly.
		if now.Sub(policy.Match) >= time.Duration(e.PeriodSeconds)*time.Second {
			// Currently, we scale down 1 by 1.
			down = 1
		}
	}

	// List instances already deployed by this scheduler
	instances := &corev1alpha1.InstanceList{}
	if err := r.List(ctx, instances, client.InNamespace(req.Namespace), client.MatchingLabels(map[string]string{lblScheduler: req.Name})); err != nil {
		log.Error(err, "unable to list Instances", "scheduler", req.Name)
		return ctrl.Result{}, err
	}

	// Compute how much instances to scale up.
	up := replicas - int32(len(instances.Items))
	if up < 0 {
		up = 0
	}

	// No need to Up/Down scale.
	if up-down == 0 {
		log.Info(fmt.Sprintf("nothing to do: %d instances to add / %d instances to remove", up, down))
		return ctrl.Result{}, nil
	}

	// Scaling up required
	if up > 0 {
		// Check StabilizationWindowSeconds
		if scheduler.Status.LastScaleUp != nil &&
			now.Sub(scheduler.Status.LastScaleUp.Time) < time.Duration(scheduler.Spec.ScaleUpRules.StabilizationWindowSeconds)*time.Second {
			// Scheduler Status for scaling up :
			// - we update all policies annotations.
			return r.endReconciliation(ctx, log, scheduler, newScaleDownPolicies, newScaleUpPolicies, scheduler.Status.LastScaleDown, scheduler.Status.LastScaleUp)
		}

		// For each new replica, we create an instance.
		log.Info(fmt.Sprintf("scheduler will scale up: %d instances to add", up), "scheduler", scheduler.Name)
		for i := 0; i < int(up); i++ {
			if err := r.Create(ctx, &corev1alpha1.Instance{
				ObjectMeta: kmeta_v1.ObjectMeta{
					GenerateName: scheduler.Name + "-",
					Namespace:    scheduler.Namespace,
					Labels:       map[string]string{lblScheduler: scheduler.Name},
				},
				Spec: corev1alpha1.InstanceSpec{
					ASG: scheduler.Spec.ASGFallback,
				},
			}); err != nil {
				log.Error(err, "unable to create Instance", "scheduler", req.Name)
				return ctrl.Result{}, err
			}
		}

		// Scheduler Status for scaling up :
		// - we remove all policies from scaleUp policies annotations.
		// - we set appropriate scaleUp time status.
		return r.endReconciliation(ctx, log, scheduler, newScaleDownPolicies, []matchedPolicy{}, scheduler.Status.LastScaleDown, &kmeta_v1.Time{Time: now})
	}

	// Scaling down required
	if down > 0 {
		// Check StabilizationWindowSeconds
		if scheduler.Status.LastScaleDown != nil &&
			now.Sub(scheduler.Status.LastScaleDown.Time) < time.Duration(scheduler.Spec.ScaleDownRules.StabilizationWindowSeconds)*time.Second {
			// Scheduler Status for scaling down :
			// - we update all policies annotations.
			return r.endReconciliation(ctx, log, scheduler, newScaleDownPolicies, newScaleUpPolicies, scheduler.Status.LastScaleDown, scheduler.Status.LastScaleUp)
		}

		// Get the older, non destroying instance
		var instance *corev1alpha1.Instance
		for _, e := range instances.Items {
			// instance being deleted, skip it
			if !e.ObjectMeta.DeletionTimestamp.IsZero() {
				continue
			}
			if instance == nil {
				instance = &e
				continue
			}
			if e.ObjectMeta.CreationTimestamp.Before(&instance.ObjectMeta.CreationTimestamp) {
				instance = &e
				continue
			}
		}

		// No instance to delete, reconciliation done !
		if instance == nil {
			log.Info("no instance to delete", "scheduler", scheduler.Name)
		} else {
			// Instance deletion
			log.Info(fmt.Sprintf("scheduler will scale down: %d instances to remove", down), "scheduler", scheduler.Name)
			if err := r.Delete(ctx, instance); err != nil {
				log.Error(err, "unable to delete Instance", "scheduler", req.Name)
				return ctrl.Result{}, err
			}
		}

		// Scheduler Status for scaling down :
		// - we remove all policies from scaleDown policies annotations.
		// - we set appropriate scaleDown time status.
		return r.endReconciliation(ctx, log, scheduler, []matchedPolicy{}, newScaleUpPolicies, &kmeta_v1.Time{Time: now}, scheduler.Status.LastScaleUp)
	}

	return ctrl.Result{}, nil
}

// endReconciliation applies the scheduler state changes at the end of the scheduler reconciliation.
// scaleDownPolicies is last matched scaleDown policies, must be empty if scaleUp happened in this reconciliation.
// scaleUpPolicies is last matched scaleUp policies, must be empty if scaleUp happened in this reconciliation.
// scaleDownTime scaleUpTime, are the dates of the last scaling events, they must be changed if one of these events took place during reconciliation.
func (r *SchedulerReconciler) endReconciliation(
	ctx context.Context,
	log logr.Logger,
	scheduler corev1alpha1.Scheduler,
	scaleDownPolicies, scaleUpPolicies []matchedPolicy,
	scaleDownTime, scaleUpTime *kmeta_v1.Time) (ctrl.Result, error) {

	// Marshal scheduler, ...
	old, err := json.Marshal(scheduler)
	if err != nil {
		log.Error(err, "failed to marshal scheduler")
		return ctrl.Result{}, err
	}

	// ... scaleDownPolicies ...
	down, err := json.Marshal(scaleDownPolicies)
	if err != nil {
		log.Error(err, "failed to marshal scale down policies")
		return ctrl.Result{}, err
	}
	// ... and scaleDownPolicies.
	up, err := json.Marshal(scaleUpPolicies)
	if err != nil {
		log.Error(err, "failed to marshal scale up policies")
		return ctrl.Result{}, err
	}

	// Then compute new scheduler to marshal it...
	scheduler.Status.LastScaleDown = scaleDownTime
	scheduler.Status.LastScaleUp = scaleUpTime
	scheduler.Annotations[annScaleDown] = string(down)
	scheduler.Annotations[annScaleUp] = string(up)
	new, err := json.Marshal(scheduler)
	if err != nil {
		log.Error(err, "failed to marshal new scheduler")
		return ctrl.Result{}, err
	}

	// ... and create a patch.
	patch, err := strategicpatch.CreateTwoWayMergePatch(old, new, scheduler)
	if err != nil {
		log.Error(err, "failed to create patch for scheduler")
		return ctrl.Result{}, err
	}

	// Apply patch to set scheduler's wanted status.
	if err = r.Patch(ctx, &scheduler, client.RawPatch(types.MergePatchType, patch)); err != nil {
		log.Error(err, "failed to patch scheduler")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// getMatchedPolicy returns the matchedPolicy matching SchedulerPolicy from a matchedPolicy slice.
func getMatchedPolicy(m []matchedPolicy, p corev1alpha1.SchedulerPolicy) *matchedPolicy {
	for _, e := range m {
		if reflect.DeepEqual(e.Policy, p) {
			return &e
		}
	}
	return nil
}

// nodeGroupIntOrFieldValue returns the desired value matching IntOrField.
// Field returns the NodeGroup Field value ans has priority over Int if a valid
// Field is given.
func nodeGroupIntOrFieldValue(ng clusterautoscaler.NodeGroup, iof corev1alpha1.IntOrField) int32 {
	if iof.FieldVal == nil {
		return iof.IntVal
	}

	switch *iof.FieldVal {
	case corev1alpha1.FieldReady:
		return ng.Health.Ready
	case corev1alpha1.FieldUnready:
		return ng.Health.Unready
	case corev1alpha1.FieldNotStarted:
		return ng.Health.NotStarted
	case corev1alpha1.FieldLongNotStarted:
		return ng.Health.LongNotStarted
	case corev1alpha1.FieldRegistered:
		return ng.Health.Registered
	case corev1alpha1.FieldLongUnregistered:
		return ng.Health.LongUnregistered
	case corev1alpha1.FieldCloudProviderTarget:
		return ng.Health.CloudProviderTarget
	}

	return iof.IntVal
}

// matchPolicy returns if given NodeGroup match desired Scheduler policy.
func matchPolicy(ng clusterautoscaler.NodeGroup, policy corev1alpha1.SchedulerPolicy) bool {
	from := nodeGroupIntOrFieldValue(ng, policy.From)
	to := nodeGroupIntOrFieldValue(ng, policy.To)

	// Perform comparison to compute Scheduler policy.
	switch policy.Operator {
	case corev1alpha1.ComparisonOperatorEqual:
		return from == to
	case corev1alpha1.ComparisonOperatorNotEqual:
		return from != to
	case corev1alpha1.ComparisonOperatorGreaterThan:
		return from > to
	case corev1alpha1.ComparisonOperatorGreaterThanOrEqual:
		return from >= to
	case corev1alpha1.ComparisonOperatorLowerThan:
		return from < to
	case corev1alpha1.ComparisonOperatorLowerThanOrEqual:
		return from <= to
	}

	return false
}

// replicas returns the number of required replicas.
func nodeGroupReplicas(ng clusterautoscaler.NodeGroup, operation corev1alpha1.IntOrArithmeticOperation) int32 {
	if operation.OperationVal == nil {
		return operation.IntVal
	}

	left := nodeGroupIntOrFieldValue(ng, operation.OperationVal.LeftOperand)
	right := nodeGroupIntOrFieldValue(ng, operation.OperationVal.RightOperand)

	// a simple func to get the biggest int32
	max := func(x, y int32) int32 {
		if x > y {
			return x
		}
		return y
	}

	// Perform arithmetic operation to compute Scheduler policy.
	switch operation.OperationVal.Operator {
	case corev1alpha1.ArithmeticOperatorPlus:
		return max(left+right, 0)
	case corev1alpha1.ArithmeticOperatorMinus:
		return max(left-right, 0)
	case corev1alpha1.ArithmeticOperatorMultiply:
		return max(left*right, 0)
	case corev1alpha1.ArithmeticOperatorDivide:
		if right != 0 {
			return max(left/right, 0)
		}
		return 0
	}

	return operation.IntVal
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
