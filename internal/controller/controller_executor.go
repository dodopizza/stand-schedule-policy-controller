package controller

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"

	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

// todo: check workitem.fireat and schedule.fireat
// todo: validation webhook
// todo: save state smth / leader election

type (
	WorkItem struct {
		policyName   string
		scheduleType apis.ConditionScheduleType
		fireAt       time.Time
	}
)

const (
	_ExecutionTimeout = time.Minute * 15
	_DeadlineTimeout  = time.Minute * 31
)

func (w *WorkItem) String() string {
	return fmt.Sprintf("%s/%s at %s", w.policyName, w.scheduleType, w.fireAt)
}

func (w *WorkItem) deadline() time.Time {
	return w.fireAt.Add(_DeadlineTimeout)
}

func (c *Controller) execute(i interface{}) error {
	now := c.clock.Now()
	item := i.(WorkItem)

	if now.Before(item.fireAt) {
		c.logger.Warn("Skip execution of policy because of current time before scheduled",
			zap.String("policy_name", item.policyName),
			zap.String("schedule_type", string(item.scheduleType)),
			zap.Stringer("time", now),
			zap.Stringer("scheduled_at_time", item.fireAt))
		return nil
	}

	if now.After(item.deadline()) {
		c.logger.Warn("Skip execution of policy because of current time after deadline",
			zap.String("policy_name", item.policyName),
			zap.String("schedule_type", string(item.scheduleType)),
			zap.Stringer("time", now),
			zap.Stringer("scheduled_deadline", item.deadline()))
		return nil
	}

	state, exists := c.state.Get(item.policyName)
	policy, err := c.lister.Stands.Get(item.policyName)
	if errors.IsNotFound(err) || !exists {
		c.logger.Warn("Skip execution of policy because it not exists", zap.String("policy_name", item.policyName))
		return nil
	}

	if err != nil {
		return err
	}

	c.logger.Info("Execute schedule of policy",
		zap.String("policy_name", item.policyName),
		zap.String("schedule_type", string(item.scheduleType)))

	ctx, cancel := context.WithTimeout(context.Background(), _ExecutionTimeout)
	defer cancel()

	switch item.scheduleType {
	case apis.StatusShutdown:
		err = c.executor.ExecuteShutdown(ctx, policy)
		state.UpdateStatus(item.scheduleType, now, err)
	case apis.StatusStartup:
		err = c.executor.ExecuteStartup(ctx, policy)
		state.UpdateStatus(item.scheduleType, now, err)
	default:
		err = fmt.Errorf("not supported schedule type specified: %s", item.scheduleType)
	}

	if err != nil {
		c.logger.Error("Failed to execute schedule of policy",
			zap.String("policy_name", item.policyName),
			zap.String("schedule_type", string(item.scheduleType)),
			zap.Error(err))
		return err
	}

	c.logger.Info("Complete to execute schedule of policy",
		zap.String("policy_name", item.policyName),
		zap.String("schedule_type", string(item.scheduleType)))
	return nil
}
