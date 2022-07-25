package controller

import (
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (c *Controller) reconcile(i interface{}) error {
	policy, err := c.lister.stands.Get(i.(string))

	if errors.IsNotFound(err) {
		c.logger.Info("Deleted policy with name removed from execution", zap.String("policy_name", i.(string)))
		return nil
	}

	_, exists := c.state.Get(policy.Name)
	if !exists {
		c.logger.Info("Deleted policy with name removed from execution", zap.String("policy_name", policy.Name))
		return nil
	}

	// todo: update status

	return nil
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
	next := schedule.GetNextTimeAfter(since)

	c.logger.Info("Schedule policy with name at time (since)",
		zap.String("policy_name", policyName),
		zap.String("schedule_type", scheduleType),
		zap.Stringer("since", since),
		zap.Stringer("at", next),
	)

	if next.IsZero() {
		c.logger.Error("Failed to schedule policy after time with error",
			zap.String("policy_name", policyName),
			zap.String("schedule_type", scheduleType),
			zap.Stringer("since", since),
		)
	}

	item := WorkItem{
		PolicyName:  policyName,
		Type:        scheduleType,
		ScheduledAt: next,
		Deadline:    next.Add(30 * time.Minute),
	}

	c.executor.EnqueueAfter(item, next.Sub(since))
}
