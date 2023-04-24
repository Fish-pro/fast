package v1

import (
	"context"
	"fmt"

	"github.com/fast-io/fast/pkg/ipsmanager"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	ipamapiv1 "github.com/fast-io/fast/api/proto/v1"
	ipsversioned "github.com/fast-io/fast/pkg/generated/clientset/versioned"
	"github.com/fast-io/fast/pkg/util"
)

type IPAMService struct {
	client ipsversioned.Interface

	podLister corelisters.PodLister
	podSynced cache.InformerSynced

	ipsManager ipsmanager.IpsManager

	ipamapiv1.UnimplementedIpServiceServer
}

func NewIPAMService(
	ctx context.Context,
	client ipsversioned.Interface,
	podInformer coreinformers.PodInformer) ipamapiv1.IpServiceServer {

	ipamSvc := &IPAMService{
		client:     client,
		podLister:  podInformer.Lister(),
		podSynced:  podInformer.Informer().HasSynced,
		ipsManager: ipsmanager.NewIpsManager(client),
	}

	go func(ctx context.Context) {
		if !cache.WaitForCacheSync(ctx.Done(), ipamSvc.podSynced) {
			klog.Error(fmt.Errorf("failed to sync informer"), "sync informer error")
			return
		}
	}(ctx)

	return ipamSvc
}

func (s *IPAMService) Health(context.Context, *ipamapiv1.HealthRequest) (*ipamapiv1.HealthResponse, error) {
	return &ipamapiv1.HealthResponse{Msg: util.HealthyOk}, nil
}

func (s *IPAMService) Allocate(ctx context.Context, req *ipamapiv1.AllocateRequest) (*ipamapiv1.AllocateResponse, error) {
	pod, err := s.podLister.Pods(req.Namespace).Get(req.Name)
	if err != nil {
		return nil, err
	}
	if !util.IsPodAlive(pod) {
		return nil, fmt.Errorf("pod is not alive")
	}

	endpoint, err := s.client.SampleV1alpha1().IpEndpoints(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	if endpoint != nil {
		return &ipamapiv1.AllocateResponse{Ip: endpoint.Status.IPs.IPv4}, nil
	}

	ips, ip, err := s.ipsManager.AllocateIP(ctx, pod)
	if err != nil {
		return nil, err
	}

	ipep, err := util.NewIpEndpoint(ip.String(), pod, ips)
	if err != nil {
		return nil, err
	}

	if err := s.ipsManager.CreateOrUpdateIpEndpoint(ctx, ipep); err != nil {
		return nil, err
	}

	return &ipamapiv1.AllocateResponse{Ip: ip.String()}, nil
}

func (s *IPAMService) Release(ctx context.Context, req *ipamapiv1.AllocateRequest) (*ipamapiv1.ReleaseResponse, error) {
	pod, err := s.podLister.Pods(req.Namespace).Get(req.Name)
	if err != nil {
		return nil, err
	}

	return nil, s.ipsManager.ReleaseIP(ctx, pod)
}

func (s *IPAMService) GetGateway(ctx context.Context, req *ipamapiv1.GatewayRequest) (*ipamapiv1.GatewayResponse, error) {
	gwIP, err := s.ipsManager.GetGateway(ctx, req.Node)
	if err != nil {
		return nil, err
	}
	return &ipamapiv1.GatewayResponse{Gateway: gwIP.String()}, nil
}
