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
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	kapps_v1 "k8s.io/api/apps/v1"
	kcore_v1 "k8s.io/api/core/v1"
	kpolicy_v1beta1 "k8s.io/api/policy/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kmeta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	corev1alpha1 "quortex.io/kubestitute/api/v1alpha1"
	"quortex.io/kubestitute/metrics"
	"quortex.io/kubestitute/utils/helper"
)

// InstanceReconcilerConfiguration wraps configuration for InstanceReconciler.
type InstanceReconcilerConfiguration struct {
	// The maximum number of concurrent Reconciles which can be run
	MaxConcurrentReconciles int
	// The global timeout for pods eviction
	EvictionGlobalTimeout int
}

// InstanceReconciler reconciles a Instance object
type InstanceReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	Configuration InstanceReconcilerConfiguration
	Kubernetes    *kubernetes.Clientset
	recorder      record.EventRecorder
	Autoscaling   *autoscaling.AutoScaling
	mu            sync.Mutex
}

const (
	// instanceFinalizer is a finalizer for Instances
	instanceFinalizer = "instance.finalizers.kubestitute.quortex.io"
	// the pod's field for Node name
	nodeNameField = "spec.nodeName"
	// evictionKind represents the kind of evictions object
	evictionKind = "Eviction"
	// evictionSubresource represents the kind of evictions object as pod's subresource
	evictionSubresource = "pods/eviction"
	// The delete pod polling interval
	pollInterval = time.Second
)

