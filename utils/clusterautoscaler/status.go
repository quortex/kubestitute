package clusterautoscaler

import (
	"time"
)

// Status contains ClusterAutoscaler status.
type Status struct {
	Time        time.Time
	ClusterWide ClusterWide
	NodeGroups  []NodeGroup
}

// ClusterWide is the global (cluster wide )
// ClusterAutoscaler status.
type ClusterWide struct {
	Health    Health
	ScaleDown ScaleDown
	ScaleUp   ScaleUp
}

// NodeGroup is the ClusterAutoscaler status
// by node group.
type NodeGroup struct {
	Name      string
	Health    NodeGroupHealth
	ScaleDown ScaleDown
	ScaleUp   ScaleUp
}

// HealthStatus describes ClusterAutoscaler status
// for Node groups Healthness.
type HealthStatus string

const (
	// HealthStatusHealthy status means that the cluster is in a good shape.
	HealthStatusHealthy HealthStatus = "Healthy"
	// HealthStatusUnhealthy status means that the cluster is in a bad shape.
	HealthStatusUnhealthy HealthStatus = "Unhealthy"
)

// Health describes the cluster wide cluster autoscaler
// Health condition.
type Health struct {
	Status                                                   HealthStatus
	Ready, Unready, NotStarted, Registered, LongUnregistered int32
	LastProbeTime                                            time.Time
	LastTransitionTime                                       time.Time
}

// NodeGroupHealth describes the individual node group cluster autoscaler
// Health condition.
type NodeGroupHealth struct {
	Health
	CloudProviderTarget, MinSize, MaxSize int32
}

// ScaleDownStatus describes ClusterAutoscaler status
// for Node groups ScaleDown.
type ScaleDownStatus string

const (
	// ScaleDownCandidatesPresent status means that there's candidates for scale down.
	ScaleDownCandidatesPresent ScaleDownStatus = "CandidatesPresent"
	// ScaleDownNoCandidates status means that there's no candidates for scale down.
	ScaleDownNoCandidates ScaleDownStatus = "NoCandidates"
)

// ScaleDown describes ClusterAutoscaler condition
// for Node groups ScaleDown.
type ScaleDown struct {
	Status             ScaleDownStatus
	Candidates         int32
	LastProbeTime      time.Time
	LastTransitionTime time.Time
}

// ScaleUpStatus describes ClusterAutoscaler status
// for Node groups ScaleUp.
type ScaleUpStatus string

const (
	// ScaleUpNeeded status means that scale up is needed.
	ScaleUpNeeded ScaleUpStatus = "Needed"
	// ScaleUpNotNeeded status means that scale up is not needed.
	ScaleUpNotNeeded ScaleUpStatus = "NotNeeded"
	// ScaleUpInProgress status means that scale up is in progress.
	ScaleUpInProgress ScaleUpStatus = "InProgress"
	// ScaleUpNoActivity status means that there has been no scale up activity recently.
	ScaleUpNoActivity ScaleUpStatus = "NoActivity"
	// ScaleUpBackoff status means that due to a recently failed scale-up no further scale-ups attempts will be made for some time.
	ScaleUpBackoff ScaleUpStatus = "Backoff"
)

// ScaleUp describes ClusterAutoscaler condition
// for Node groups ScaleUp.
type ScaleUp struct {
	Status             ScaleUpStatus
	LastProbeTime      time.Time
	LastTransitionTime time.Time
}

// GetNodeGroupWithName returns the NodeGroup in slice matching name.
func GetNodeGroupWithName(nodeGroups []NodeGroup, name string) *NodeGroup {
	for _, e := range nodeGroups {
		if e.Name == name {
			return &e
		}
	}
	return nil
}
