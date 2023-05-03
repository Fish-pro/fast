package ipsmanager

import (
	"context"
	"fmt"
	"net"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	ipsv1alpha1 "github.com/fast-io/fast/pkg/apis/ips/v1alpha1"
	ipsversioned "github.com/fast-io/fast/pkg/generated/clientset/versioned"
	"github.com/fast-io/fast/pkg/scheme"
	"github.com/fast-io/fast/pkg/util"
)

const (
	IpsPodAnnotation = "fast.io/ips"
	DefaultIpsName   = "default-ips"
)

type IpsManager interface {
	AllocateIP(ctx context.Context, pod *corev1.Pod) (*ipsv1alpha1.Ips, net.IP, error)
	ReleaseIP(ctx context.Context, pod *corev1.Pod) error
	UpdateIpsStatus(ctx context.Context, ips *ipsv1alpha1.Ips, nowStatus ipsv1alpha1.IpsStatus) error
	CreateOrUpdateIpEndpoint(ctx context.Context, ipep *ipsv1alpha1.IpEndpoint) error
	NewIpEndpoint(ip string, pod *corev1.Pod, ips *ipsv1alpha1.Ips) (*ipsv1alpha1.IpEndpoint, error)
}

type ipsManager struct {
	client ipsversioned.Interface
}

func NewIpsManager(client ipsversioned.Interface) IpsManager {
	return &ipsManager{client: client}
}

func (c *ipsManager) AllocateIP(ctx context.Context, pod *corev1.Pod) (*ipsv1alpha1.Ips, net.IP, error) {
	ipsName := pod.Annotations[IpsPodAnnotation]
	if len(ipsName) == 0 {
		ipsName = DefaultIpsName
	}
	ips, err := c.client.SampleV1alpha1().Ipses().Get(ctx, ipsName, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	if ips.Status.AllocatedIPCount >= ips.Status.TotalIPCount {
		return nil, nil, fmt.Errorf("ips %s/%s not enough ip addresses to allocate", ips.Namespace, ips.Name)
	}

	excludeIps := make([]string, 0)
	for k := range ips.Status.AllocatedIPs {
		excludeIps = append(excludeIps, k)
	}

	canAllocateIps := util.ExcludeIPs(ips.Spec.IPs, excludeIps)
	if len(canAllocateIps) == 0 {
		return nil, nil, fmt.Errorf("ips %s/%s not enough ip addresses to allocate", ips.Namespace, ips.Name)
	}
	ip := net.ParseIP(canAllocateIps[0].String())

	nowStatus := ips.Status
	if nowStatus.AllocatedIPs == nil {
		nowStatus.AllocatedIPs = make(map[string]ipsv1alpha1.AllocatedPod)
	}
	nowStatus.AllocatedIPs[ip.String()] = ipsv1alpha1.AllocatedPod{
		Pod:    fmt.Sprintf("%s/%s", pod.Namespace, pod.Name),
		PodUid: string(pod.UID),
	}

	if err := c.UpdateIpsStatus(ctx, ips, nowStatus); err != nil {
		return nil, nil, err
	}
	return ips, ip, nil
}

func (c *ipsManager) ReleaseIP(ctx context.Context, pod *corev1.Pod) error {
	ipsName := pod.Annotations[IpsPodAnnotation]
	if len(ipsName) == 0 {
		ipsName = DefaultIpsName
	}
	ips, err := c.client.SampleV1alpha1().Ipses().Get(ctx, ipsName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if ips == nil || ips.Status.AllocatedIPs == nil || len(ips.Status.AllocatedIPs) == 0 {
		return nil
	}

	nowStatus := ips.Status
	delete(nowStatus.AllocatedIPs, pod.Status.PodIP)
	if nowStatus.AllocatedIPCount > 0 {
		nowStatus.AllocatedIPCount--
	}
	return c.UpdateIpsStatus(ctx, ips, nowStatus)
}

func (c *ipsManager) UpdateIpsStatus(ctx context.Context, ips *ipsv1alpha1.Ips, nowStatus ipsv1alpha1.IpsStatus) error {
	if !equality.Semantic.DeepEqual(ips.Status, nowStatus) {
		ips.Status = nowStatus
		return retry.RetryOnConflict(retry.DefaultRetry, func() error {
			_, updateErr := c.client.SampleV1alpha1().Ipses().UpdateStatus(ctx, ips, metav1.UpdateOptions{})
			if updateErr == nil {
				return nil
			}
			got, err := c.client.SampleV1alpha1().Ipses().Get(ctx, ips.Name, metav1.GetOptions{})
			if err == nil {
				ips = got.DeepCopy()
				ips.Status = nowStatus
			} else {
				klog.Error(err, "failed to update ips", "name", ips.Name)
			}
			return updateErr
		})
	}
	return nil
}

func (c *ipsManager) CreateOrUpdateIpEndpoint(ctx context.Context, ipep *ipsv1alpha1.IpEndpoint) error {
	got, err := c.client.SampleV1alpha1().IpEndpoints(ipep.Namespace).Get(ctx, ipep.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := c.client.SampleV1alpha1().IpEndpoints(ipep.Namespace).Create(ctx, ipep, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}
	ipep.SetResourceVersion(got.GetResourceVersion())
	if _, err = c.client.SampleV1alpha1().IpEndpoints(ipep.Namespace).Update(ctx, ipep, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *ipsManager) NewIpEndpoint(ip string, pod *corev1.Pod, ips *ipsv1alpha1.Ips) (*ipsv1alpha1.IpEndpoint, error) {
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
