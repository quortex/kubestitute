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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SchedulerTrigger describe a trigger for the Scheduler.
type SchedulerTrigger string

// All defined SchedulerTriggers
const (
	SchedulerTriggerClusterAutoscaler SchedulerTrigger = "ClusterAutoscaler"
)

// SchedulerSpec defines the desired state of Scheduler
type SchedulerSpec struct {

	// The Scheduler Trigger
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:={"ClusterAutoscaler"}
	Trigger SchedulerTrigger `json:"trigger,omitempty"`

	// The name of the autoscaling group, which the scheduler will use to
	// apply the rules.
	// +kubebuilder:validation:Required
	ASGTarget string `json:"autoscalingGroupTarget"`

	// The name of the autoscaling group, in which the scheduler will trigger
	// fallback instances.
	// This autoscaling group must not be managed by the cluster-autoscaler !!!
	// +kubebuilder:validation:Required
	ASGFallback string `json:"autoscalingGroupFallback"`

	// Scheduler rules used to match criteria on Target ASG to trigger Scale Up
	// on Fallback ASG.
	// +kubebuilder:validation:Optional
	ScaleUpRules ScaleUpRules `json:"scaleUpRules"`

	// Scheduler rules used to match criteria on Target ASG to trigger Scale Down
	// on Fallback ASG.
	// +kubebuilder:validation:Optional
	ScaleDownRules ScaleDownRules `json:"scaleDownRules"`
}

// ScaleDownRules configures the scaling behavior for Instance scale downs.
type ScaleDownRules struct {
	// A cooldown for consecutive scale down operations.
	// +kubebuilder:validation:Optional
	StabilizationWindowSeconds int32 `json:"stabilizationWindowSeconds"`

	// Policies is a list of potential scaling polices which can be evaluated for scaling decisions.
	// At least one policy must be specified.
	// Instances will be scaled down one by one.
	// +kubebuilder:validation:Required
	Policies []SchedulerPolicy `json:"policies,omitempty"`
}

// ScaleUpRules configures the scaling behavior for Instance scale ups.
type ScaleUpRules struct {
	// A cooldown for consecutive scale up operations.
	// +kubebuilder:validation:Optional
	StabilizationWindowSeconds int32 `json:"stabilizationWindowSeconds"`

	// Policies is a list of potential scaling polices which can be evaluated for scaling decisions.
	// At least one policy must be specified.
	// For scale ups the matching policy which triggers the highest number of replicas
	// will be used.
	// +kubebuilder:validation:Required
	Policies []AdvancedSchedulerPolicy `json:"policies,omitempty"`
}

// SchedulerPolicy is a single policy which must hold true for a specified past interval.
type SchedulerPolicy struct {
	// LeftOperand is the left operand of the comparison. It could be the target ASG Health field from
	// which this policy is applied or an integer.
	// +kubebuilder:validation:Required
	LeftOperand IntOrField `json:"leftOperand"`

	// A comparison operator used to apply policy between From and To.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:={"equal","notEqual","greaterThan","greaterThanOrEqual","lowerThan","lowerThanOrEqual"}
	Operator ComparisonOperator `json:"operator"`

	// RightOperand is the left operand of the comparison. It could be the target ASG Health field from
	// which this policy is applied or an integer.
	// +kubebuilder:validation:Required
	RightOperand IntOrField `json:"rightOperand"`

	// PeriodSeconds specifies the window of time for which the policy should hold true.
	// +kubebuilder:validation:Optional
	PeriodSeconds int32 `json:"periodSeconds"`
}

// AdvancedSchedulerPolicy is a policy that allow arithmetic operation to compute replicas.
type AdvancedSchedulerPolicy struct {
	SchedulerPolicy `json:",inline"`

	// Replicas specify the replicas to Scale.
	// +kubebuilder:validation:Required
	Replicas IntOrArithmeticOperation `json:"replicas"`
}

// IntOrField is a type that can hold an int32 or a Field.
type IntOrField struct {
	// An Int for value.
	// +kubebuilder:validation:Optional
	IntVal int32 `json:"int,omitempty"`

	// An Field for value.
	// +kubebuilder:validation:Enum:={"Ready","Unready","NotStarted","LongNotStarted","Registered","LongUnregistered","CloudProviderTarget"}
	// +kubebuilder:validation:Optional
	FieldVal *Field `json:"field,omitempty"`
}

// IntOrArithmeticOperation is a type that can hold an int32 or
// an arithmetic operation.
type IntOrArithmeticOperation struct {
	// An Int for value.
	// +kubebuilder:validation:Optional
	IntVal int32 `json:"int,omitempty"`

	// An arithmetic operation..
	// +kubebuilder:validation:Optional
	OperationVal *ArithmeticOperation `json:"operation,omitempty"`
}

// Field describes a SchedulerPolicy Field.
// It is based on ASG health status.
type Field string

// All Field constants
const (
	FieldReady               Field = "Ready"
	FieldUnready             Field = "Unready"
	FieldNotStarted          Field = "NotStarted"
	FieldLongNotStarted      Field = "LongNotStarted"
	FieldRegistered          Field = "Registered"
	FieldLongUnregistered    Field = "LongUnregistered"
	FieldCloudProviderTarget Field = "CloudProviderTarget"
)

// ComparisonOperator describes comparison operators
type ComparisonOperator string

// All ComparisonOperator constants
const (
	ComparisonOperatorEqual              = "equal"
	ComparisonOperatorNotEqual           = "notEqual"
	ComparisonOperatorGreaterThan        = "greaterThan"
	ComparisonOperatorGreaterThanOrEqual = "greaterThanOrEqual"
	ComparisonOperatorLowerThan          = "lowerThan"
	ComparisonOperatorLowerThanOrEqual   = "lowerThanOrEqual"
)

// ArithmeticOperator describes arithmetic operators
type ArithmeticOperator string

// All ArithmeticOperator constants
const (
	ArithmeticOperatorPlus     = "plus"
	ArithmeticOperatorMinus    = "minus"
	ArithmeticOperatorMultiply = "multiply"
	ArithmeticOperatorDivide   = "divide"
)

// ArithmeticOperation describes an arithmetic operation.
type ArithmeticOperation struct {
	// LeftOperand is the left operand of the operation.
	// +kubebuilder:validation:Required
	LeftOperand IntOrField `json:"leftOperand"`

	// An arithmetic operator used to apply policy between From and To.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:={"plus","minus","multiply","divide"}
	Operator ArithmeticOperator `json:"operator"`

	// RightOperand is the right operand of the operation.
	// +kubebuilder:validation:Required
	RightOperand IntOrField `json:"rightOperand"`
}

// SchedulerStatus defines the observed state of Scheduler
type SchedulerStatus struct {
	// The last time this scheduler has perform a scale up.
	LastScaleUp *metav1.Time `json:"lastScaleUp,omitempty"`
	// The last time this scheduler has perform a scale down.
	LastScaleDown *metav1.Time `json:"lastScaleDown,omitempty"`
}

// +kubebuilder:object:root=true

// Scheduler is the Schema for the schedulers API
type Scheduler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SchedulerSpec   `json:"spec,omitempty"`
	Status SchedulerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SchedulerList contains a list of Scheduler
type SchedulerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Scheduler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Scheduler{}, &SchedulerList{})
}