//+kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=instances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=instances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=instances/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;patch
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;patch
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=pods/eviction,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx, "instance", req.NamespacedName, "reconciliationID", uuid.New().String())

	log.V(1).Info("Instance reconciliation started")
	defer log.V(1).Info("Instance reconciliation done")

	var instance corev1alpha1.Instance
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		err = client.IgnoreNotFound(err)
		if err != nil {
			log.Error(err, "Unable to fetch Instance")
		}

		return ctrl.Result{}, err
	}

	// *********************************
	// **** DELETION RECONCILIATION ****
	// *********************************

	// The object is being deleted, so we perform deletion tasks before removing the finalizer.
	// Note that we can only perform deletion reconciliation if the reconciliation has been completed
	// to the end to be able to delete the instances correctly.
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() &&
		instance.Status.EC2InstanceID != "" {
		return r.reconcileDeletion(ctx, instance, log)
	}

	// ******************************************
	// **** CREATION / UPDATE RECONCILIATION ****
	// ******************************************

	// We do not allow concurrency on reconciliations related to the creation or
	// update of Instances because most of the tasks carried out during these
	// reconciliations (increment of the capacity of the ASGs, mapping of the ids
	// of ec2 instances on the CRs. ..) do not support it.
	if r.Configuration.MaxConcurrentReconciles > 1 {
		r.mu.Lock()
		defer r.mu.Unlock()
	}

	// Add finalizer if there are none.
	if !helper.ContainsString(instance.ObjectMeta.Finalizers, instanceFinalizer) {
		instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, instanceFinalizer)
		log.V(1).Info("Add finalizer to Instance", "finalizer", instanceFinalizer)
		return ctrl.Result{}, r.Update(ctx, &instance)
	}

	// 1st STEP
	//
	// Initialize Instance with TriggerScaling state.
	if instance.Status.State == corev1alpha1.InstanceStateNone {
		// Next step, we will trigger a new EC2 instance.
		instance.Status.State = corev1alpha1.InstanceStateTriggerScaling
		log.V(1).Info("Updating Instance Status", "state", instance.Status.State)
		return ctrl.Result{}, r.Status().Update(ctx, &instance)
	}

	// Retrieve AWS autoscaling group.
	asgName := instance.Spec.ASG
	res, err := r.Autoscaling.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{&asgName},
	})
	if err == nil && len(res.AutoScalingGroups) == 0 {
		err = fmt.Errorf("autoscaling group not found")
	}
	if err != nil {
		log.Error(err, "Failed to get autoscaling group", "name", asgName)
		return ctrl.Result{}, err
	}
	asg := res.AutoScalingGroups[0]

	// 2nd STEP
	//
	// Trigger a new Instance in the ASG
	if instance.Status.State == corev1alpha1.InstanceStateTriggerScaling {
		capacity := *asg.DesiredCapacity + 1
		log.Info("Incrementing autoscaling group desired capacity", "name", asgName, "capacity", capacity)
		_, err := r.Autoscaling.SetDesiredCapacity(&autoscaling.SetDesiredCapacityInput{
			AutoScalingGroupName: asg.AutoScalingGroupName,
			DesiredCapacity:      &capacity,
			HonorCooldown:        &instance.Spec.HonorCooldown,
		})
		if err != nil {
			log.Error(err, "Failed to set autoscaling group desired capacity", "name", asg.AutoScalingGroupName, "capacity", capacity)
			return ctrl.Result{}, err
		}

		// Next step, we will wait for EC2 instance joining the ASG.
		instance.Status.State = corev1alpha1.InstanceStateWaitInstance
		log.V(1).Info("Updating Instance", "state", instance.Status.State)
		return ctrl.Result{}, r.Status().Update(ctx, &instance)
	}

	// 3rd STEP
	//
	// Try to get a new joining instance in the ASG to associate it to our Instance.
	if instance.Status.State == corev1alpha1.InstanceStateWaitInstance {
		log.Info("Waiting for ec2 instance")
		ec2Instances := asg.Instances
		// List all instances
		instances := &corev1alpha1.InstanceList{}
		if err := r.List(ctx, instances); err != nil {
			log.Error(err, "Failed to list instances")
			return ctrl.Result{}, err
		}

		// We select an EC2 instance to attach to our Instance resource.
		// Seems that instances in ASG are sorted from newer to older,
		// anyway that's not a big deal to associate an instance that has
		// not been scheduled, by this Instance resource.
		var instanceID string
		for i := len(ec2Instances) - 1; i >= 0; i-- {
			e := ec2Instances[i]

			// Select only Pending or InService Instance
			if !helper.ContainsString([]string{
				autoscaling.LifecycleStatePending,
				autoscaling.LifecycleStatePendingWait,
				autoscaling.LifecycleStatePendingProceed,
				autoscaling.LifecycleStateInService,
			}, aws.StringValue(e.LifecycleState)) {
				continue
			}

			alreadyUsed := false
			for _, inst := range instances.Items {
				// Exclude itself
				if inst.Name == instance.Name {
					continue
				}
				// Exclude Instances with empty IDs
				if inst.Status.EC2InstanceID == "" {
					continue
				}
				// Already attached EC2 instance
				if aws.StringValue(e.InstanceId) == inst.Status.EC2InstanceID {
					alreadyUsed = true
				}
			}

			if !alreadyUsed {
				instanceID = aws.StringValue(e.InstanceId)
				break
			}
		}

		// Seems that new instance has not joined yet the ASG, we'll try it later.
		if instanceID == "" {
			log.Info("Instance has not joined yet the ASG, retry in 5 sec")
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}

		// Next step, we will wait for Node to join the cluster.
		instance.Status.State = corev1alpha1.InstanceStateWaitNode
		instance.Status.EC2InstanceID = instanceID
		log.Info("EC2 instance retrieved", "ec2InstanceID", instance.Status.EC2InstanceID)
		log.V(1).Info("Updating Instance", "state", instance.Status.State, "ec2InstanceID", instance.Status.EC2InstanceID)
		return ctrl.Result{}, r.Status().Update(ctx, &instance)
	}

	// 4th STEP
	//
	// Try to get a new joining Node to associate it to our Instance.
	if instance.Status.State == corev1alpha1.InstanceStateWaitNode {
		log.Info("Waiting for node")
		// List nodes
		nodes := &kcore_v1.NodeList{}
		if err := r.List(ctx, nodes); err != nil {
			log.Error(err, "Unable to list Nodes")
			return ctrl.Result{}, err
		}

		// We try to get a node matching the Instance ID to attach it to the Instance Status
		for _, e := range nodes.Items {
			// Node is not managed by AWS, the reconciler does not support it.
			if !isAWSNode(e) {
				continue
			}
			instanceID := ec2InstanceID(e)
			if instanceID == "" {
				continue
			}

			// Final reconciliation step, the node matching instanceID has been identified.
			if instance.Status.EC2InstanceID == instanceID {
				log.Info("Node retrieved", "node", instance.Status.Node)

				// Increment scaled_up_nodes_total metric.
				metrics.ScaledUpNodesTotal.WithLabelValues(asgName, instance.Labels[lblScheduler]).Add(float64(1))

				// Instance ready, reconciliation done.
				instance.Status.State = corev1alpha1.InstanceStateReady
				instance.Status.Node = e.GetName()
				log.V(1).Info("Updating Instance", "state", instance.Status.State, "node", instance.Status.Node)
				return ctrl.Result{}, r.Status().Update(ctx, &instance)
			}
		}
	}

	return ctrl.Result{}, nil
}

