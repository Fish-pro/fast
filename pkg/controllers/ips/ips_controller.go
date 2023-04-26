package ips

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	ipsv1alpha1 "github.com/fast-io/fast/pkg/apis/ips/v1alpha1"
	ipsversioned "github.com/fast-io/fast/pkg/generated/clientset/versioned"
	"github.com/fast-io/fast/pkg/generated/clientset/versioned/scheme"
	ipsinformers "github.com/fast-io/fast/pkg/generated/informers/externalversions/ips/v1alpha1"
	ipslisters "github.com/fast-io/fast/pkg/generated/listers/ips/v1alpha1"
	"github.com/fast-io/fast/pkg/util"
)

const (
	// maxRetries is the number of times ips will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// ips is going to be requeued:
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries     = 15
	ControllerName = "ips-controller"
)

var defaultValue = uint32(1)

// Controller define the option of controller
type Controller struct {
	kubeClient kubernetes.Interface
	client     ipsversioned.Interface

	// lister define the cache object
	lister ipslisters.IpsLister

	// synced define the sync for relist
	ipsSynced cache.InformerSynced

	// Ips that need to be synced
	queue workqueue.RateLimitingInterface

	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder
}

func (c *Controller) Name() string {
	return ControllerName
}

// NewController return a controller and add event handler
func NewController(
	ctx context.Context,
	kubeClient kubernetes.Interface,
	client ipsversioned.Interface,
	informer ipsinformers.IpsInformer) (*Controller, error) {
	logger := klog.FromContext(ctx)

	logger.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	controller := &Controller{
		client:           client,
		kubeClient:       kubeClient,
		lister:           informer.Lister(),
		ipsSynced:        informer.Informer().HasSynced,
		eventBroadcaster: eventBroadcaster,
		eventRecorder:    eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: ControllerName}),
		queue: workqueue.NewRateLimitingQueueWithConfig(workqueue.DefaultControllerRateLimiter(), workqueue.RateLimitingQueueConfig{
			Name: ControllerName,
		}),
	}

	logger.Info("Setting up event handlers")
	_, err := informer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.enqueue(logger, obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			controller.enqueue(logger, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			controller.enqueue(logger, obj)
		},
	}, time.Second*30)
	if err != nil {
		logger.Error(err, "Failed to setting up event handlers")
		return nil, err
	}

	return controller, nil
}

// Run worker and sync the queue obj to self logic
func (c *Controller) Run(ctx context.Context) {
	defer utilruntime.HandleCrash()

	// Start events processing pipeline.
	c.eventBroadcaster.StartStructuredLogging(0)
	c.eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: c.kubeClient.CoreV1().Events(metav1.NamespaceAll)})
	defer c.eventBroadcaster.Shutdown()

	defer c.queue.ShutDown()

	logger := klog.FromContext(ctx)
	// Start the informer factories to begin populating the informer caches
	logger.Info("Starting controller", "controller", ControllerName)
	defer logger.Info("Shutting down controller", "controller", ControllerName)

	// Wait for the caches to be synced before starting worker
	logger.Info("Waiting for informer caches to sync")
	if !cache.WaitForCacheSync(ctx.Done(), c.ipsSynced) {
		logger.Error(fmt.Errorf("failed to sync informer"), "Informer caches to sync bad")
		return
	}

	logger.Info("Starting worker")
	go wait.UntilWithContext(ctx, c.runWorker, time.Second)

	<-ctx.Done()
}

// runWorker wait obj by queue
func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *Controller) processNextWorkItem(ctx context.Context) bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.syncHandler(ctx, key.(string))
	c.handleErr(ctx, err, key)

	return true
}

func (c *Controller) handleErr(ctx context.Context, err error, key interface{}) {
	logger := klog.FromContext(ctx)
	if err == nil || apierrors.HasStatusCause(err, v1.NamespaceTerminatingCause) {
		c.queue.Forget(key)
		return
	}
	ns, name, keyErr := cache.SplitMetaNamespaceKey(key.(string))
	if keyErr != nil {
		logger.Error(err, "Failed to split meta namespace cache key", "cacheKey", key)
	}

	if c.queue.NumRequeues(key) < maxRetries {
		logger.V(2).Info("Error syncing ips", "ips", klog.KRef(ns, name), "err", err)
		c.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	logger.V(2).Info("Dropping ips out of the queue", "ips", klog.KRef(ns, name), "err", err)
	c.queue.Forget(key)
}

// syncHandler sync the ips object
func (c *Controller) syncHandler(ctx context.Context, key string) error {
	logger := klog.FromContext(ctx)

	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Error(err, "Failed to split meta namespace cache key", "cacheKey", key)
		return err
	}

	startTime := time.Now()
	logger.V(4).Info("Started syncing ips", "ips", name, "startTime", startTime)
	defer func() {
		logger.V(4).Info("Finished syncing ips", "ips", name, "duration", time.Since(startTime))
	}()

	obj, err := c.lister.Get(name)
	if apierrors.IsNotFound(err) {
		logger.Info("Ips not found", "ips", name)
		return nil
	} else if err != nil {
		logger.Error(err, "Failed to get ips", "ips", name)
		return err
	}
	ips := obj.DeepCopy()

	if !ips.DeletionTimestamp.IsZero() {
		return nil
	}

	count := 0
	for _, ip := range ips.Spec.IPs {
		count += len(util.ParseIPRange(ip))
	}
	nowStatus := ipsv1alpha1.IpsStatus{}
	nowStatus.TotalIPCount = count
	nowStatus.AllocatedIPCount = len(nowStatus.AllocatedIPs)

	return c.updateIpsStatusIfNeed(ctx, ips, nowStatus)
}

// updateIpsStatusIfNeed update status if we need
func (c *Controller) updateIpsStatusIfNeed(ctx context.Context, ips *ipsv1alpha1.Ips, status ipsv1alpha1.IpsStatus) error {
	logger := klog.FromContext(ctx)
	if !equality.Semantic.DeepEqual(ips.Status, status) {
		ips.Status = status
		return retry.RetryOnConflict(retry.DefaultRetry, func() error {
			_, updateErr := c.client.SampleV1alpha1().Ipses().UpdateStatus(ctx, ips, metav1.UpdateOptions{})
			if updateErr == nil {
				return nil
			}
			got, err := c.client.SampleV1alpha1().Ipses().Get(ctx, ips.Name, metav1.GetOptions{})
			if err == nil {
				ips = got.DeepCopy()
				ips.Status = status
			} else {
				logger.Error(err, "Failed to get ips", "ips", ips.Name)
			}
			return fmt.Errorf("failed to update ips %s status: %w", ips.Name, updateErr)
		})
	}
	return nil
}

// If nodeName is used, it is not queued if there is no match
func (c *Controller) enqueue(logger klog.Logger, obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %s: %w", key, err))
		return
	}

	c.queue.Add(key)
}
