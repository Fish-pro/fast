package v1

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	ipamapiv1 "github.com/fast-io/fast/pkg/api/proto/v1"
	ipsversioned "github.com/fast-io/fast/pkg/generated/clientset/versioned"
	"github.com/fast-io/fast/pkg/ipsmanager"
	"github.com/fast-io/fast/pkg/util"
)

type IPAMService struct {
	client     ipsversioned.Interface
	kubeClient kubernetes.Interface
	logger     *zap.Logger

	ipsManager ipsmanager.IpsManager

	ipamapiv1.UnimplementedIpServiceServer
}

func NewIPAMService(
	ctx context.Context,
	kubeClient kubernetes.Interface,
	client ipsversioned.Interface,
	logger *zap.Logger) ipamapiv1.IpServiceServer {
	return &IPAMService{
		client:     client,
		kubeClient: kubeClient,
		logger:     logger,
		ipsManager: ipsmanager.NewIpsManager(client),
	}
}

func (s *IPAMService) Health(context.Context, *ipamapiv1.HealthRequest) (*ipamapiv1.HealthResponse, error) {
	return &ipamapiv1.HealthResponse{Health: ipamapiv1.HealthyType_Healthy}, nil
}

func (s *IPAMService) Allocate(ctx context.Context, req *ipamapiv1.AllocateRequest) (*ipamapiv1.AllocateResponse, error) {
	if len(req.Namespace) == 0 || len(req.Name) == 0 {
		return nil, fmt.Errorf("namespace or name can not be none")
	}
	s.logger.Info("allocate ip", zap.String("namespace", req.Namespace), zap.String("name", req.Name))

	pod, err := s.kubeClient.CoreV1().Pods(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		s.logger.Error("get pod from lister error", zap.Error(err))
		return nil, err
	}
	if !util.IsPodAlive(pod) {
		return nil, fmt.Errorf("pod is not alive")
	}

	ipep, err := s.client.SampleV1alpha1().IpEndpoints(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		s.logger.Error("get ip endpoint error", zap.Error(err))
		return nil, err
	}
	if ipep != nil && len(ipep.Status.IPs.IPv4) > 0 {
		s.logger.Info("ip endpoint exist", zap.String("ip", ipep.Status.IPs.IPv4))
		return &ipamapiv1.AllocateResponse{Ip: ipep.Status.IPs.IPv4}, nil
	}

	allocateResult, err := s.ipsManager.AllocateIP(ctx, pod)
	if err != nil {
		s.logger.Error("failed to allocate ip", zap.Error(err))
		return nil, err
	}

	ipep, err = s.ipsManager.NewIpEndpoint(allocateResult.IPsName, pod, allocateResult.IP)
	if err != nil {
		s.logger.Error("failed to new ip endpoint")
		return nil, err
	}

	if err := s.ipsManager.CreateIpEndpoint(ctx, ipep); err != nil {
		s.logger.Error("failed to create or update ip endpoint", zap.Error(err))
		return nil, err
	}
	s.logger.Info("allocate ip successfully", zap.String("ip", allocateResult.IP))

	return &ipamapiv1.AllocateResponse{Ip: allocateResult.IP}, nil
}

func (s *IPAMService) Release(ctx context.Context, req *ipamapiv1.AllocateRequest) (*ipamapiv1.ReleaseResponse, error) {
	if len(req.Namespace) == 0 || len(req.Name) == 0 {
		return nil, fmt.Errorf("namespace or name can not be none")
	}
	s.logger.Info("release ip", zap.String("namespace", req.Namespace), zap.String("name", req.Name))

	return &ipamapiv1.ReleaseResponse{}, s.ipsManager.ReleaseIP(ctx, req.Namespace, req.Name)
}
