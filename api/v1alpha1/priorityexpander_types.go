/*
Copyright 2021.

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

// PriorityExpanderSpec defines the desired state of PriorityExpander
type PriorityExpanderSpec struct {
	// The Go template to parse, which will generate the priority expander
	// config map for cluster autoscaler to use.
	//+kubebuilder:validation:Required
	Template string `json:"template"`
}

// PriorityExpanderStatus defines the observed state of PriorityExpander
type PriorityExpanderStatus struct {
	// The last time the prioriry exchanger was updated.
	LastSuccessfulUpdate *metav1.Time `json:"lastSuccessfulUpdate,omitempty"`

	// State of last update.
	State string `json:"state,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PriorityExpander is the Schema for the priorityexpanders API
type PriorityExpander struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PriorityExpanderSpec   `json:"spec,omitempty"`
	Status PriorityExpanderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PriorityExpanderList contains a list of PriorityExpander
type PriorityExpanderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PriorityExpander `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PriorityExpander{}, &PriorityExpanderList{})
}
