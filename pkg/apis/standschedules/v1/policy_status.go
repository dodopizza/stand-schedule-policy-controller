/*
Copyright Dodo Engineering

Authored by The Infrastructure Platform Team.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConditionType string
type ConditionStatus string

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
	StatusStartup ConditionStatus = "Startup"
	// StatusShutdown means that current status for shutdown operation
	StatusShutdown ConditionStatus = "Shutdown"
)

// StandSchedulePolicyStatus is a status for StandSchedulePolicy resource.
type StandSchedulePolicyStatus struct {
	// Conditions defines current service state of policy.
	Conditions []PolicyStatusCondition `json:"conditions"`
}

type PolicyStatusCondition struct {
	// Type is the type of the condition.
	Type ConditionType `json:"type"`
	// Status is the status of the condition.
	// Can be StatusStartup or StatusShutdown
	Status ConditionStatus `json:"status"`
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
