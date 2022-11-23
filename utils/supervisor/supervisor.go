// Package supervisor is used to perform ASG supoervision tasks.
package supervisor

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/go-logr/logr"

	"quortex.io/kubestitute/metrics"
	"quortex.io/kubestitute/utils/helper"
)

// Supervisor describes an ASG supervisor.
type Supervisor interface {
	SetASGs(asgs []string)
}

type supervisor struct {
	asgs        map[string]struct{}
	autoscaling *autoscaling.AutoScaling
}

// New returns a new Supervisor instances with given supervision interval.
func New(autoscaling *autoscaling.AutoScaling, interval time.Duration, log logr.Logger) Supervisor {
	// Instantiate Supervisor with aws-ec2-adapter client requirements
	sup := &supervisor{
		autoscaling: autoscaling,
	}

	// Configure ticker to refresh ASGs metrics based on time interval.
	t := time.NewTicker(interval)
	go func() {
		for range t.C {
			for k := range sup.asgs {
				go sup.refreshMetrics(k, log)
			}
		}
	}()

	return sup
}

// SetASGs reconfigure Supervisor to watch for given ASG names.
func (s *supervisor) SetASGs(asgs []string) {
	set := make(map[string]struct{})
	for _, e := range asgs {
		set[e] = struct{}{}
	}
	s.asgs = set
}

// refreshMetrics refresh metrics for given asg name.
func (s *supervisor) refreshMetrics(asgName string, log logr.Logger) {
	// Retrieve AWS autoscaling group.
	log.Info("ASG supervision start", "autoscalingGroup", asgName)
	res, err := s.autoscaling.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{&asgName},
	})
	if err == nil && len(res.AutoScalingGroups) == 0 {
		err = fmt.Errorf("autoscaling group not found")
	}
	if err != nil {
		log.Error(err, "Failed to get autoscaling group", "name", asgName)
		return
	}
	asg := res.AutoScalingGroups[0]

	// Set ASG capacity metrics.
	metrics.AutoscalingGroupDesiredCapacity.WithLabelValues(asgName).Set(float64(aws.Int64Value(asg.DesiredCapacity)))
	metrics.AutoscalingGroupMinSize.WithLabelValues(asgName).Set(float64(aws.Int64Value(asg.MinSize)))
	metrics.AutoscalingGroupMaxSize.WithLabelValues(asgName).Set(float64(aws.Int64Value(asg.MaxSize)))
	metrics.AutoscalingGroupCapacity.WithLabelValues(asgName).Set(float64(len(
		filterInstancesWithLifecycleStates(
			asg.Instances,
			autoscaling.LifecycleStatePending,
			autoscaling.LifecycleStatePendingWait,
			autoscaling.LifecycleStatePendingProceed,
			autoscaling.LifecycleStateInService,
		),
	)))
}

// filterInstancesWithLifecycleStates returns a filtered slice of AutoscalingGroupInstance with given lifecycleStates.
func filterInstancesWithLifecycleStates(inst []*autoscaling.Instance, lifecycleStates ...string) []*autoscaling.Instance {
	res := []*autoscaling.Instance{}
	for _, e := range inst {
		if helper.ContainsString(lifecycleStates, aws.StringValue(e.LifecycleState)) {
			res = append(res, e)
		}
	}
	return res
}
