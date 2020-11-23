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
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/go-logr/logr"
	kcore_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1alpha1 "quortex.io/kubestitute/api/v1alpha1"
	ec2adapter "quortex.io/kubestitute/client/ec2adapter/client"
	"quortex.io/kubestitute/client/ec2adapter/client/operations"
	"quortex.io/kubestitute/client/ec2adapter/models"
)

// InstanceReconciler reconciles a Instance object
type InstanceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// instanceFinalizer is a finalizer for Instances
const instanceFinalizer = "instance.finalizers.kubestitute.quortex.io"

// +kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=instances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=instances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;update;patch

// Reconcile reconciles the Instance requested state with the current state.
func (r *InstanceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("instance", req.NamespacedName)

	var instance corev1alpha1.Instance
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		err = client.IgnoreNotFound(err)
		if err != nil {
			log.Error(err, "unable to fetch Instance")
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

	// **********************************
	// **** LIFECYCLE RECONCILIATION ****
	// **********************************

	// 1st STEP
	//
	// Initialize Instance with TriggerScaling state and add the finalizer.
	if instance.Status.State == corev1alpha1.InstanceStateNone {
		instance.Status.State = corev1alpha1.InstanceStateTriggerScaling
		instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, instanceFinalizer)
		log.Info(fmt.Sprintf("setting status: %s, finalizer: %s", instance.Status.State, instanceFinalizer))
		return ctrl.Result{}, r.Update(ctx, &instance)
	}

	// Instantiate aws-ec2-adapter client requirements
	ec2adapterCli := ec2adapter.NewHTTPClientWithConfig(nil, &ec2adapter.TransportConfig{
		Host:    "localhost:8008",
		Schemes: []string{"http"},
	})

	// Check readiness of the ec2 adapter service and
	// requeue until availability.
	_, err := ec2adapterCli.Operations.Ping(&operations.PingParams{Context: ctx})
	if err != nil {
		log.Info("ec2 adapter not available, retrying in 5 seconds.", "err", err)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Retrieve AWS autoscaling group.
	asgName := instance.Spec.ASG
	res, err := ec2adapterCli.Operations.GetAutoscalingGroup(&operations.GetAutoscalingGroupParams{
		Context: ctx,
		Name:    asgName,
	})
	if err != nil {
		log.Error(err, "failed to get autoscaling group", "name", asgName)
		return ctrl.Result{}, err
	}
	asg := res.Payload

	// 2nd STEP
	//
	// Trigger a new Instance in the ASG
	if instance.Status.State == corev1alpha1.InstanceStateTriggerScaling {
		capacity := asg.DesiredCapacity + 1
		_, err := ec2adapterCli.Operations.SetAutoscalingGroupDesiredCapacity(&operations.SetAutoscalingGroupDesiredCapacityParams{
			Context: ctx,
			Name:    asg.Name,
			Request: &models.SetDesiredCapacityRequest{
				DesiredCapacity: capacity,
				HonorCooldown:   instance.Spec.HonorCooldown,
			},
		})
		if err != nil {
			log.Error(err, "failed to set autoscaling group desired capacity", "name", asg.Name, "capacity", capacity)
			return ctrl.Result{}, err
		}
		instance.Status.State = corev1alpha1.InstanceStateWaitInstance
		log.Info(fmt.Sprintf("setting status: %s", instance.Status.State))
		return ctrl.Result{}, r.Update(ctx, &instance)
	}

	// 3rd STEP
	//
	// Try to get a new joining instance in the ASG to associate it to our Instance.
	if instance.Status.State == corev1alpha1.InstanceStateWaitInstance {
		ec2Instances := asg.Instances
		// List all instances
		instances := &corev1alpha1.InstanceList{}
		if err := r.List(ctx, instances); err != nil {
			log.Error(err, "failed to list instances")
			return ctrl.Result{}, err
		}

		// We select an EC2 instance to attach to our Instance resource.
		// Seems that instances in ASG are sorted from newer to older,
		// anyway that's not a big deal to associate an instance that has
		// not been scheduled, by this Instance resource.
		var instanceID string
		for i := len(ec2Instances) - 1; i >= 0; i-- {
			e := ec2Instances[i]
			flag := false
			for _, inst := range instances.Items {
				// Exclude itself
				if inst.Name == instance.Name {
					continue
				}
				// Exclude Instances with empty IDs
				if inst.Status.EC2InstanceID == "" {
					continue
				}
				// Already attched EC2 instance
				if e.InstanceID == inst.Status.EC2InstanceID {
					flag = true
					break
				}
			}
			if !flag {
				instanceID = e.InstanceID
			}
		}

		// Seems that new instance has not joined yet the ASG, we'll try it later.
		if instanceID == "" {
			log.Info("instance has not joined yet the ASG, retry in 5 sec")
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}

		// Upgrade Instance with status and instanceID.
		instance.Status.State = corev1alpha1.InstanceStateWaitNode
		instance.Status.EC2InstanceID = instanceID
		log.Info(fmt.Sprintf("setting status: %s, ec2InstanceID: %s", instance.Status.State, instance.Status.EC2InstanceID))
		return ctrl.Result{}, r.Update(ctx, &instance)
	}

	// 4th STEP
	//
	// Try to get a new joining Node to associate it to our Instance.
	if instance.Status.State == corev1alpha1.InstanceStateWaitNode {
		// List nodes
		nodes := &kcore_v1.NodeList{}
		if err := r.List(ctx, nodes); err != nil {
			log.Error(err, "unable to list Nodes")
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
				instance.Status.State = corev1alpha1.InstanceStateReady
				instance.Status.Node = e.GetName()
				log.Info(fmt.Sprintf("setting status: %s, node: %s", instance.Status.State, instance.Status.Node))
				return ctrl.Result{}, r.Update(ctx, &instance)
			}
		}
	}

	return ctrl.Result{}, nil
}

