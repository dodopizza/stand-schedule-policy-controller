package executor

import (
	"context"

	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/azure"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

type (
	Executor struct {
		logger *zap.Logger
		azure  azure.Interface
		kube   kubernetes.Interface
		lister *kubernetes.ListerGroup
	}
)

func New(l *zap.Logger, az azure.Interface, k kubernetes.Interface, lister *kubernetes.ListerGroup) *Executor {
	return &Executor{
		logger: l.Named("executor"),
		azure:  az,
		kube:   k,
		lister: lister,
	}
}

func (ex *Executor) ExecuteShutdown(ctx context.Context, policy *apis.StandSchedulePolicy) error {
	return multierr.Combine(
		ex.executeShutdownKube(ctx, policy),
		ex.executeShutdownAzure(ctx, policy.Spec.Resources.Azure),
	)
}

func (ex *Executor) ExecuteStartup(ctx context.Context, policy *apis.StandSchedulePolicy) error {
	return multierr.Combine(
		ex.executeStartupAzure(ctx, policy.Spec.Resources.Azure),
		ex.executeStartupKube(ctx, policy),
	)
}
