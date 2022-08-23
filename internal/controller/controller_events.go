package controller

import (
	"go.uber.org/zap"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/state"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

func (c *Controller) add(obj *apis.StandSchedulePolicy) {
	c.logger.Debug("Discovered policy object", zap.String("policy_name", obj.Name))
	ps, err := state.NewPolicyState(&obj.Spec.Schedules)
	if err != nil {
		c.logger.Error("Policy object has invalid format", zap.String("policy_name", obj.Name), zap.Error(err))
		return
	}

	c.logger.Info("Added policy object", zap.String("policy_name", obj.Name))
	c.state.AddOrUpdate(obj.Name, ps)
	c.enqueueReconcile(obj.Name)
}

func (c *Controller) update(oldObj, newObj *apis.StandSchedulePolicy) {
	c.logger.Info("Sync policy object with", zap.String("policy_name", newObj.Name))
	oldState, err := state.NewPolicyState(&oldObj.Spec.Schedules)
	if err != nil {
		c.logger.Error("Policy object has invalid format", zap.String("policy_name", oldObj.Name), zap.Error(err))
		return
	}
	newState, err := state.NewPolicyState(&newObj.Spec.Schedules)
	if err != nil {
		c.logger.Error("Policy object has invalid format", zap.String("policy_name", newObj.Name), zap.Error(err))
		return
	}

	if !oldState.ScheduleEquals(newState) {
		c.state.AddOrUpdate(newObj.Name, newState)
	}

	c.enqueueReconcile(newObj.Name)
}

func (c *Controller) delete(obj *apis.StandSchedulePolicy) {
	c.logger.Info("Deleted policy object", zap.String("policy_name", obj.Name))
	c.state.Delete(obj.Name)
	c.enqueueReconcile(obj.Name)
}
