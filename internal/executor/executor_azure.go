package executor

import (
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

// todo: handle external resources

func (in *Executor) executeShutdownAzure(policy *apis.StandSchedulePolicy) error {
	return nil
}

func (in *Executor) executeStartupAzure(policy *apis.StandSchedulePolicy) error {
	return nil
}
