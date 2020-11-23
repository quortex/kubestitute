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
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

// +kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=instances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.kubestitute.quortex.io,resources=instances/status,verbs=get;update;patch

func (r *InstanceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("instance", req.NamespacedName)

	var instance corev1alpha1.Instance
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		log.Error(err, "unable to fetch Instance")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Set Instance TriggerScaling state.
	if instance.Status.State == corev1alpha1.InstanceStateNone {
		instance.Status.State = corev1alpha1.InstanceStateTriggerScaling
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
	res, err := ec2adapterCli.Operations.GetAutoscalingGroup(&operations.GetAutoscalingGroupParams{
		Context: ctx,
		Name:    instance.Spec.ASG,
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	asg := res.Payload

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
		return ctrl.Result{}, r.Update(ctx, &instance)
	}

	// Try to get a new joining instance in the ASG to associate it to our Instance.
	if instance.Status.State == corev1alpha1.InstanceStateWaitInstance {
		ec2Instances := asg.Instances
		// List all instances
		instances := &corev1alpha1.InstanceList{}
		if err := r.List(ctx, instances); err != nil {
			log.Error(err, "failed to list instances")
			return ctrl.Result{}, err
		}

		// We remove EC2 instances already handled by Instance resources from slice.
		for i, e := range ec2Instances {
			for _, inst := range instances.Items {
				if inst.Status.EC2InstanceID == "" {
					break
				}
				if e.InstanceID == inst.Status.EC2InstanceID {
					ec2Instances = append(ec2Instances[:i], ec2Instances[i+1:]...)
					break
				}
			}
		}

		// Seems that new instance has not joined yet the ASG, we'll try it later.
		if len(ec2Instances) == 0 {
			log.Info("instance has not joined yet the ASG, retry in 5 sec")
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}

		// Seems that instances in ASG are sorted from newer to older,
		// anyway that's not a big deal to associate an instance that has
		// not been scheduled, by this Instance resource.
		ec2Inst := ec2Instances[len(ec2Instances)-1]
		instance.Status.State = corev1alpha1.InstanceStateReady
		instance.Status.EC2InstanceID = ec2Inst.InstanceID

		return ctrl.Result{}, r.Update(ctx, &instance)
	}

	return ctrl.Result{}, nil
}

func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Instance{}).
		Complete(r)
}
