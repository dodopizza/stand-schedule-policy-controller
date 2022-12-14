package app

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/utils/clock"

	"github.com/dodopizza/stand-schedule-policy-controller/config"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/azure"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/controller"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/http"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
	"github.com/dodopizza/stand-schedule-policy-controller/pkg/httpserver"
)

type (
	App struct {
		logger     *zap.Logger
		kube       kubernetes.Interface
		az         azure.Interface
		server     *httpserver.Server
		controller *controller.Controller
		interrupt  chan struct{}
	}
)

func New(cfg *config.Config, l *zap.Logger) (*App, error) {
	k, err := kubernetes.NewForAccessType(cfg.Kube.AccessType)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize kubernetes client")
	}

	az, err := azure.NewForConfig(&cfg.Azure)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize azure client")
	}

	hs := httpserver.New(http.NewRouter(), httpserver.Port(cfg.Http.Port))
	c := controller.NewController(&cfg.Controller, l, clock.RealClock{}, k, az)

	return &App{
		logger:     l,
		kube:       k,
		az:         az,
		server:     hs,
		controller: c,
	}, nil
}

func Run(l *zap.Logger, cfg *config.Config) {
	app, err := New(cfg, l)
	if err != nil {
		l.Fatal("Failed to initialize app", zap.Error(err))
	}

	app.logger.Info("Application starting")
	app.SetupSignalHandlers()
	app.server.Start()
	app.controller.Start(app.interrupt)
	app.logger.Info("Application started")

	select {
	case <-app.interrupt:
		app.logger.Info("Interruption received")
	case err = <-app.controller.Notify():
		app.logger.Error("Controller failure", zap.Error(err))
	case err = <-app.server.Notify():
		app.logger.Error("Http server failure", zap.Error(err))
	}

	app.logger.Info("Application stopping")
	if err := app.server.Shutdown(); err != nil {
		app.logger.Error("Http server shutdown failure", zap.Error(err))
	}
	app.logger.Info("Application stopped")
}

func (app *App) SetupSignalHandlers() {
	app.interrupt = make(chan struct{})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		app.logger.Info("Signal received", zap.Stringer("signal", <-signals))
		close(app.interrupt)
	}()
}
