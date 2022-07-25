package controller

import (
	"go.uber.org/zap"

	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

func (c *Controller) add(obj *apis.StandSchedulePolicy) {
	c.logger.Debug("Discovered policy object with name", zap.String("policy_name", obj.Name))
	state, err := NewScheduleState(obj)
	if err != nil {
		c.logger.Error("Policy object with name has invalid format", zap.String("policy_name", obj.Name), zap.Error(err))
		return
	}

	c.logger.Info("Added policy object with name", zap.String("policy_name", obj.Name))
	c.state.AddOrUpdate(obj.Name, state)
	c.reconciler.Enqueue(obj.Name)

	if err := c.schedule(c.clock.Now(), obj.Name, state); err != nil {
		c.logger.Error("Failed to schedule policy with name",
			zap.String("policy_name", obj.Name),
			zap.Error(err))
	}
}

func (c *Controller) update(oldObj, newObj *apis.StandSchedulePolicy) {
	c.logger.Info("Sync policy object with name", zap.String("policy_name", newObj.Name))
	oldState, err := NewScheduleState(newObj)
	if err != nil {
		c.logger.Error("Policy object with name has invalid format", zap.String("policy_name", oldObj.Name), zap.Error(err))
		return
	}
	newState, err := NewScheduleState(newObj)
	if err != nil {
		c.logger.Error("Policy object with name has invalid format", zap.String("policy_name", newObj.Name), zap.Error(err))
		return
	}

	if oldState.IsScheduleEquals(newState) {
		c.reconciler.Enqueue(newObj.Name)
		return
	}

	c.state.AddOrUpdate(newObj.Name, newState)

	if err := c.schedule(c.clock.Now(), newObj.Name, newState); err != nil {
		c.logger.Error("Failed to schedule policy with name",
			zap.String("policy_name", newObj.Name),
			zap.Error(err))
	}
}

func (c *Controller) delete(obj *apis.StandSchedulePolicy) {
	c.logger.Info("Deleted policy object with name", zap.String("policy_name", obj.Name))
	c.state.Delete(obj.Name)
	c.reconciler.Enqueue(obj.Name)
}
