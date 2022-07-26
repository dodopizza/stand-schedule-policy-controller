package state

import (
	"reflect"
	"time"

	"github.com/robfig/cron"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

type (
	ScheduleState struct {
		Schedule    cron.Schedule
		Override    time.Time
		FireAt      time.Time
		CompletedAt time.Time
		FailedAt    time.Time
	}
)

func NewSchedule(schedule, override string) (*ScheduleState, error) {
	sc, err := cron.ParseStandard(schedule)
	if err != nil {
		return nil, err
	}

	ov, _ := time.Parse(time.RFC3339, override)

	return &ScheduleState{
		Schedule: sc,
		Override: ov,
	}, nil
}

func (ss *ScheduleState) GetExecutedTime() time.Time {
	if !ss.CompletedAt.IsZero() {
		return ss.CompletedAt
	}
	return ss.FailedAt
}

func (ss *ScheduleState) GetNextExecutionTime(since time.Time) time.Time {
	// todo: store when override expires ?

	if ss.Override.After(since) {
		return ss.Override
	}

	return ss.Schedule.Next(since)
}

func (ss *ScheduleState) SetFiredSince(since time.Time) {
	ss.FireAt = ss.GetNextExecutionTime(since)
	ss.FailedAt = time.Time{}
	ss.CompletedAt = time.Time{}
}

func (ss *ScheduleState) SetCompleted(at time.Time) {
	ss.CompletedAt = at
	ss.FailedAt = time.Time{}
}

func (ss *ScheduleState) SetFailed(at time.Time) {
	ss.FailedAt = at
	ss.CompletedAt = time.Time{}
}

func (ss *ScheduleState) GetConditions(st apis.ConditionScheduleType) []apis.StatusCondition {
	conditions := []apis.StatusCondition{}

	if !ss.FireAt.IsZero() {
		conditions = append(conditions, apis.StatusCondition{
			Type:               apis.ConditionScheduled,
			Status:             st,
			LastTransitionTime: meta.NewTime(ss.FireAt),
		})
	}

	if !ss.CompletedAt.IsZero() {
		conditions = append(conditions, apis.StatusCondition{
			Type:               apis.ConditionCompleted,
			Status:             st,
			LastTransitionTime: meta.NewTime(ss.CompletedAt),
		})
	}

	if !ss.FailedAt.IsZero() {
		conditions = append(conditions, apis.StatusCondition{
			Type:               apis.ConditionFailed,
			Status:             st,
			LastTransitionTime: meta.NewTime(ss.FailedAt),
		})
	}

	return conditions
}

func (ss *ScheduleState) ScheduleRequired(current time.Time) bool {
	// not scheduled at all
	if ss.FireAt.IsZero() {
		return true
	}

	executed := ss.GetExecutedTime()

	// already scheduled but not executed
	if executed.IsZero() {
		return false
	}

	// schedule when current > (next - executed) / 2
	next := ss.GetNextExecutionTime(current)
	delta := next.Sub(executed) / 2

	return current.After(executed.Add(delta))
}

func (ss *ScheduleState) Equals(other *ScheduleState) bool {
	return reflect.DeepEqual(ss.Schedule, other.Schedule) && ss.Override == other.Override
}
