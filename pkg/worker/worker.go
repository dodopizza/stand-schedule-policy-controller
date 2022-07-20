package worker

import (
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/clock"
)

type (
	ReconcileFunc func(key string) error
	Config        struct {
		Name        string
		Retries     int
		Threadiness int
	}
	Worker struct {
		logger      *zap.Logger
		config      *Config
		queue       workqueue.DelayingInterface
		reconcile   ReconcileFunc
		rateLimiter workqueue.RateLimiter
	}
)

func New(
	cfg *Config,
	l *zap.Logger,
	clock clock.WithTicker,
	reconcile ReconcileFunc,
) *Worker {
	return &Worker{
		logger:      l.Named("worker"),
		config:      cfg,
		queue:       workqueue.NewDelayingQueueWithCustomClock(clock, cfg.Name),
		reconcile:   reconcile,
		rateLimiter: workqueue.DefaultControllerRateLimiter(),
	}
}

func (w *Worker) Start(interrupt <-chan struct{}) {
	defer runtime.HandleCrash()

	for i := 0; i < w.config.Threadiness; i++ {
		go wait.Until(w.process, time.Second, interrupt)
	}
}

func (w *Worker) Shutdown() {
	w.queue.ShutDown()
}

func (w *Worker) Enqueue(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		return
	}
	w.queue.Add(key)
}

func (w *Worker) GetQueue() workqueue.DelayingInterface {
	return w.queue
}

func (w *Worker) process() {
	for w.next() {
	}
}

func (w *Worker) next() bool {
	key, quit := w.queue.Get()
	if quit {
		return false
	}

	defer w.queue.Done(key)

	err := w.reconcile(key.(string))
	if err == nil {
		w.rateLimiter.Forget(key)
		return true
	}

	w.logger.Info("Failed to process key with error", zap.Any("key", key), zap.Error(err))
	runtime.HandleError(err)

	if w.rateLimiter.NumRequeues(key) < w.config.Retries {
		w.logger.Info("Requeue key with error", zap.Any("key", key), zap.Error(err))
		w.queue.AddAfter(key, w.rateLimiter.When(key))
		return true
	}

	w.logger.Info("Forget failed key with error", zap.Any("key", key), zap.Error(err))
	w.rateLimiter.Forget(key)
	return true
}
