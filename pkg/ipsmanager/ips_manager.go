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
	IpsPodAnnotation    = "fast.io/ips"
	DefaultIpsName      = "default-ips"
	IPsManagerFinalizer = "fast.io/ips-manager"
)

type IpsManager interface {
	AllocateIP(ctx context.Context, pod *corev1.Pod) (*ipsv1alpha1.Ips, net.IP, error)
	ReleaseIP(ctx context.Context, namespace, name string) error
	UpdateIpsStatus(ctx context.Context, ips *ipsv1alpha1.Ips, nowStatus ipsv1alpha1.IpsStatus) error
	NewIpEndpoint(ctx context.Context, ips *ipsv1alpha1.Ips, pod *corev1.Pod, ip string) (*ipsv1alpha1.IpEndpoint, error)
	CreateIpEndpoint(ctx context.Context, ipep *ipsv1alpha1.IpEndpoint) error
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
	obj, err := c.client.SampleV1alpha1().Ipses().Get(ctx, ipsName, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	ips := obj.DeepCopy()

	if ips.Status.AllocatedIPCount >= ips.Status.TotalIPCount {
		return nil, nil, fmt.Errorf("ips %s/%s not enough ip addresses to allocate", ips.Namespace, ips.Name)
	}

	allIps := make([]net.IP, 0)
	for _, IP := range ips.Spec.IPs {
		allIps = append(allIps, util.ParseIPRange(IP)...)
	}

	excludeIps := make([]net.IP, 0)
	for k, v := range ips.Status.AllocatedIPs {
		if v.PodUid == string(pod.UID) {
			return ips, net.ParseIP(k), nil
		}
		excludeIps = append(excludeIps, net.ParseIP(k))
	}

	canAllocateIps := util.ExcludeIPs(allIps, excludeIps)
	if len(canAllocateIps) == 0 {
		return nil, nil, fmt.Errorf("ips %s/%s not enough ip addresses to allocate", ips.Namespace, ips.Name)
	}
	ip := canAllocateIps[0]

	if ips.Status.AllocatedIPs == nil {
		ips.Status.AllocatedIPs = make(map[string]ipsv1alpha1.AllocatedPod)
	}
	ips.Status.AllocatedIPs[ip.String()] = ipsv1alpha1.AllocatedPod{
		Pod:    fmt.Sprintf("%s/%s", pod.Namespace, pod.Name),
		PodUid: string(pod.UID),
	}

	if err := c.UpdateIpsStatus(ctx, obj, ips.Status); err != nil {
		return nil, nil, err
	}
	return ips, ip, nil
}

func (c *ipsManager) ReleaseIP(ctx context.Context, namespace, name string) error {
	ipep, err := c.client.SampleV1alpha1().IpEndpoints(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if ipep == nil {
		return nil
	}

	ipsList, err := c.client.SampleV1alpha1().Ipses().List(ctx, metav1.ListOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	releaseIP := ipep.Status.IPs.IPv4
	for _, obj := range ipsList.Items {
		ips := obj.DeepCopy()
		if ips.Status.AllocatedIPs == nil || len(ips.Status.AllocatedIPs) == 0 {
			continue
		}
		if _, ok := ips.Status.AllocatedIPs[releaseIP]; !ok {
			continue
		}
		delete(ips.Status.AllocatedIPs, releaseIP)
		if err := c.UpdateIpsStatus(ctx, &obj, ips.Status); err != nil {
			return err
		}
	}

	if controllerutil.ContainsFinalizer(ipep, IPsManagerFinalizer) {
		controllerutil.RemoveFinalizer(ipep, IPsManagerFinalizer)
		if _, err := c.client.SampleV1alpha1().IpEndpoints(namespace).Update(ctx, ipep, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}
	return nil
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

func (c *ipsManager) NewIpEndpoint(ctx context.Context, ips *ipsv1alpha1.Ips, pod *corev1.Pod, ip string) (*ipsv1alpha1.IpEndpoint, error) {
	ipep := &ipsv1alpha1.IpEndpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Finalizers: []string{IPsManagerFinalizer},
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

func (c *ipsManager) CreateIpEndpoint(ctx context.Context, ipep *ipsv1alpha1.IpEndpoint) error {
	created, err := c.client.SampleV1alpha1().IpEndpoints(ipep.Namespace).Create(ctx, ipep, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	ipep.SetResourceVersion(created.GetResourceVersion())
	if _, err = c.client.SampleV1alpha1().IpEndpoints(ipep.Namespace).UpdateStatus(ctx, ipep, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}
