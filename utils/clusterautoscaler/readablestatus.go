package clusterautoscaler

import (
	"time"
)

// Status contains ClusterAutoscaler Status.
// Deprecated: Use ClusterAutoscalerStatus instead.
type Status struct {
	Time        time.Time
	ClusterWide ClusterWide
	NodeGroups  []NodeGroup
}

// ClusterWide is the global (cluster wide )
// ClusterAutoscaler status.
// Deprecated: Use ClusterWideStatus instead.
type ClusterWide struct {
	Health    Health
	ScaleDown ScaleDown
	ScaleUp   ScaleUp
}

// NodeGroup is the ClusterAutoscaler status
// by node group.
// Deprecated: Use NodeGroupStatus instead.
type NodeGroup struct {
	Name      string
	Health    NodeGroupHealth
	ScaleDown ScaleDown
	ScaleUp   ScaleUp
}

// HealthStatus describes ClusterAutoscaler status
// for Node groups Healthness.
// Deprecated: Use ClusterHealthCondition instead.
type HealthStatus string

const (
	// HealthStatusHealthy status means that the cluster is in a good shape.
	// Deprecated: Use ClusterAutoscalerHealthy instead.
	HealthStatusHealthy HealthStatus = "Healthy"
	// HealthStatusUnhealthy status means that the cluster is in a bad shape.
	// Deprecated: Use ClusterAutoscalerUnhealthy instead.
	HealthStatusUnhealthy HealthStatus = "Unhealthy"
)

// Health describes the cluster wide cluster autoscaler
// Health condition.
// Deprecated: Use ClusterHealthCondition instead.
type Health struct {
	Status                                                   HealthStatus
	Ready, Unready, NotStarted, Registered, LongUnregistered int32
	LastProbeTime                                            time.Time
	LastTransitionTime                                       time.Time
}

// NodeGroupHealth describes the individual node group cluster autoscaler
// Health condition.
// Deprecated: Use NodeGroupHealthCondition instead.
type NodeGroupHealth struct {
	Health
	CloudProviderTarget, MinSize, MaxSize int32
}

// ScaleDownStatus describes ClusterAutoscaler status
// for Node groups ScaleDown.
// Deprecated: Use ClusterAutoscalerConditionStatus instead.
type ScaleDownStatus string

const (
	// ScaleDownCandidatesPresent status means that there's candidates for scale down.
	// Deprecated: Use ClusterAutoscalerCandidatesPresent instead.
	ScaleDownCandidatesPresent ScaleDownStatus = "CandidatesPresent"
	// ScaleDownNoCandidates status means that there's no candidates for scale down.
	// Deprecated: Use ClusterAutoscalerNoCandidates instead.
	ScaleDownNoCandidates ScaleDownStatus = "NoCandidates"
)

// ScaleDown describes ClusterAutoscaler condition
// for Node groups ScaleDown.
// Deprecated: Use ScaleDownCondition instead.
type ScaleDown struct {
	Status             ScaleDownStatus
	Candidates         int32
	LastProbeTime      time.Time
	LastTransitionTime time.Time
}

// ScaleUpStatus describes ClusterAutoscaler status
// for Node groups ScaleUp.
// Deprecated: Use ClusterAutoscalerConditionStatus instead.
type ScaleUpStatus string

const (
	// ScaleUpNeeded status means that scale up is needed.
	// Deprecated: Use ClusterAutoscalerNeeded instead.
	ScaleUpNeeded ScaleUpStatus = "Needed"
	// ScaleUpNotNeeded status means that scale up is not needed.
	// Deprecated: Use ClusterAutoscalerNotNeeded instead.
	ScaleUpNotNeeded ScaleUpStatus = "NotNeeded"
	// ScaleUpInProgress status means that scale up is in progress.
	// Deprecated: Use ClusterAutoscalerInProgress instead.
	ScaleUpInProgress ScaleUpStatus = "InProgress"
	// ScaleUpNoActivity status means that there has been no scale up activity recently.
	// Deprecated: Use ClusterAutoscalerNoActivity instead.
	ScaleUpNoActivity ScaleUpStatus = "NoActivity"
	// ScaleUpBackoff status means that due to a recently failed scale-up no further scale-ups attempts will be made for some time.
	// Deprecated: Use ClusterAutoscalerBackoff instead.
	ScaleUpBackoff ScaleUpStatus = "Backoff"
)

// ScaleUp describes ClusterAutoscaler condition
// for Node groups ScaleUp.
// Deprecated: Use ClusterScaleUpCondition or NodeGroupScaleUpCondition instead.
type ScaleUp struct {
	Status             ScaleUpStatus
	LastProbeTime      time.Time
	LastTransitionTime time.Time
}

// GetNodeGroupWithName returns the NodeGroup in slice matching name.
func GetNodeGroupWithName(nodeGroups []NodeGroupStatus, name string) *NodeGroupStatus {
	for _, e := range nodeGroups {
		if e.Name == name {
			return &e
		}
	}
	return nil
}
