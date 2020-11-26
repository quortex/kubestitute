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
	// +kubebuilder:validation:Required
	ASGFallback string `json:"autoscalingGroupFallback"`

	// Scheduler rules used to match criteria on Target ASG to trigger Scale Up
	// on Fallback ASG.
	// +kubebuilder:validation:Optional
	ScaleUpRules SchedulerRules `json:"scaleUpRules"`

	// Scheduler rules used to match criteria on Target ASG to trigger Scale Down
	// on Fallback ASG.
	// +kubebuilder:validation:Optional
	ScaleDownRules SchedulerRules `json:"scaleDownRules"`
}

// SchedulerRules configures the scaling behavior for Instance scheduling.
type SchedulerRules struct {
	// StabilizationWindowSeconds is the number of seconds for which past recommendations should be
	// considered while scaling up or scaling down.
	// +kubebuilder:validation:Optional
	StabilizationWindowSeconds int32 `json:"stabilizationWindowSeconds"`

	// Policies is a list of potential scaling polices which can be used during scaling.
	// At least one policy must be specified.
	// +kubebuilder:validation:Required
	Policies []SchedulerPolicy `json:"policies,omitempty"`
}

// SchedulerPolicy is a single policy which must hold true for a specified past interval.
type SchedulerPolicy struct {
	// From is the target ASG Health field from which this policy is applied.
	// +kubebuilder:validation:Required
	From IntOrField `json:"from"`

	// An arithmetic operator used to apply policy between From and To.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:={"=","!=",">",">=","<","<="}
	ArithmeticOperator ArithmeticOperator `json:"operator"`

	// To is the target ASG Health field to which this policy is applied or a fixed value.
	// +kubebuilder:validation:Required
	To IntOrField `json:"to"`

	// PeriodSeconds specifies the window of time for which the policy should hold true.
	// +kubebuilder:validation:Optional
	PeriodSeconds int32 `json:"periodSeconds"`
}

// Type represents the stored type of IntOrField.
type Type int64

// Type constants
const (
	AnInt  Type = iota // The IntOrString holds an int.
	AField             // The IntOrString holds a Field.
)

// IntOrField is a type that can hold an int32 or a Field.
type IntOrField struct {
	// An Int for value.
	// +kubebuilder:validation:Optional
	IntVal int32 `json:"int"`

	// An Field for value.
	// +kubebuilder:validation:Enum:={"Ready","Unready","NotStarted","LongNotStarted","Registered","LongUnregistered","CloudProviderTarget"}
	FieldVal Field `json:"field"`
}

// Field describes a SchedulerPolicy Field.
// It is based on ASG health status.
type Field string

// All Field constants
const (
	FieldReady               = "Ready"
	FieldUnready             = "Unready"
	FieldNotStarted          = "NotStarted"
	FieldLongNotStarted      = "LongNotStarted"
	FieldRegistered          = "Registered"
	FieldLongUnregistered    = "LongUnregistered"
	FieldCloudProviderTarget = "CloudProviderTarget"
)

// ArithmeticOperator describes arithmetic operators
type ArithmeticOperator string

// All ArithmeticOperator constants
const (
	ArithmleticOperatorEqual              = "="
	ArithmleticOperatorNotEqual           = "!="
	ArithmleticOperatorGreaterThan        = ">"
	ArithmleticOperatorGreaterThanOrEqual = ">="
	ArithmleticOperatorLowerThan          = "<"
	ArithmleticOperatorLowerThanOrEqual   = "<="
)

// SchedulerStatus defines the observed state of Scheduler
type SchedulerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
