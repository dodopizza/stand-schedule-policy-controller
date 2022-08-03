package controller

import (
	"context"
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/state"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

func (c *Controller) reconcile(i interface{}) error {
	policyName := i.(string)
	policy, err := c.lister.stands.Get(policyName)

	if errors.IsNotFound(err) {
		c.logger.Info("Deleted policy with name removed from execution", zap.String("policy_name", policyName))
		return nil
	}

	ps, exists := c.state.Get(policy.Name)
	if !exists {
		c.logger.Info("Deleted policy with name removed from execution", zap.String("policy_name", policy.Name))
		return nil
	}

	c.scheduleIfRequired(policy, ps)

	c.logger.Info("Update policy status", zap.String("policy_name", policy.Name))
	policy.Status.UpdateConditions(ps.GetConditions())

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

func (c *Controller) scheduleIfRequired(policy *apis.StandSchedulePolicy, ps *state.PolicyState) {
	ts := c.clock.Now()

	if ps.ScheduleRequired(apis.StatusShutdown, ts) {
		c.schedule(ts, policy.Name, apis.StatusShutdown, ps.GetSchedule(apis.StatusShutdown))
	}

	if ps.ScheduleRequired(apis.StatusStartup, ts) {
		c.schedule(ts, policy.Name, apis.StatusStartup, ps.GetSchedule(apis.StatusStartup))
	}
}

func (c *Controller) schedule(
	ts time.Time,
	policyName string,
	scheduleType apis.ConditionScheduleType,
	schedule *state.ScheduleState,
) {
	schedule.SetFiredAfter(ts)

	if schedule.GetFireTime().IsZero() {
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
		zap.Stringer("at", schedule.GetFireTime()))

	item := WorkItem{
		policyName:   policyName,
		scheduleType: scheduleType,
		fireAt:       schedule.GetFireTime(),
	}

	c.executor.EnqueueAfter(item, schedule.GetFireTime().Sub(ts))
}