// reconcileDeletion handle Instance deletion tasks
func (r *InstanceReconciler) reconcileDeletion(
	ctx context.Context,
	instance corev1alpha1.Instance,
	log logr.Logger) (ctrl.Result, error) {

	// 1st STEP
	//
	// The instance has already scheduled a kubernetes Node.
	// We need to properly drain this Node before deleting instance.
	// if instance.Status.State == corev1alpha1.InstanceStateReady {
	// 	// TODO: node drain implementation
	// }

	// Instantiate aws-ec2-adapter client requirements
	ec2adapterCli := ec2adapter.NewHTTPClientWithConfig(nil, &ec2adapter.TransportConfig{
		Host:    "localhost:8008",
		Schemes: []string{"http"},
	})

	// Check readiness of the ec2 adapter service and
	// requeue until availability.
	_, err := ec2adapterCli.Operations.Ping(&operations.PingParams{Context: ctx})
	if err != nil {
		log.Info("ec2 adapter not available, retrying in 5 seconds.", "err", err)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// 2nd STEP
	//
	// We detach the EC2 instance from the Autoscaling Group.
	if instance.Status.State == corev1alpha1.InstanceStateDetachInstance ||
		instance.Status.State == corev1alpha1.InstanceStateWaitNode {
		// Retrieve AWS autoscaling group.
		asgName := instance.Spec.ASG
		res, err := ec2adapterCli.Operations.GetAutoscalingGroup(&operations.GetAutoscalingGroupParams{
			Context: ctx,
			Name:    asgName,
		})
		if err != nil {
			log.Error(err, "failed to get autoscaling group", "name", asgName)
			return ctrl.Result{}, err
		}
		asg := res.Payload

		// Then, if instance is part of Autoscaling Group, we terminate it.
		if i := instanceWithID(asg.Instances, instance.Status.EC2InstanceID); i != nil {
			group := asgName
			instanceID := instance.Status.EC2InstanceID
			if _, err := ec2adapterCli.Operations.DetachAutoscalingGroupInstances(&operations.DetachAutoscalingGroupInstancesParams{
				Context: ctx,
				Name:    group,
				Request: &models.DetachInstancesRequest{
					InstanceIds:                    []string{instanceID},
					ShouldDecrementDesiredCapacity: true,
				},
			}); err != nil {
				log.Error(err, "failed to detach instance", "group", group, "instance", instanceID)
				return ctrl.Result{}, err
			}
		}

		instance.Status.State = corev1alpha1.InstanceStateTerminateInstance
		log.Info(fmt.Sprintf("setting status: %s", instance.Status.State))
		return ctrl.Result{}, r.Update(ctx, &instance)
	}

	// 3rd STEP
	//
	// We terminate the EC2 instance.
	if instance.Status.State == corev1alpha1.InstanceStateTerminateInstance {
		instanceID := instance.Status.EC2InstanceID
		if _, err := ec2adapterCli.Operations.TerminateInstance(&operations.TerminateInstanceParams{
			Context: ctx,
			ID:      instanceID,
		}); err != nil {
			log.Error(err, "failed to terminate instance", "instance", instanceID)
			return ctrl.Result{}, err
		}
	}

	// remove our finalizer from the list and update it.
	instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, instanceFinalizer)
	return ctrl.Result{}, r.Update(ctx, &instance)
}

