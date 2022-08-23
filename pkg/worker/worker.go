package worker

import (
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/clock"
)

type (
	ReconcileFunc func(key interface{}) error
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

	go w.shutdown(interrupt)
}

func (w *Worker) Enqueue(item interface{}) {
	w.logger.Debug("Enqueue key", zap.Any("worker_key", item))
	w.queue.Add(item)
}

func (w *Worker) EnqueueAfter(item interface{}, duration time.Duration) {
	w.logger.Debug("Enqueue key deferred", zap.Any("worker_key", item), zap.Stringer("worker_key_after", duration))
	w.queue.AddAfter(item, duration)
}

func (w *Worker) shutdown(interrupt <-chan struct{}) {
	<-interrupt
	w.queue.ShutDown()
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

	err := w.reconcile(key)
	if err == nil {
		w.rateLimiter.Forget(key)
		return true
	}

	w.logger.Info("Failed to process key", zap.Any("worker_key", key), zap.Error(err))
	runtime.HandleError(err)

	if w.rateLimiter.NumRequeues(key) < w.config.Retries {
		w.logger.Info("Requeue key", zap.Any("worker_key", key), zap.Error(err))
		w.queue.AddAfter(key, w.rateLimiter.When(key))
		return true
	}

	w.logger.Info("Forget failed key", zap.Any("worker_key", key), zap.Error(err))
	w.rateLimiter.Forget(key)
	return true
}
