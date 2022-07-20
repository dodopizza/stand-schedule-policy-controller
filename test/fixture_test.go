//go:build integration

package test

import (
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/controller"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
	"github.com/dodopizza/stand-schedule-policy-controller/pkg/clock"
)

var (
	_KubeConfigEnvVar = "TEST_KUBECONFIG_PATH"
	_Time             = time.Now()
)

type (
	fixture struct {
		cfg       *controller.Config
		kube      kubernetes.Interface
		azure     *azure
		clock     clock.FrozenClock
		logger    *zap.Logger
		interrupt chan struct{}
	}
	azure struct{}
)

func NewFixture(t *testing.T) *fixture {
	k, err := kubernetes.NewForTest(_KubeConfigEnvVar)
	if err != nil {
		t.Fatal(err)
	}

	l, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}

	return &fixture{
		cfg:       &controller.Config{},
		kube:      k,
		azure:     &azure{},
		clock:     clock.NewFrozenClock(_Time),
		logger:    l,
		interrupt: make(chan struct{}),
	}
}

func (f *fixture) CreateController() *controller.Controller {
	return controller.NewController(
		f.kube,
		f.azure,
		f.logger,
		f.cfg,
		f.clock,
	)
}
