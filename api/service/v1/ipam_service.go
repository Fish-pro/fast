package v1

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	coreinformers "k8s.io/client-go/informers/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	ipamapiv1 "github.com/fast-io/fast/api/proto/v1"
	ipsversioned "github.com/fast-io/fast/pkg/generated/clientset/versioned"
	ipsinformers "github.com/fast-io/fast/pkg/generated/informers/externalversions/ips/v1alpha1"
	ipslisters "github.com/fast-io/fast/pkg/generated/listers/ips/v1alpha1"
	"github.com/fast-io/fast/pkg/util"
)

const IpsPodAnnotation = "fast.io/ips"

type IPAMService struct {
	ctx    context.Context
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
	client ipsversioned.Interface,
	podInformer coreinformers.PodInformer,
	ipsInformer ipsinformers.IpsInformer,
	ipepInformer ipsinformers.IpEndpointInformer) ipamapiv1.IpServiceServer {

	ipamSvc := &IPAMService{
		ctx:        ctx,
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
		return &ipamapiv1.IPAMResponse{Ip: *ipep.Status.IPs.IPv4}, nil
	}

	val, ok := pod.Annotations[IpsPodAnnotation]
	if ok {
		ips, err := s.ipsLister.Get(val)
		if err != nil {
			return nil, err
		}
		if *ips.Status.AllocatedIPCount >= *ips.Status.TotalIPCount {
			return nil, fmt.Errorf("ips %s/%s not enough ip addresses to allocate", ips.Namespace, ips.Name)
		}
	}

	return nil, nil
}

func (s *IPAMService) Release(context.Context, *ipamapiv1.IPAMRequest) (*ipamapiv1.IPAMResponse, error) {
	return nil, nil
}