// NodeMapper is a mapper reconciling Instances for Node events
type NodeMapper struct {
	cli client.Client
	log logr.Logger
}

// Map function reconciling Instance associated to node.
func (m *NodeMapper) Map(obj handler.MapObject) []reconcile.Request {
	ctx := context.Background()
	log := m.log.WithName("inputmapper")

	// Obtain the modified node
	node, ok := obj.Object.(*kcore_v1.Node)
	if !ok {
		log.Error(fmt.Errorf("fail to cast object %s into Node", obj.Meta.GetName()),
			"fail to cast object into Node", "kind", obj.Object.GetObjectKind().GroupVersionKind, "obj", obj)
		return []reconcile.Request{}
	}

	// Node is not managed by AWS, the reconciler does not support it.
	if !isAWSNode(*node) {
		log.Info("node is not managed by AWS, the reconciler does not support it.")
		return []reconcile.Request{}
	}

	// Get instance ID from node's Provider field
	// E.g for an AWS EKS / KOPS node
	// providerID: aws:///eu-west-1b/i-0d0b3844fcfb3c137
	ec2InstanceID := ec2InstanceID(*node)
	if ec2InstanceID == "" {
		log.Error(fmt.Errorf("node has no valid ProviderID"), "name", node.Name)
		return []reconcile.Request{}
	}

	// List instances
	instances := &corev1alpha1.InstanceList{}
	if err := m.cli.List(ctx, instances); err != nil {
		log.Error(err, "unable to list Instances")
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
}

// isAWSNode returns if given Node is identified as an AWS cluster Node
func isAWSNode(node kcore_v1.Node) bool {
	return !strings.HasPrefix(node.Spec.ProviderID, "aws://")
}

// ec2InstanceID returns the EC2 instance ID from a given Node
func ec2InstanceID(node kcore_v1.Node) string {
	return path.Base(node.Spec.ProviderID)
}

// SetupWithManager instantiates and returns the InstanceReconciler controller.
func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Instance{}).
		Watches(
			&source.Kind{Type: &kcore_v1.Node{}},
			&handler.EnqueueRequestsFromMapFunc{ToRequests: &NodeMapper{cli: mgr.GetClient(), log: mgr.GetLogger()}},
		).
		Complete(r)
}

// containsString remove given string from slice.
func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// instanceWithID returns the instance with the given ID or nil
func instanceWithID(slice []*models.AutoscalingGroupInstance, id string) *models.AutoscalingGroupInstance {
	for _, e := range slice {
		if e.InstanceID == id {
			return e
		}
	}
	return nil
}
