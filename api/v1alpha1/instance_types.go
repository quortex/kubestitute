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

// InstanceSpec defines the desired state of Instance
type InstanceSpec struct {

	// The AutoScaling Group name.
	// +kubebuilder:validation:Required
	ASG string `json:"asg"`

	// Indicates whether Amazon EC2 Auto Scaling waits for the cooldown period to
	// complete before initiating a scaling activity to set your Auto Scaling group
	// to its new capacity. By default, Amazon EC2 Auto Scaling does not honor the
	// cooldown period during manual scaling activities.
	// +kubebuilder:validation:Optional
	HonorCooldown bool `json:"honorCooldown"`
}

// InstanceState describes the instance state.
type InstanceState string

// All defined InstanceStates
const (
	InstanceStateNone           InstanceState = ""
	InstanceStateTriggerScaling InstanceState = "Trigger Scaling"
	InstanceStateWaitInstance   InstanceState = "Waiting Instance"
	InstanceStateReady          InstanceState = "Ready"
)

// InstanceStatus defines the observed state of Instance
type InstanceStatus struct {

	// The current state of the instance
	State InstanceState `json:"state,omitempty"`

	// The current state of the instance
	EC2InstanceID string `json:"ec2InstanceID,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.state",description="The Instance status"
// +kubebuilder:printcolumn:name="EC2 INSTANCE",type="string",JSONPath=".status.ec2InstanceID",description="The EC2 Instance ID"

// Instance is the Schema for the instances API
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSpec   `json:"spec,omitempty"`
	Status InstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// InstanceList contains a list of Instance
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Instance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Instance{}, &InstanceList{})
}
