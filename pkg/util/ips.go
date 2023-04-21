package util

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	ipsv1alpha1 "github.com/fast-io/fast/pkg/apis/ips/v1alpha1"
	"github.com/fast-io/fast/pkg/scheme"
)

func NewIpEndpoint(ip string, pod *corev1.Pod, ips *ipsv1alpha1.Ips) (*ipsv1alpha1.IpEndpoint, error) {
	ipep := &ipsv1alpha1.IpEndpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
		Status: ipsv1alpha1.IpEndpointStatus{
			UID:  string(pod.UID),
			Node: pod.Spec.NodeName,
			IPs: ipsv1alpha1.IPAllocationDetail{
				IPv4:     ip,
				IPv4Pool: ips.Name,
			},
		},
	}

	if err := controllerutil.SetOwnerReference(pod, ipep, scheme.Scheme); err != nil {
		return nil, err
	}
	return ipep, nil
}
