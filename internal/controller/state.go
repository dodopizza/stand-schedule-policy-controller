package controller

import (
	"sync"
	"time"

	"github.com/robfig/cron"

	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

type (
	State struct {
		lock sync.Mutex
		data map[string]*ScheduleState
	}

	ScheduleState struct {
		startup  *Schedule
		shutdown *Schedule
	}

	Schedule struct {
		schedule cron.Schedule
		override time.Time
	}
)

func NewControllerState() *State {
	return &State{
		data: make(map[string]*ScheduleState),
	}
}

func NewScheduleState(policy *apis.StandSchedulePolicy) (*ScheduleState, error) {
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

	return &ScheduleState{
		startup:  startup,
		shutdown: shutdown,
	}, nil
}

func NewSchedule(schedule, override string) (*Schedule, error) {
	sc, err := cron.ParseStandard(schedule)
	if err != nil {
		return nil, err
	}

	ov, _ := time.Parse(time.RFC3339, override)

	return &Schedule{
		schedule: sc,
		override: ov,
	}, nil
}

func (s *ScheduleState) IsScheduleEquals(other *ScheduleState) bool {
	return s.startup.Equals(other.startup) && s.shutdown.Equals(other.shutdown)
}

func (s *Schedule) GetNextTimeAfter(since time.Time) time.Time {
	// todo: store when override expires ?

	if s.override.After(since) {
		return s.override
	}
	return s.schedule.Next(since)
}

func (s *Schedule) Equals(other *Schedule) bool {
	return s.schedule == other.schedule && s.override == other.override
}

func (s *State) AddOrUpdate(key string, info *ScheduleState) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[key] = info
}

func (s *State) Get(key string) (*ScheduleState, bool) {
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