// reconcileDeletion handle Instance deletion tasks
func (r *InstanceReconciler) reconcileDeletion(
	ctx context.Context,
	instance corev1alpha1.Instance,
	log logr.Logger,
) (ctrl.Result, error) {
	// 1st STEP
	//
	// Initialize Instance with the desired state.
	// If Instance is awaiting node, we detach the instance directly.
	// If Node already joined the cluster we drain it.
	if instance.Status.State == corev1alpha1.InstanceStateReady {
		// Next step, we will drain the associated Node.
		instance.Status.State = corev1alpha1.InstanceStateDrainNode
		log.V(1).Info("Updating Instance", "state", instance.Status.State)
		return ctrl.Result{}, r.Status().Update(ctx, &instance)
	} else if instance.Status.State == corev1alpha1.InstanceStateWaitNode {
		// Next step, we will detach the EC2 instance from the ASG.
		instance.Status.State = corev1alpha1.InstanceStateTerminateInstance
		log.V(1).Info("Updating Instance", "state", instance.Status.State)
		return ctrl.Result{}, r.Status().Update(ctx, &instance)
	}

	// 2nd STEP
	//
	// The instance has already scheduled a kubernetes Node.
	// We need to properly drain this Node before deleting instance.
	if instance.Status.State == corev1alpha1.InstanceStateDrainNode {
		// Get the Instance resource associated Node
		nodeName := instance.Status.Node
		log.Info("Draining node", "node", nodeName)
		var node kcore_v1.Node
		err := r.Get(ctx, types.NamespacedName{Name: nodeName}, &node)
		// In the case of a not found error, it seems that the node no longer exists.
		// We can therefore consider this phase of reconciliation to be accomplished.
		if client.IgnoreNotFound(err) != nil {
			log.Info("Node no longer exist", "node", nodeName)
			return ctrl.Result{}, err
		} else if err == nil {

			// First, we cordon the Node (set it as unschedulable)
			log.Info("Cordon node", "node", nodeName)
			err := r.cordonNode(ctx, &node)
			if err != nil {
				log.Error(err, "Unable to cordon Node", "node", nodeName)
				return ctrl.Result{}, err
			}

			// Get pods scheduled on that Node
			pods := &kcore_v1.PodList{}
			if err := r.List(ctx, pods, client.MatchingFieldsSelector{
				Selector: fields.SelectorFromSet(fields.Set{nodeNameField: nodeName}),
			}); err != nil {
				log.Error(err, "Failed to list node's pods", "node", nodeName)
				return ctrl.Result{}, err
			}

			// Evict pods
			// We don't care about errors here.
			// Either we can't process them or the eviction has timeout.
			log.Info("Evicting pods", "node", nodeName)
			if err := r.evictPods(
				ctx,
				log,
				instance,
				nodeName,
				instance.Labels[lblScheduler],
				filterPods(pods.Items, r.deletedFilter, r.daemonSetFilter),
			); err != nil {
				log.Error(err, "Failed to evict pods", "node", nodeName)
				return ctrl.Result{}, err
			}
		}

		// Next step, we will detach the EC2 instance from the ASG.
		instance.Status.State = corev1alpha1.InstanceStateTerminateInstance
		log.V(1).Info("Updating Instance", "state", instance.Status.State)
		return ctrl.Result{}, r.Status().Update(ctx, &instance)
	}

	// 3rd STEP
	//
	// We detach the EC2 instance from the Autoscaling Group.
	if instance.Status.State == corev1alpha1.InstanceStateTerminateInstance {
		// Retrieve AWS autoscaling group.
		asgName := instance.Spec.ASG
		res, err := r.Autoscaling.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{&asgName},
		})
		if err == nil && len(res.AutoScalingGroups) == 0 {
			err = fmt.Errorf("autoscaling group not found")
		}
		if err != nil {
			log.Error(err, "Failed to get autoscaling group", "name", asgName)
			return ctrl.Result{}, err
		}
		asg := res.AutoScalingGroups[0]

		// Then, if instance is part of Autoscaling Group, we terminate it.
		if i := instanceWithID(asg.Instances, instance.Status.EC2InstanceID); i != nil {
			group := asgName
			instanceID := instance.Status.EC2InstanceID
			if _, err := r.Autoscaling.TerminateInstanceInAutoScalingGroup(&autoscaling.TerminateInstanceInAutoScalingGroupInput{
				InstanceId:                     aws.String(instanceID),
				ShouldDecrementDesiredCapacity: aws.Bool(true),
			}); err != nil {
				log.Error(err, "Failed to terminate instance", "group", group, "instance", instanceID)
				return ctrl.Result{}, err
			}
		}

		// Increment scaled_down_nodes_total metric.
		metrics.ScaledDownNodesTotal.WithLabelValues(asgName, instance.Labels[lblScheduler]).Add(float64(1))
	}

	// remove our finalizer from the list and update it.
	log.V(1).Info("Remove finalizer from Instance", "finalizer", instanceFinalizer)
	instance.ObjectMeta.Finalizers = helper.RemoveString(instance.ObjectMeta.Finalizers, instanceFinalizer)
	return ctrl.Result{}, r.Update(ctx, &instance)
}

