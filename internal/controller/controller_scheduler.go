package controller

import (
	"time"

	"go.uber.org/zap"
)

func (c *Controller) schedule(i interface{}) error {
	key := i.(string)
	since := c.clock.Now()

	pair, exists := c.state.Get(key)
	if !exists {
		// todo: remove from execution queue ?
		c.logger.Info("Deleted policy with name removed from execution", zap.String("policy_name", key))
		return nil
	}

	// todo: schedule startup

	next := pair.shutdown.GetNextTimeAfter(since)
	c.logger.Info("Schedule policy with name at time (since)",
		zap.String("policy_name", key),
		zap.Stringer("since", since),
		zap.Stringer("at", next),
	)

	if next.IsZero() {
		c.logger.Error("Failed to schedule policy after time with error",
			zap.String("policy_name", key),
			zap.Stringer("since", since),
		)
		return nil
	}

	item := WorkItem{
		PolicyName:  key,
		Type:        "shutdown",
		ScheduledAt: next,
		Deadline:    next.Add(30 * time.Minute),
	}
	c.executor.EnqueueAfter(item, next.Sub(since))

	return nil
}
