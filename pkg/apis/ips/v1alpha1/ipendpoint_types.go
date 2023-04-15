package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +kubebuilder:subresource:status
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IpEndpoint is the Schema for the fast api
type IpEndpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status IpsStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IpEndpointList contains a list of IpEndpoint
type IpEndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ips `json:"items"`
}

// IpEndpointStatus defines the observed state of Ips
type IpEndpointStatus struct {
	// +kubebuilder:validation:Required
	UID string `json:"uid"`

	// +kubebuilder:validation:Required
	Node string `json:"node"`

	// +kubebuilder:validation:Required
	IPs IPAllocationDetail `json:"ips"`
}

type IPAllocationDetail struct {
	// +kubebuilder:validation:Required
	NIC string `json:"interface"`

	// +kubebuilder:validation:Optional
	IPv4 *string `json:"ipv4,omitempty"`

	// +kubebuilder:validation:Optional
	IPv6 *string `json:"ipv6,omitempty"`

	// +kubebuilder:validation:Optional
	IPv4Pool *string `json:"ipv4Pool,omitempty"`

	// +kubebuilder:validation:Optional
	IPv6Pool *string `json:"ipv6Pool,omitempty"`
}
