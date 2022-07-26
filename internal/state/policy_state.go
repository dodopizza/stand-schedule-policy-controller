package state

import (
	"time"

	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

type (
	PolicyState struct {
		startup  *ScheduleState
		shutdown *ScheduleState
	}
)

func NewPolicyState(policy *apis.StandSchedulePolicy) (*PolicyState, error) {
	startup, err := NewSchedule(
		policy.Spec.Schedule.Startup,
		policy.Spec.Schedule.StartupOverride,
	)
	if err != nil {
		return nil, err
	}

	shutdown, err := NewSchedule(
		policy.Spec.Schedule.Shutdown,
		policy.Spec.Schedule.ShutdownOverride,
	)
	if err != nil {
		return nil, err
	}

	return &PolicyState{
		startup:  startup,
		shutdown: shutdown,
	}, nil
}

func (ps *PolicyState) GetSchedule(st apis.ConditionScheduleType) *ScheduleState {
	switch st {
	case apis.StatusStartup:
		return ps.startup
	case apis.StatusShutdown:
		return ps.shutdown
	}
	return nil
}

func (ps *PolicyState) GetConditions() []apis.StatusCondition {
	conditions := []apis.StatusCondition{}
	conditions = append(conditions, ps.startup.GetConditions(apis.StatusStartup)...)
	conditions = append(conditions, ps.shutdown.GetConditions(apis.StatusShutdown)...)
	return conditions
}

func (ps *PolicyState) UpdateStatus(st apis.ConditionScheduleType, at time.Time, err error) {
	schedule := ps.GetSchedule(st)

	if err != nil {
		schedule.SetFailed(at)
	} else {
		schedule.SetCompleted(at)
	}
}

func (ps *PolicyState) ScheduleEquals(other *PolicyState) bool {
	return ps.startup.Equals(other.startup) && ps.shutdown.Equals(other.shutdown)
}

func (ps *PolicyState) ScheduleRequired(st apis.ConditionScheduleType, at time.Time) bool {
	switch st {
	case apis.StatusStartup:
		return ps.startup.ScheduleRequired(at)
	case apis.StatusShutdown:
		return ps.shutdown.ScheduleRequired(at)
	}
	return false
}
