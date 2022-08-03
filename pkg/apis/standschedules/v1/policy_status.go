/*
Copyright Dodo Engineering

Authored by The Infrastructure Platform Team.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConditionType string
type ConditionScheduleType string

const (
	// ConditionScheduled means that policy actions are in progress.
	ConditionScheduled ConditionType = "Scheduled"
	// ConditionCompleted means that policy actions completed and successful.
	ConditionCompleted ConditionType = "Completed"
	// ConditionFailed means that policy actions completed and failed.
	ConditionFailed ConditionType = "Failed"
)

const (
	// StatusStartup means that current status for startup operation
	StatusStartup ConditionScheduleType = "Startup"
	// StatusShutdown means that current status for shutdown operation
	StatusShutdown ConditionScheduleType = "Shutdown"
)

// StandSchedulePolicyStatus is a status for StandSchedulePolicy resource.
type StandSchedulePolicyStatus struct {
	// Conditions defines current service state of policy.
	Conditions []StatusCondition `json:"conditions,omitempty"`
	// Startup defines status of startup schedule
	Startup ScheduleStatus `json:"startup,omitempty"`
	// Shutdown defines status of shutdown schedule
	Shutdown ScheduleStatus `json:"shutdown,omitempty"`
}

type StatusCondition struct {
	// Type is the type of the condition.
	Type ConditionType `json:"type"`
	// Status is the status of the condition.
	// Can be Startup or Shutdown.
	Status ConditionScheduleType `json:"status"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

type ScheduleStatus struct {
	// ScheduledTime defines time when cron scheduled
	ScheduledTime metav1.Time `json:"scheduledTime,omitempty"`
	// Status defines how schedule finished
	Status string `json:"status,omitempty"`
	// Status defines when schedule finished
	StatusTime metav1.Time `json:"statusTime,omitempty"`
}

func (in *StandSchedulePolicyStatus) GetScheduleStatus(st ConditionScheduleType) *ScheduleStatus {
	switch st {
	case StatusStartup:
		return &in.Startup
	case StatusShutdown:
		return &in.Shutdown
	}
	return nil
}

func (in *StandSchedulePolicyStatus) UpdateConditions(conditions []StatusCondition) {
	in.Conditions = conditions
	in.Startup.UpdateFromConditions(StatusStartup, conditions)
	in.Shutdown.UpdateFromConditions(StatusShutdown, conditions)
}

func (in *ScheduleStatus) UpdateFromConditions(st ConditionScheduleType, conditions []StatusCondition) {
	for _, condition := range conditions {
		if condition.Status != st {
			continue
		}

		if condition.Type == ConditionScheduled {
			in.ScheduledTime = condition.LastTransitionTime
		}

		if condition.Type == ConditionCompleted || condition.Type == ConditionFailed {
			in.Status = string(condition.Type)
			in.StatusTime = condition.LastTransitionTime
		}
	}
}
