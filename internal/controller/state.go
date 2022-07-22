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
		data map[string]*ScheduleInfo
	}

	ScheduleInfo struct {
		startupSchedule  cron.Schedule
		startupOverride  time.Time
		shutdownSchedule cron.Schedule
		shutdownOverride time.Time
	}
)

func NewControllerState() *State {
	return &State{
		data: make(map[string]*ScheduleInfo),
	}
}

func NewScheduleInfo(policy *apis.StandSchedulePolicy) (*ScheduleInfo, error) {
	startupSchedule, err := cron.ParseStandard(policy.Spec.Schedule.Startup)
	if err != nil {
		return nil, err
	}

	shutdownSchedule, err := cron.ParseStandard(policy.Spec.Schedule.Shutdown)
	if err != nil {
		return nil, err
	}

	startupOverride, _ := time.Parse(time.RFC3339, policy.Annotations[apis.AnnotationScheduleStartupTime])
	shutdownOverride, _ := time.Parse(time.RFC3339, policy.Annotations[apis.AnnotationScheduleShutdownTime])

	return &ScheduleInfo{
		startupSchedule:  startupSchedule,
		startupOverride:  startupOverride,
		shutdownSchedule: shutdownSchedule,
		shutdownOverride: shutdownOverride,
	}, nil
}

func (s *State) AddOrUpdate(key string, info *ScheduleInfo) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// todo: notifications
	s.data[key] = info
}

func (s *State) Delete(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.data, key)
}
