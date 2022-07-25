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
	policy, err := c.lister.stands.Get(i.(string))

	if errors.IsNotFound(err) {
		c.logger.Info("Deleted policy with name removed from execution", zap.String("policy_name", i.(string)))
		return nil
	}

	state, exists := c.state.Get(policy.Name)
	if !exists {
		c.logger.Info("Deleted policy with name removed from execution", zap.String("policy_name", policy.Name))
		return nil
	}

	c.logger.Info("Update policy status", zap.String("policy_name", policy.Name))
	conditions := []apis.PolicyStatusCondition{}
	conditions = append(conditions, state.startup.Conditions(apis.ScheduleStartup)...)
	conditions = append(conditions, state.shutdown.Conditions(apis.ScheduleShutdown)...)
	policy.Status.Conditions = conditions

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

func (c *Controller) schedule(
	since time.Time,
	policyName string,
	scheduleState *ScheduleState,
) {
	c.scheduleWorkItem(since, policyName, "shutdown", scheduleState.shutdown)
	c.scheduleWorkItem(since, policyName, "startup", scheduleState.startup)
}

func (c *Controller) scheduleWorkItem(
	since time.Time,
	policyName string,
	scheduleType string,
	schedule *Schedule,
) {
	schedule.UpdateScheduledTime(since)

	c.logger.Info("Schedule policy with name at time (since)",
		zap.String("policy_name", policyName),
		zap.String("schedule_type", scheduleType),
		zap.Stringer("since", since),
		zap.Stringer("at", schedule.fireAt),
	)

	if schedule.fireAt.IsZero() {
		c.logger.Error("Failed to schedule policy after time with error",
			zap.String("policy_name", policyName),
			zap.String("schedule_type", scheduleType),
			zap.Stringer("since", since),
		)
	}

	item := WorkItem{
		policyName:   policyName,
		scheduleType: scheduleType,
		fireAt:       schedule.fireAt,
	}

	c.executor.EnqueueAfter(item, schedule.fireAt.Sub(since))
}
