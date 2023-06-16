package ipsmanager

import (
	"context"
	"fmt"
	"net"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
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
	AllocateIP(ctx context.Context, pod *corev1.Pod) (*AllocateResult, error)
	ReleaseIP(ctx context.Context, namespace, name string) error
	NewIpEndpoint(ipsName string, pod *corev1.Pod, ip string) (*ipsv1alpha1.IpEndpoint, error)
	CreateIpEndpoint(ctx context.Context, ipep *ipsv1alpha1.IpEndpoint) error
}

type ipsManager struct {
	client ipsversioned.Interface
}

type AllocateResult struct {
	Namespace string
	Name      string
	IP        string
	IPsName   string
}

func NewIpsManager(client ipsversioned.Interface) IpsManager {
	return &ipsManager{client: client}
}

func (c *ipsManager) AllocateIP(ctx context.Context, pod *corev1.Pod) (*AllocateResult, error) {
	ipsName := getIpsNameByPod(pod)

	var res AllocateResult
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		ips, err := c.client.SampleV1alpha1().Ipses().Get(ctx, ipsName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if ips.Status.AllocatedIPCount >= ips.Status.TotalIPCount {
			return fmt.Errorf("ips %s/%s not enough ip addresses to allocate", ips.Namespace, ips.Name)
		}

		allIps := make([]net.IP, 0)
		for _, IP := range ips.Spec.IPs {
			allIps = append(allIps, util.ParseIPRange(IP)...)
		}

		excludeIps := make([]net.IP, 0)
		for k := range ips.Status.AllocatedIPs {
			excludeIps = append(excludeIps, net.ParseIP(k))
		}

		canAllocateIps := util.ExcludeIPs(allIps, excludeIps)
		if len(canAllocateIps) == 0 {
			return fmt.Errorf("ips %s/%s not enough ip addresses to allocate", ips.Namespace, ips.Name)
		}
		ip := canAllocateIps[0]

		if ips.Status.AllocatedIPs == nil {
			ips.Status.AllocatedIPs = make(map[string]ipsv1alpha1.AllocatedPod)
		}
		ips.Status.AllocatedIPs[ip.String()] = ipsv1alpha1.AllocatedPod{
			Pod:    fmt.Sprintf("%s/%s", pod.Namespace, pod.Name),
			PodUid: string(pod.UID),
		}

		if _, err := c.client.SampleV1alpha1().Ipses().UpdateStatus(ctx, ips, metav1.UpdateOptions{}); err != nil {
			return err
		}
		res = AllocateResult{Namespace: pod.Namespace, Name: pod.Name, IPsName: ipsName, IP: ip.String()}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to allocate IP from ips %s: %w", ipsName, err)
	}

	return &res, nil
}

func (c *ipsManager) ReleaseIP(ctx context.Context, namespace, name string) error {
	ipep, err := c.client.SampleV1alpha1().IpEndpoints(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	releaseIP := ipep.Status.IPs.IPv4
	ipsName := ipep.Status.IPs.IPv4Pool
	if len(releaseIP) == 0 || len(ipsName) == 0 {
		return fmt.Errorf("failed to get release IP(%s) or ips name(%s)", releaseIP, ipsName)
	}
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		ips, err := c.client.SampleV1alpha1().Ipses().Get(ctx, ipsName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if _, ok := ips.Status.AllocatedIPs[releaseIP]; !ok {
			return nil
		}
		delete(ips.Status.AllocatedIPs, releaseIP)

		if _, err := c.client.SampleV1alpha1().Ipses().UpdateStatus(ctx, ips, metav1.UpdateOptions{}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to release ip for ipe %s; %w", ipsName, err)
	}
	return c.removeIpEndpointFinalizer(ctx, ipep)
}

func (c *ipsManager) removeIpEndpointFinalizer(ctx context.Context, ipep *ipsv1alpha1.IpEndpoint) error {
	if controllerutil.ContainsFinalizer(ipep, IPsManagerFinalizer) {
		controllerutil.RemoveFinalizer(ipep, IPsManagerFinalizer)
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			got, err := c.client.SampleV1alpha1().IpEndpoints(ipep.Namespace).Get(ctx, ipep.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			ipep.SetResourceVersion(got.GetResourceVersion())
			_, err = c.client.SampleV1alpha1().IpEndpoints(ipep.Namespace).Update(ctx, ipep, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to remove finalizer for ip endpoint: %w", err)
		}
	}
	return nil
}

func (c *ipsManager) NewIpEndpoint(ipsName string, pod *corev1.Pod, ip string) (*ipsv1alpha1.IpEndpoint, error) {
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
				IPv4Pool: ipsName,
			},
		},
	}
	if err := controllerutil.SetOwnerReference(pod, ipep, scheme.Scheme); err != nil {
		return nil, err
	}
	return ipep, nil
}

func (c *ipsManager) CreateIpEndpoint(ctx context.Context, ipep *ipsv1alpha1.IpEndpoint) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var (
			got *ipsv1alpha1.IpEndpoint
			err error
		)
		got, err = c.client.SampleV1alpha1().IpEndpoints(ipep.Namespace).Get(ctx, ipep.Name, metav1.GetOptions{})
		if err != nil && apierrors.IsNotFound(err) {
			if got, err = c.client.SampleV1alpha1().IpEndpoints(ipep.Namespace).Create(ctx, ipep, metav1.CreateOptions{}); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		ipep.SetResourceVersion(got.GetResourceVersion())
		if _, err := c.client.SampleV1alpha1().IpEndpoints(ipep.Namespace).UpdateStatus(ctx, ipep, metav1.UpdateOptions{}); err != nil {
			return err
		}
		return nil
	})
}

func getIpsNameByPod(pod *corev1.Pod) string {
	ipsName := pod.Annotations[IpsPodAnnotation]
	if len(ipsName) == 0 {
		ipsName = DefaultIpsName
	}
	return ipsName
}
