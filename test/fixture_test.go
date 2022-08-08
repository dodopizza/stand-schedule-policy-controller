//go:build integration

package test

import (
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
	clock "k8s.io/utils/clock/testing"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/controller"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
)

var (
	_WorkingDir, _      = os.Getwd()
	_DefaultKubeCfgPath = _WorkingDir + "/../bin/kubeconfig.yaml"
	_KubeCfgEnvVar      = "TEST_KUBECONFIG_PATH"
	_Time               = time.Now()
)

type (
	fixture struct {
		cfg       *controller.Config
		kube      kubernetes.Interface
		azure     *azureFixture
		clock     *clock.FakeClock
		interrupt chan struct{}
		t         *testing.T
		cleanup   *fixtureCleanup
	}
)

func NewFixture(t *testing.T) *fixture {
	cfgPath := os.Getenv(_KubeCfgEnvVar)
	if cfgPath == "" {
		t.Setenv(_KubeCfgEnvVar, _DefaultKubeCfgPath)
	}

	k, err := kubernetes.NewForTest(_KubeCfgEnvVar)
	if err != nil {
		t.Fatal(err)
	}

	cleanup := NewFixtureCleanup(t, k)

	return &fixture{
		cfg: &controller.Config{
			ObjectsResyncSeconds:  10,
			PoliciesResyncSeconds: 10,
			ReconcilerThreadiness: 1,
			ExecutorThreadiness:   1,
			WorkerQueueRetries:    5,
		},
		kube:      k,
		azure:     &azureFixture{},
		clock:     clock.NewFakeClock(_Time),
		interrupt: cleanup.interrupt,
		t:         t,
		cleanup:   cleanup,
	}
}

func (f *fixture) WithClockTime(ts time.Time) *fixture {
	_Time = ts
	f.clock = clock.NewFakeClock(_Time)
	return f
}

func (f *fixture) WithoutCleanup() *fixture {
	f.cleanup.cleanup = false
	return f
}

func (f *fixture) CreateController() *controller.Controller {
	l, err := zap.NewDevelopment()
	if err != nil {
		f.t.Fatal(err)
	}

	f.cleanup.controller = controller.NewController(f.cfg, l, f.clock, f.kube, f.azure)

	return f.cleanup.controller
}

func (f *fixture) IncreaseTime(d time.Duration) {
	nextTime := _Time.Add(d)
	f.t.Logf("Increase controller time from %s to %s", _Time, nextTime)
	f.clock.SetTime(nextTime)
	_Time = nextTime
}
