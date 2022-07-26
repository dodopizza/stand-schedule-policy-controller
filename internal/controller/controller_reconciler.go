package controller

import (
	"context"
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

func (c *Controller) reconcile(i interface{}) error {
	policyName := i.(string)
	policy, err := c.lister.stands.Get(policyName)

	if errors.IsNotFound(err) {
		c.logger.Info("Deleted policy with name removed from execution", zap.String("policy_name", policyName))
		return nil
	}

	state, exists := c.state.Get(policy.Name)
	if !exists {
		c.logger.Info("Deleted policy with name removed from execution", zap.String("policy_name", policy.Name))
		return nil
	}

	c.scheduleIfRequired(policy, state)

	c.logger.Info("Update policy status", zap.String("policy_name", policy.Name))
	policy.Status.Conditions = state.GetConditions()

	_, err = c.kube.StandSchedulesClient().
		StandSchedulesV1().
		StandSchedulePolicies().
		UpdateStatus(context.Background(), policy, meta.UpdateOptions{})

	if err != nil {
		c.logger.Error("Failed to update policy status",
			zap.String("policy_name", policy.Name),
			zap.Error(err))
	}

	return err
}

func (c *Controller) scheduleIfRequired(policy *apis.StandSchedulePolicy, state *PolicyState) {
	ts := c.clock.Now()

	if state.shutdown.ScheduleRequired(ts) {
		c.schedule(ts, policy.Name, apis.StatusShutdown, state.shutdown)
	}

	if state.startup.ScheduleRequired(ts) {
		c.schedule(ts, policy.Name, apis.StatusStartup, state.startup)
	}
}

func (c *Controller) schedule(
	ts time.Time,
	policyName string,
	scheduleType apis.ConditionScheduleType,
	schedule *ScheduleState,
) {
	schedule.SetFiredSince(ts)

	if schedule.fireAt.IsZero() {
		c.logger.Error("Failed to schedule policy",
			zap.String("policy_name", policyName),
			zap.String("schedule_type", string(scheduleType)),
			zap.Stringer("since", ts))
		return
	}

	c.logger.Info("Schedule policy",
		zap.String("policy_name", policyName),
		zap.String("schedule_type", string(scheduleType)),
		zap.Stringer("since", ts),
		zap.Stringer("at", schedule.fireAt))

	item := WorkItem{
		policyName:   policyName,
		scheduleType: scheduleType,
		fireAt:       schedule.fireAt,
	}

	c.executor.EnqueueAfter(item, schedule.fireAt.Sub(ts))
}
