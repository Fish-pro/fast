package gc

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	ipsversioned "github.com/fast-io/fast/pkg/generated/clientset/versioned"
	"github.com/fast-io/fast/pkg/generated/clientset/versioned/scheme"
	"github.com/fast-io/fast/pkg/ipsmanager"
)

const (
	// maxRetries is the number of times pod will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// pod is going to be requeued:
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries     = 15
	ControllerName = "gc-manager"
)

var defaultValue = uint32(1)

// Controller define the option of controller
type Controller struct {
	kubeClient kubernetes.Interface
	client     ipsversioned.Interface

	// podLister define the cache pod
	podLister corelisters.PodLister
	// synced define the sync for relist
	podSynced cache.InformerSynced

	ipsManager ipsmanager.IpsManager

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
	podInformer coreinformers.PodInformer) (*Controller, error) {
	logger := klog.FromContext(ctx)

	logger.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	controller := &Controller{
		client:           client,
		kubeClient:       kubeClient,
		podLister:        podInformer.Lister(),
		podSynced:        podInformer.Informer().HasSynced,
		ipsManager:       ipsmanager.NewIpsManager(client),
		eventBroadcaster: eventBroadcaster,
		eventRecorder:    eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: ControllerName}),
		queue: workqueue.NewRateLimitingQueueWithConfig(workqueue.DefaultControllerRateLimiter(), workqueue.RateLimitingQueueConfig{
			Name: ControllerName,
		}),
	}

	logger.Info("Setting up event handlers")
	_, err := podInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
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
	if !cache.WaitForCacheSync(ctx.Done(), c.podSynced) {
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
		logger.V(2).Info("Error syncing pod", "pod", klog.KRef(ns, name), "err", err)
		c.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	logger.V(2).Info("Dropping pod out of the queue", "pod", klog.KRef(ns, name), "err", err)
	c.queue.Forget(key)
}

// syncHandler sync the pod object
func (c *Controller) syncHandler(ctx context.Context, key string) error {
	logger := klog.FromContext(ctx)

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Error(err, "Failed to split meta namespace cache key", "cacheKey", key)
		return err
	}

	startTime := time.Now()
	logger.V(4).Info("Started syncing pod", "pod", name, "startTime", startTime)
	defer func() {
		logger.V(4).Info("Finished syncing pod", "pod", name, "duration", time.Since(startTime))
	}()

	obj, err := c.podLister.Pods(ns).Get(name)
	if apierrors.IsNotFound(err) {
		logger.Info("Ips not found", "pod", name)
		return nil
	} else if err != nil {
		logger.Error(err, "Failed to get pod", "pod", name)
		return err
	}
	pod := obj.DeepCopy()

	return c.ipsManager.ReleaseIP(ctx, pod)
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
