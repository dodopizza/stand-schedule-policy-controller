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
		schedule    cron.Schedule
		override    time.Time
		fireAt      time.Time
		completedAt time.Time
		failedAt    time.Time
	}
)

func NewSchedule(schedule apis.CronSchedule) (*ScheduleState, error) {
	sc, err := cron.ParseStandard(schedule.Cron)
	if err != nil {
		return nil, err
	}

	ov, _ := time.Parse(time.RFC3339, schedule.Override)

	return &ScheduleState{
		schedule: sc,
		override: ov,
	}, nil
}

func (ss *ScheduleState) GetExecutedTime() time.Time {
	if !ss.completedAt.IsZero() {
		return ss.completedAt
	}
	return ss.failedAt
}

func (ss *ScheduleState) GetFireTime() time.Time {
	return ss.fireAt
}

func (ss *ScheduleState) GetNextExecutionTime(since time.Time) time.Time {
	if ss.override.After(since) {
		return ss.override
	}
	return ss.schedule.Next(since)
}

func (ss *ScheduleState) GetConditions(st apis.ConditionScheduleType) []apis.StatusCondition {
	conditions := []apis.StatusCondition{}

	if !ss.fireAt.IsZero() {
		conditions = append(conditions, apis.StatusCondition{
			Type:               apis.ConditionScheduled,
			Status:             st,
			LastTransitionTime: meta.NewTime(ss.fireAt),
		})
	}

	if !ss.completedAt.IsZero() {
		conditions = append(conditions, apis.StatusCondition{
			Type:               apis.ConditionCompleted,
			Status:             st,
			LastTransitionTime: meta.NewTime(ss.completedAt),
		})
	}

	if !ss.failedAt.IsZero() {
		conditions = append(conditions, apis.StatusCondition{
			Type:               apis.ConditionFailed,
			Status:             st,
			LastTransitionTime: meta.NewTime(ss.failedAt),
		})
	}

	return conditions
}

func (ss *ScheduleState) SetFiredAfter(ts time.Time) {
	ss.fireAt = ss.GetNextExecutionTime(ts)
	ss.failedAt = time.Time{}
	ss.completedAt = time.Time{}
}

func (ss *ScheduleState) SetCompleted(at time.Time) {
	ss.completedAt = at
	ss.failedAt = time.Time{}
}

func (ss *ScheduleState) SetFailed(at time.Time) {
	ss.failedAt = at
	ss.completedAt = time.Time{}
}

func (ss *ScheduleState) ScheduleRequired(current time.Time) bool {
	// not scheduled at all
	if ss.fireAt.IsZero() {
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
	return reflect.DeepEqual(ss.schedule, other.schedule) && ss.override == other.override
}
