package state

import (
	"time"

	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

type (
	PolicyState struct {
		Startup  *ScheduleState
		Shutdown *ScheduleState
	}
)

func NewPolicyState(policy *apis.StandSchedulePolicy) (*PolicyState, error) {
	startup, err := NewSchedule(
		policy.Spec.Schedule.Startup,
		policy.ObjectMeta.Annotations[apis.AnnotationScheduleStartupTime],
	)
	if err != nil {
		return nil, err
	}

	shutdown, err := NewSchedule(
		policy.Spec.Schedule.Shutdown,
		policy.ObjectMeta.Annotations[apis.AnnotationScheduleShutdownTime],
	)
	if err != nil {
		return nil, err
	}

	return &PolicyState{
		Startup:  startup,
		Shutdown: shutdown,
	}, nil
}

func (ps *PolicyState) GetSchedule(st apis.ConditionScheduleType) *ScheduleState {
	switch st {
	case apis.StatusStartup:
		return ps.Startup
	case apis.StatusShutdown:
		return ps.Shutdown
	}
	return nil
}

func (ps *PolicyState) UpdateStatus(at time.Time, err error, st apis.ConditionScheduleType) {
	schedule := ps.GetSchedule(st)

	if err != nil {
		schedule.SetFailed(at)
	} else {
		schedule.SetCompleted(at)
	}
}

func (ps *PolicyState) ScheduleEquals(other *PolicyState) bool {
	return ps.Startup.Equals(other.Startup) && ps.Shutdown.Equals(other.Shutdown)
}

func (ps *PolicyState) GetConditions() []apis.StatusCondition {
	conditions := []apis.StatusCondition{}
	conditions = append(conditions, ps.Startup.GetConditions(apis.StatusStartup)...)
	conditions = append(conditions, ps.Shutdown.GetConditions(apis.StatusShutdown)...)
	return conditions
}
