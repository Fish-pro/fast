package v1

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	ipamapiv1 "github.com/fast-io/fast/api/proto/v1"
	ipsv1alpha1 "github.com/fast-io/fast/pkg/apis/ips/v1alpha1"
	ipsversioned "github.com/fast-io/fast/pkg/generated/clientset/versioned"
	ipsinformers "github.com/fast-io/fast/pkg/generated/informers/externalversions/ips/v1alpha1"
	ipslisters "github.com/fast-io/fast/pkg/generated/listers/ips/v1alpha1"
	"github.com/fast-io/fast/pkg/scheme"
	"github.com/fast-io/fast/pkg/util"
)

const (
	IpsPodAnnotation = "fast.io/ips"
	DefaultIpsName   = "default-ips"
)

var logger logr.Logger

type IPAMService struct {
	client ipsversioned.Interface

	podLister  corelisters.PodLister
	ipsLister  ipslisters.IpsLister
	ipepLister ipslisters.IpEndpointLister

	podSynced  cache.InformerSynced
	ipsSynced  cache.InformerSynced
	ipepSynced cache.InformerSynced

	ipamapiv1.UnimplementedIpServiceServer
}

func NewIPAMService(
	ctx context.Context,
	l logr.Logger,
	client ipsversioned.Interface,
	podInformer coreinformers.PodInformer,
	ipsInformer ipsinformers.IpsInformer,
	ipepInformer ipsinformers.IpEndpointInformer) ipamapiv1.IpServiceServer {

	logger = l
	ipamSvc := &IPAMService{
		client:     client,
		podLister:  podInformer.Lister(),
		ipsLister:  ipsInformer.Lister(),
		ipepLister: ipepInformer.Lister(),
		podSynced:  podInformer.Informer().HasSynced,
		ipsSynced:  ipsInformer.Informer().HasSynced,
		ipepSynced: ipepInformer.Informer().HasSynced,
	}

	go func(ctx context.Context) {
		if !cache.WaitForCacheSync(ctx.Done(), ipamSvc.podSynced, ipamSvc.ipsSynced, ipamSvc.ipepSynced) {
			logger.Error(fmt.Errorf("failed to sync informer"), "sync informer error")
			return
		}
	}(ctx)

	return ipamSvc
}

func (s *IPAMService) Start(ctx context.Context) {
	if !cache.WaitForCacheSync(ctx.Done(), s.podSynced, s.ipsSynced) {
		return
	}
}

func (s *IPAMService) Health(context.Context, *ipamapiv1.HealthRequest) (*ipamapiv1.HealthResponse, error) {
	return &ipamapiv1.HealthResponse{Msg: "ok"}, nil
}

func (s *IPAMService) Allocate(ctx context.Context, req *ipamapiv1.IPAMRequest) (*ipamapiv1.IPAMResponse, error) {
	pod, err := s.podLister.Pods(req.Namespace).Get(req.Name)
	if err != nil {
		return nil, err
	}
	if !util.IsPodAlive(pod) {
		return nil, fmt.Errorf("pod is not alive")
	}

	ipep, err := s.ipepLister.IpEndpoints(req.Namespace).Get(req.Name)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	if ipep != nil {
		return &ipamapiv1.IPAMResponse{Ip: ipep.Status.IPs.IPv4}, nil
	}

	if val, ok := pod.Annotations[IpsPodAnnotation]; ok {
		ipsObj, err := s.ipsLister.Get(val)
		if err != nil {
			return nil, err
		}
		ips := ipsObj.DeepCopy()

		ip, err := AllocateIp(s.client, ips, pod)
		if err != nil {
			return nil, err
		}

		ipep, err := newIpEndpoint(ip, pod, ips)
		if err != nil {
			return nil, err
		}

		if err := s.createOrUpdateIpEndpoint(ctx, ipep); err != nil {
			return nil, err
		}

		return &ipamapiv1.IPAMResponse{Ip: ip}, nil
	}

	defaultIps, err := s.ipsLister.Get(DefaultIpsName)
	if err != nil {
		return nil, fmt.Errorf("get default ips error: %v", err)
	}
	ip, err := AllocateIp(s.client, defaultIps, pod)
	if err != nil {
		return nil, err
	}

	return &ipamapiv1.IPAMResponse{Ip: ip}, nil
}

func (s *IPAMService) Release(context.Context, *ipamapiv1.IPAMRequest) (*ipamapiv1.IPAMResponse, error) {
	return nil, nil
}

func updateIpsStatusIfNeed(ctx context.Context, client ipsversioned.Interface, ips *ipsv1alpha1.Ips, nowStatus ipsv1alpha1.IpsStatus) error {
	if !equality.Semantic.DeepEqual(ips.Status, nowStatus) {
		ips.Status = nowStatus
		return retry.RetryOnConflict(retry.DefaultRetry, func() error {
			_, updateErr := client.SampleV1alpha1().Ipses().UpdateStatus(ctx, ips, metav1.UpdateOptions{})
			if updateErr == nil {
				return nil
			}
			got, err := client.SampleV1alpha1().Ipses().Get(ctx, ips.Name, metav1.GetOptions{})
			if err == nil {
				ips = got.DeepCopy()
				ips.Status = nowStatus
			} else {
				logger.Error(err, "failed to update ips", "name", ips.Name)
			}
			return updateErr
		})
	}
	return nil
}

func AllocateIp(client ipsversioned.Interface, ips *ipsv1alpha1.Ips, pod *corev1.Pod) (string, error) {
	if *ips.Status.AllocatedIPCount >= *ips.Status.TotalIPCount {
		return "", fmt.Errorf("ips %s/%s not enough ip addresses to allocate", ips.Namespace, ips.Name)
	}

	var excludeIps []string
	for k := range ips.Status.AllocatedIPs {
		excludeIps = append(excludeIps, k)
	}
	canAllocateIps := util.ExcludeIPs(ips.Spec.IPs, excludeIps)
	if len(canAllocateIps) == 0 {
		return "", fmt.Errorf("ips %s/%s not enough ip addresses to allocate", ips.Namespace, ips.Name)
	}
	ip := canAllocateIps[0].String()

	nowStatus := ips.Status
	if nowStatus.AllocatedIPs == nil {
		nowStatus.AllocatedIPs = make(map[string]ipsv1alpha1.AllocatedPod)
	}
	nowStatus.AllocatedIPs[ip] = ipsv1alpha1.AllocatedPod{
		Pod:    fmt.Sprintf("%s/%s", pod.Namespace, pod.Name),
		PodUid: string(pod.UID),
	}

	if err := updateIpsStatusIfNeed(context.TODO(), client, ips, nowStatus); err != nil {
		return "", err
	}

	return ip, nil
}

func newIpEndpoint(ip string, pod *corev1.Pod, ips *ipsv1alpha1.Ips) (*ipsv1alpha1.IpEndpoint, error) {
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

func (s *IPAMService) createOrUpdateIpEndpoint(ctx context.Context, ipep *ipsv1alpha1.IpEndpoint) error {
	got, err := s.client.SampleV1alpha1().IpEndpoints(ipep.Namespace).Get(ctx, ipep.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := s.client.SampleV1alpha1().IpEndpoints(ipep.Namespace).Create(ctx, ipep, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}
	ipep.SetResourceVersion(got.GetResourceVersion())
	if _, err = s.client.SampleV1alpha1().IpEndpoints(ipep.Namespace).Update(ctx, ipep, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}