// isAWSNode returns if given Node is identified as an AWS cluster Node
func isAWSNode(node kcore_v1.Node) bool {
	return strings.HasPrefix(node.Spec.ProviderID, "aws://")
}

// ec2InstanceID returns the EC2 instance ID from a given Node
func ec2InstanceID(node kcore_v1.Node) string {
	return path.Base(node.Spec.ProviderID)
}

// instanceWithID returns the instance with the given ID or nil
func instanceWithID(slice []*autoscaling.Instance, id string) *autoscaling.Instance {
	for _, e := range slice {
		if aws.StringValue(e.InstanceId) == id {
			return e
		}
	}
	return nil
}

// cordonNode cordon the given Node (mark it as unschedulable).
func (r *InstanceReconciler) cordonNode(
	ctx context.Context,
	node *kcore_v1.Node,
) error {
	// To cordon a Node, patch it to set it Unschedulable.
	old, err := json.Marshal(node)
	if err != nil {
		return err
	}
	node.Spec.Unschedulable = true
	new, err := json.Marshal(node)
	if err != nil {
		return err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(old, new, node)
	if err != nil {
		return err
	}
	if err = r.Patch(ctx, node, client.RawPatch(types.StrategicMergePatchType, patch)); err != nil {
		return err
	}

	return nil
}

// evictPods evict given pods and returns when all pods have been
// successfully deleted, error occurred or evictionGlobalTimeout expired.
// This code is largely inspired by kubectl cli source code.
func (r *InstanceReconciler) evictPods(ctx context.Context, log logr.Logger, instance corev1alpha1.Instance, nodeName, scheduler string, pods []kcore_v1.Pod) error {
	returnCh := make(chan error, 1)
	policyGroupVersion, err := CheckEvictionSupport(r.Kubernetes)
	if err != nil {
		return err
	}

	evictionGlobalTimeout := time.Duration(r.Configuration.EvictionGlobalTimeout) * time.Second
	ctx, cancel := context.WithTimeout(ctx, evictionGlobalTimeout)
	defer cancel()

	// A reusable function to record pod evictions error events.
	recordEvictionErrorEvent := func(recorder record.EventRecorder, instance *corev1alpha1.Instance, pod kcore_v1.Pod, err error) {
		recorder.Eventf(instance, kcore_v1.EventTypeWarning, "FailedEviction", "Pod eviction failed %s/%s err: %s", pod.Namespace, pod.Name, err.Error())
	}

	for _, pod := range pods {
		go func(pod kcore_v1.Pod, returnCh chan error) {
			for {
				log.Info("Evicting pod", "name", pod.Name, "namespace", pod.Namespace)
				select {
				case <-ctx.Done():
					// return here or we'll leak a goroutine.
					returnCh <- fmt.Errorf("error when evicting pods/%q -n %q: global timeout reached: %v", pod.Name, pod.Namespace, evictionGlobalTimeout)
					return
				default:
				}

				// Create a temporary pod so we don't mutate the pod in the loop.
				activePod := pod
				err := r.evictPod(ctx, activePod, policyGroupVersion)
				if err == nil {
					metrics.EvictedPodsTotal.WithLabelValues(instance.Spec.ASG, nodeName, scheduler).Add(float64(1))
					r.recorder.Eventf(&instance, kcore_v1.EventTypeNormal, "SuccessfulEviction", "Pod evicted %s/%s", pod.Namespace, pod.Name)
					break
				} else if apierrors.IsNotFound(err) {
					returnCh <- nil
					return
				} else if apierrors.IsTooManyRequests(err) {
					log.Error(err, "Failed to evict pod (will retry after 5s)", "name", pod.Name, "namespace", pod.Namespace)
					recordEvictionErrorEvent(r.recorder, &instance, pod, err)
					time.Sleep(5 * time.Second)
				} else if !activePod.ObjectMeta.DeletionTimestamp.IsZero() && apierrors.IsForbidden(err) && apierrors.HasStatusCause(err, kcore_v1.NamespaceTerminatingCause) {
					// an eviction request in a deleting namespace will throw a forbidden error,
					// if the pod is already marked deleted, we can ignore this error, an eviction
					// request will never succeed, but we will waitForDelete for this pod.
					break
				} else if apierrors.IsForbidden(err) && apierrors.HasStatusCause(err, kcore_v1.NamespaceTerminatingCause) {
					// an eviction request in a deleting namespace will throw a forbidden error,
					// if the pod is not marked deleted, we retry until it is.
					log.Error(err, "Failed to evict pod (will retry after 5s)", "name", pod.Name, "namespace", pod.Namespace)
					recordEvictionErrorEvent(r.recorder, &instance, pod, err)
					time.Sleep(5 * time.Second)
				} else {
					returnCh <- fmt.Errorf("error when evicting pods/%q -n %q: %v", activePod.Name, activePod.Namespace, err)
					recordEvictionErrorEvent(r.recorder, &instance, pod, err)
					return
				}
			}
			_, err := r.waitForDelete(ctx, []kcore_v1.Pod{pod})
			if err == nil {
				returnCh <- nil
			} else {
				returnCh <- fmt.Errorf("error when waiting for pod %q terminating: %v", pod.Name, err)
			}
		}(pod, returnCh)
	}

	doneCount := 0
	var errors []error

	numPods := len(pods)
	for doneCount < numPods {
		//nolint:gosimple
		select {
		case err := <-returnCh:
			doneCount++
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	return utilerrors.NewAggregate(errors)
}

// evictPod will evict the given pod, or return an error if it couldn't
// This code is largely inspired by kubectl cli source code.
func (r *InstanceReconciler) evictPod(ctx context.Context, pod kcore_v1.Pod, policyGroupVersion string) error {
	gracePeriod := int64(time.Second * 30)
	if pod.Spec.TerminationGracePeriodSeconds != nil && *pod.Spec.TerminationGracePeriodSeconds < gracePeriod {
		gracePeriod = *pod.Spec.TerminationGracePeriodSeconds
	}

	eviction := &kpolicy_v1beta1.Eviction{
		TypeMeta: kmeta_v1.TypeMeta{
			APIVersion: policyGroupVersion,
			Kind:       evictionKind,
		},
		ObjectMeta: kmeta_v1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
		DeleteOptions: &kmeta_v1.DeleteOptions{GracePeriodSeconds: &gracePeriod},
	}

	// Remember to change change the URL manipulation func when Eviction's version change
	return r.Kubernetes.PolicyV1beta1().Evictions(eviction.Namespace).Evict(ctx, eviction)
}

// CheckEvictionSupport uses Discovery API to find out if the server support
// eviction subresource If support, it will return its groupVersion; Otherwise,
// it will return an empty string.
// This code is largely inspired by kubectl cli source code.
func CheckEvictionSupport(clientset kubernetes.Interface) (string, error) {
	discoveryClient := clientset.Discovery()
	groupList, err := discoveryClient.ServerGroups()
	if err != nil {
		return "", err
	}
	foundPolicyGroup := false
	var policyGroupVersion string
	for _, group := range groupList.Groups {
		if group.Name == "policy" {
			foundPolicyGroup = true
			policyGroupVersion = group.PreferredVersion.GroupVersion
			break
		}
	}
	if !foundPolicyGroup {
		return "", nil
	}
	resourceList, err := discoveryClient.ServerResourcesForGroupVersion("v1")
	if err != nil {
		return "", err
	}
	for _, resource := range resourceList.APIResources {
		if resource.Name == evictionSubresource && resource.Kind == evictionKind {
			return policyGroupVersion, nil
		}
	}
	return "", nil
}

// deleteTimeout compute the delete timeout from given pods.
func deleteTimeout(pods []kcore_v1.Pod) time.Duration {
	// We return the max DeletionGracePeriodSeconds from pods with
	// a 30sec overhead.
	maxGrace := int64(30)
	for _, e := range pods {
		if grace := e.DeletionGracePeriodSeconds; grace != nil {
			if *grace > maxGrace {
				maxGrace = *grace
			}
		}
	}

	return time.Duration(maxGrace+30) * time.Second
}

// waitForDelete poll pods to check their deletion.
// This code is largely inspired by kubectl cli source code.
func (r *InstanceReconciler) waitForDelete(ctx context.Context, pods []kcore_v1.Pod) ([]kcore_v1.Pod, error) {
	err := wait.PollImmediate(pollInterval, deleteTimeout(pods), func() (bool, error) {
		pendingPods := []kcore_v1.Pod{}
		for i, pod := range pods {
			p := &kcore_v1.Pod{}
			err := r.Get(ctx, types.NamespacedName{
				Namespace: pod.Namespace,
				Name:      pod.Name,
			}, p)
			if apierrors.IsNotFound(err) || (p != nil && p.ObjectMeta.UID != pod.ObjectMeta.UID) {
				continue
			} else if err != nil {
				return false, err
			} else {
				pendingPods = append(pendingPods, pods[i])
			}
		}
		pods = pendingPods
		if len(pendingPods) > 0 {
			select {
			case <-ctx.Done():
				return false, fmt.Errorf("Eviction global timeout reached")
			default:
				return false, nil
			}
		}
		return true, nil
	})
	return pods, err
}

// PodFilter describes functions to filter pods from slice
type PodFilter func([]kcore_v1.Pod) []kcore_v1.Pod

// filterPods filter a pod slice with given filters
func filterPods(pods []kcore_v1.Pod, filters ...PodFilter) []kcore_v1.Pod {
	for _, f := range filters {
		pods = f(pods)
	}
	return pods
}

// daemonSetFilter filter dameonsets pods
//
//nolint:unused
func (r *InstanceReconciler) daemonSetFilter(pods []kcore_v1.Pod) []kcore_v1.Pod {
	// Note that we return false in cases where the pod is DaemonSet managed,
	// regardless of flags.
	//
	// The exception is for pods that are orphaned (the referencing
	// management resource - including DaemonSet - is not found).
	for i := len(pods) - 1; i >= 0; i-- {
		pod := pods[i]
		controllerRef := kmeta_v1.GetControllerOf(&pod)
		if controllerRef == nil || controllerRef.Kind != kapps_v1.SchemeGroupVersion.WithKind("DaemonSet").Kind {
			continue
		}

		// Any finished pod can be removed.
		if pod.Status.Phase == kcore_v1.PodSucceeded || pod.Status.Phase == kcore_v1.PodFailed {
			continue
		}

		if _, err := r.Kubernetes.AppsV1().DaemonSets(pod.Namespace).Get(context.TODO(), controllerRef.Name, kmeta_v1.GetOptions{}); err != nil {
			// remove orphaned pods
			if apierrors.IsNotFound(err) {
				pods = append(pods[:i], pods[i+1:]...)
				continue
			}

			continue
		}

		pods = append(pods[:i], pods[i+1:]...)
	}

	return pods
}

// deletedFilter filter already deleted pods
func (r *InstanceReconciler) deletedFilter(pods []kcore_v1.Pod) []kcore_v1.Pod {
	for i := len(pods) - 1; i >= 0; i-- {
		pod := pods[i]
		if !pod.ObjectMeta.DeletionTimestamp.IsZero() {
			pods = append(pods[:i], pods[i+1:]...)
		}
	}
	return pods
}

// SetupWithManager sets up the controller with the Manager.
func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Index nodeName pod's spec field to get pods scheduled on Node
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&kcore_v1.Pod{},
		nodeNameField,
		func(o client.Object) []string {
			pod := o.(*kcore_v1.Pod)
			return []string{pod.Spec.NodeName}
		},
	); err != nil {
		return err
	}

	// Get event recorder
	r.recorder = mgr.GetEventRecorderFor("Instance")

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Instance{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.Configuration.MaxConcurrentReconciles,
		}).
		Watches(
			&source.Kind{Type: &kcore_v1.Node{}},
			handler.EnqueueRequestsFromMapFunc(func(o client.Object) []reconcile.Request {
				ctx := context.Background()
				log := ctrllog.Log.WithName("instancemapper")

				// Obtain the modified node
				node, ok := o.(*kcore_v1.Node)
				if !ok {
					log.Error(fmt.Errorf("fail to cast object %s into Node", o.GetName()),
						"Fail to cast object into Node", "kind", o.GetObjectKind().GroupVersionKind, "obj", o)
					return []reconcile.Request{}
				}

				// Node is not managed by AWS, the reconciler does not support it.
				if !isAWSNode(*node) {
					log.Info("Node is not managed by AWS, the reconciler does not support it", "node", node.Name)
					return []reconcile.Request{}
				}

				// Get instance ID from node's Provider field
				// E.g for an AWS EKS / KOPS node
				// providerID: aws:///eu-west-1b/i-0d0b3844fcfb3c137
				ec2InstanceID := ec2InstanceID(*node)
				if ec2InstanceID == "" {
					log.Error(fmt.Errorf("Node has no valid ProviderID"), "name", node.Name)
					return []reconcile.Request{}
				}

				// List instances
				instances := &corev1alpha1.InstanceList{}
				if err := r.List(ctx, instances); err != nil {
					log.Error(err, "Unable to list Instances")
					return []reconcile.Request{}
				}

				// If an Instance match the ec2InstanceID, we reconcile it.
				for _, e := range instances.Items {
					if e.Status.EC2InstanceID == ec2InstanceID {
						return []reconcile.Request{
							{
								NamespacedName: types.NamespacedName{
									Namespace: e.Namespace,
									Name:      e.Name,
								},
							},
						}
					}
				}

				// No matching Instance, no reconciliation required.
				return []reconcile.Request{}
			}),
		).
		Complete(r)
}
