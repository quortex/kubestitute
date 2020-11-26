package clusterautoscaler

import (
	"time"

	corev1alpha1 "quortex.io/kubestitute/api/v1alpha1"
)

// GetNodeGroupWithName returns the NodeGroup in slice matching name.
func GetNodeGroupWithName(nodeGroups []NodeGroup, name string) *NodeGroup {
	for _, e := range nodeGroups {
		if e.Name == name {
			return &e
		}
	}
	return nil
}

// NodeGroupIntOrFieldValue returns the desired value matching IntOrField.
// Field returns the NodeGroup Field value ans has priority over Int if a valid
// Field is given.
func NodeGroupIntOrFieldValue(ng NodeGroup, iof corev1alpha1.IntOrField) int32 {
	switch iof.FieldVal {
	case corev1alpha1.FieldNone:
		return iof.IntVal
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

// A function for test purpose
type now func() time.Time

// MatchPolicy returns if given NodeGroup match desired Scheduler policy.
func (n NodeGroup) MatchPolicy(policy corev1alpha1.SchedulerPolicy) bool {
	return n.matchPolicy(policy, time.Now)
}

func (n NodeGroup) matchPolicy(policy corev1alpha1.SchedulerPolicy, now now) bool {
	from := NodeGroupIntOrFieldValue(n, policy.From)
	to := NodeGroupIntOrFieldValue(n, policy.To)

	// Get seconds between now and last transition time on node group.
	// If policy period is not compliant, it fail.
	diff := now().Sub(n.Health.LastTransitionTime).Seconds()
	if policy.PeriodSeconds < int32(diff) {
		return false
	}

	// Perform arithmetic comparison to compute Scheduler policy.
	switch policy.ArithmeticOperator {
	case corev1alpha1.ArithmeticOperatorEqual:
		return from == to
	case corev1alpha1.ArithmeticOperatorNotEqual:
		return from != to
	case corev1alpha1.ArithmeticOperatorGreaterThan:
		return from > to
	case corev1alpha1.ArithmeticOperatorGreaterThanOrEqual:
		return from >= to
	case corev1alpha1.ArithmeticOperatorLowerThan:
		return from < to
	case corev1alpha1.ArithmeticOperatorLowerThanOrEqual:
		return from <= to
	}

	return false
}
