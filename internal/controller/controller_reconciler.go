package controller

import (
	"time"

	"go.uber.org/zap"
)

func (c *Controller) reconcile(i interface{}) error {
	// todo: update status

	//since := c.clock.Now()
	//
	//policy, err := c.lister.stands.Get(i.(string))
	//if errors.IsNotFound(err) {
	//	c.logger.Info("Deleted policy with name removed from execution", zap.String("policy_name", policy.Name))
	//	return nil
	//}
	//
	//scheduleState, exists := c.state.Get(policy.Name)
	//if !exists {
	//	c.logger.Info("Deleted policy with name removed from execution", zap.String("policy_name", policy.Name))
	//	return nil
	//}
	//
	//
	//return c.schedule(since, policy.Name, scheduleState)

	return nil
}

func (c *Controller) schedule(
	since time.Time,
	policyName string,
	scheduleState *ScheduleState,
) error {
	// todo: schedule startup

	next := scheduleState.shutdown.GetNextTimeAfter(since)

	c.logger.Info("Schedule policy with name at time (since)",
		zap.String("policy_name", policyName),
		zap.Stringer("since", since),
		zap.Stringer("at", next),
	)

	if next.IsZero() {
		c.logger.Error("Failed to schedule policy after time with error",
			zap.String("policy_name", policyName),
			zap.Stringer("since", since),
		)
		return nil
	}

	item := WorkItem{
		PolicyName:  policyName,
		Type:        "shutdown",
		ScheduledAt: next,
		Deadline:    next.Add(30 * time.Minute),
	}
	c.executor.EnqueueAfter(item, next.Sub(since))

	return nil
}
