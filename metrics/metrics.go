package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	namespace = "kubestitute"
)

// All custom metrics.
var (
	ScaledUpNodesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "scaled_up_nodes_total",
			Help:      "Number of nodes added by kubestitute.",
		},
		[]string{
			"autoscaling_group_name",
			"scheduler_name",
		},
	)
	ScaledDownNodesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "scaled_down_nodes_total",
			Help:      "Number of nodes removed by kubestitute.",
		},
		[]string{
			"autoscaling_group_name",
			"scheduler_name",
		},
	)
	EvictedPodsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "evicted_pods_total",
			Help:      "Number of pods evicted by kubestitute.",
		},
		[]string{
			"autoscaling_group_name",
			"node_name",
			"scheduler_name",
		},
	)
	AutoscalingGroupDesiredCapacity = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "autoscaling_group_desired_capacity",
			Help:      "The desired size of the autoscaling group.",
		},
		[]string{
			"autoscaling_group_name",
		},
	)
	AutoscalingGroupCapacity = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "autoscaling_group_capacity",
			Help:      "The current autoscaling group capacity (Pending and InService instances).",
		},
		[]string{
			"autoscaling_group_name",
		},
	)
	AutoscalingGroupMinSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "autoscaling_group_min_size",
			Help:      "The minimum size of the autoscaling group.",
		},
		[]string{
			"autoscaling_group_name",
		},
	)
	AutoscalingGroupMaxSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "autoscaling_group_max_size",
			Help:      "The maximum size of the autoscaling group.",
		},
		[]string{
			"autoscaling_group_name",
		},
	)
	PriorityExpanderTemplateError = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "priority_expander_template_error",
			Help:      "Is 1 if template is unparsable.",
		},
	)
	SchedulerTargetNodeGroupStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "scheduler_node_group_status",
			Help:      "Represents status of the target node group.",
		},
		[]string{"node_group_name", "scale_up_status"},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(
		ScaledUpNodesTotal,
		ScaledDownNodesTotal,
		EvictedPodsTotal,
		AutoscalingGroupDesiredCapacity,
		AutoscalingGroupCapacity,
		AutoscalingGroupMinSize,
		AutoscalingGroupMaxSize,
		PriorityExpanderTemplateError,
		SchedulerTargetNodeGroupStatus,
	)
}
