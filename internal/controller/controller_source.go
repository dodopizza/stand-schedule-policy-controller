package controller

import (
	"go.uber.org/zap"

	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

func (c *Controller) add(obj *apis.StandSchedulePolicy) {
	c.logger.Debug("Discovered policy object with name", zap.String("policy_name", obj.Name))
	pair, err := NewScheduleState(obj)
	if err != nil {
		c.logger.Error("Policy object with name has invalid format", zap.String("policy_name", obj.Name), zap.Error(err))
		return
	}
	c.logger.Info("Added policy object with name", zap.String("policy_name", obj.Name))
	c.state.AddOrUpdate(obj.Name, pair)
	c.scheduler.Enqueue(obj.Name)
}

func (c *Controller) update(oldObj, newObj *apis.StandSchedulePolicy) {
	// skip same versions here
	if oldObj.ResourceVersion == newObj.ResourceVersion {
		return
	}

	c.logger.Info("Sync policy object with name", zap.String("policy_name", newObj.Name))
	pair, err := NewScheduleState(newObj)
	if err != nil {
		c.logger.Error("Policy object with name has invalid format", zap.String("policy_name", newObj.Name), zap.Error(err))
		return
	}

	// todo: enqueue only if schedules changed
	c.state.AddOrUpdate(newObj.Name, pair)
	c.scheduler.Enqueue(newObj.Name)
}

func (c *Controller) delete(obj *apis.StandSchedulePolicy) {
	c.logger.Info("Deleted policy object with name", zap.String("policy_name", obj.Name))
	c.state.Delete(obj.Name)
	c.scheduler.Enqueue(obj.Name)
}
