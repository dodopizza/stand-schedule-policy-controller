package controller

import (
	"reflect"
	"sync"
	"time"

	"github.com/robfig/cron"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

type (
	State struct {
		lock sync.Mutex
		data map[string]*PolicyState
	}

	PolicyState struct {
		startup  *ScheduleState
		shutdown *ScheduleState
	}

	ScheduleState struct {
		schedule    cron.Schedule
		override    time.Time
		fireAt      time.Time
		completedAt time.Time
		failedAt    time.Time
	}
)

func NewControllerState() *State {
	return &State{
		data: make(map[string]*PolicyState),
	}
}

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
		startup:  startup,
		shutdown: shutdown,
	}, nil
}

func NewSchedule(schedule, override string) (*ScheduleState, error) {
	sc, err := cron.ParseStandard(schedule)
	if err != nil {
		return nil, err
	}

	ov, _ := time.Parse(time.RFC3339, override)

	return &ScheduleState{
		schedule: sc,
		override: ov,
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

func (ps *PolicyState) UpdateStatus(at time.Time, err error, st apis.ConditionScheduleType) {
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

func (ps *PolicyState) GetConditions() []apis.StatusCondition {
	conditions := []apis.StatusCondition{}
	conditions = append(conditions, ps.startup.GetConditions(apis.StatusStartup)...)
	conditions = append(conditions, ps.shutdown.GetConditions(apis.StatusShutdown)...)
	return conditions
}

func (ss *ScheduleState) GetNextTimeAfter(since time.Time) time.Time {
	// todo: store when override expires ?

	if ss.override.After(since) {
		return ss.override
	}

	return ss.schedule.Next(since)
}

func (ss *ScheduleState) SetFiredSince(since time.Time) {
	ss.fireAt = ss.GetNextTimeAfter(since)
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

func (ss *ScheduleState) Equals(other *ScheduleState) bool {
	return reflect.DeepEqual(ss.schedule, other.schedule) && ss.override == other.override
}

func (s *State) AddOrUpdate(key string, info *PolicyState) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[key] = info
}

func (s *State) Get(key string) (*PolicyState, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	v, exists := s.data[key]
	return v, exists
}

func (s *State) Delete(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.data, key)
}
