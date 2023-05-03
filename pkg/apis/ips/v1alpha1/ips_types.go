package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:resource:scope="Cluster",singular="ips",path="ipses"
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
	AllocatedIPs map[string]AllocatedPod `json:"allocatedIPs,omitempty"`

	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Optional
	TotalIPCount int `json:"totalIPCount,omitempty"`

	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Optional
	AllocatedIPCount int `json:"allocatedIPCount,omitempty"`
}

type AllocatedPod struct {
	Pod    string `json:"pod,omitempty"`
	PodUid string `json:"poduid,omitempty"`
}
