/*
Copyright 2023 The fast Authors.

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

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:subresource:status
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Ips is the Schema for the fast api
type Ips struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IpsSpec   `json:"spec,omitempty"`
	Status IpsStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IpsList contains a list of Ips
type IpsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ips `json:"items"`
}

// IpsSpec defines the desired state of Ips
type IpsSpec struct {
	// +kubebuilder:validation:Required
	Subnet string `json:"subnet"`

	// +kubebuilder:validation:Optional
	IPs []string `json:"ips,omitempty"`

	// +kubebuilder:validation:Optional
	PodAffinity *metav1.LabelSelector `json:"podAffinity,omitempty"`

	// +kubebuilder:validation:Optional
	NamespaceAffinity *metav1.LabelSelector `json:"namespaceAffinity,omitempty"`

	// +kubebuilder:validation:Optional
	NodeAffinity *metav1.LabelSelector `json:"nodeAffinity,omitempty"`
}

// IpsStatus defines the observed state of Ips
type IpsStatus struct {
	// +kubebuilder:validation:Optional
	AllocatedIPs *string `json:"allocatedIPs,omitempty"`

	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Optional
	TotalIPCount *int64 `json:"totalIPCount,omitempty"`

	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Optional
	AllocatedIPCount *int64 `json:"allocatedIPCount,omitempty"`
}
