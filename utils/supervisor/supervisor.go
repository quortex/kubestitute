// Package supervisor is used to perform ASG supoervision tasks.
package supervisor

import (
	"context"
	"time"

	"github.com/go-logr/logr"

	ec2adapter "quortex.io/kubestitute/clients/ec2adapter/client"
	"quortex.io/kubestitute/clients/ec2adapter/client/operations"
	"quortex.io/kubestitute/clients/ec2adapter/models"
	"quortex.io/kubestitute/metrics"
	"quortex.io/kubestitute/utils/helper"
)

// Supervisor describes an ASG supervisor.
type Supervisor interface {
	SetASGs(asgs []string)
}

type supervisor struct {
	asgs          map[string]struct{}
	ec2adapterCli *ec2adapter.AwsEc2Adapter
}

// New returns a new Supervisor instances with given supervision interval.
func New(interval time.Duration, log logr.Logger) Supervisor {
	// Instantiate Supervisor with aws-ec2-adapter client requirements
	sup := &supervisor{
		ec2adapterCli: ec2adapter.NewHTTPClientWithConfig(nil, &ec2adapter.TransportConfig{
			Host:    "localhost:8008",
			Schemes: []string{"http"},
		}),
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
	ctx := context.Background()
	// Retrieve AWS autoscaling group.
	log.Info("ASG supervision start", "autoscalingGroup", asgName)
	res, err := s.ec2adapterCli.Operations.GetAutoscalingGroup(&operations.GetAutoscalingGroupParams{
		Context: ctx,
		Name:    asgName,
	})
	if err != nil {
		log.Error(err, "Failed to get autoscaling group", "name", asgName)
		return
	}
	asg := res.Payload

	// Set ASG capacity metrics.
	metrics.AutoscalingGroupDesiredCapacity.WithLabelValues(asgName).Set(float64(asg.DesiredCapacity))
	metrics.AutoscalingGroupMinSize.WithLabelValues(asgName).Set(float64(asg.MinSize))
	metrics.AutoscalingGroupMaxSize.WithLabelValues(asgName).Set(float64(asg.MaxSize))
	metrics.AutoscalingGroupCapacity.WithLabelValues(asgName).Set(float64(len(
		filterInstancesWithLifecycleStates(
			asg.Instances,
			models.AutoscalingGroupInstanceLifecycleStatePending,
			models.AutoscalingGroupInstanceLifecycleStatePendingWait,
			models.AutoscalingGroupInstanceLifecycleStatePendingProceed,
			models.AutoscalingGroupInstanceLifecycleStateInService,
		),
	)))
}

// filterInstancesWithLifecycleStates returns a filtered slice of AutoscalingGroupInstance with given lifecycleStates.
func filterInstancesWithLifecycleStates(inst []*models.AutoscalingGroupInstance, lifecycleStates ...string) []*models.AutoscalingGroupInstance {
	res := []*models.AutoscalingGroupInstance{}
	for _, e := range inst {
		if helper.ContainsString(lifecycleStates, e.LifecycleState) {
			res = append(res, e)
		}
	}
	return res
}
