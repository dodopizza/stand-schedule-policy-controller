/*
Copyright Dodo Engineering

Authored by The Infrastructure Platform Team.
*/

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PolicyStatusConditionType string

const (
	// PolicyCreated indicates whether policy resource just created and ready to be scheduled.
	PolicyCreated PolicyStatusConditionType = "Created"
	// PolicyScheduled means that policy actions are in progress.
	PolicyScheduled PolicyStatusConditionType = "Scheduled"
	// PolicyCompleted means that policy actions completed and successful.
	PolicyCompleted PolicyStatusConditionType = "Completed"
	// PolicyFailed means that policy actions completed and failed.
	PolicyFailed PolicyStatusConditionType = "Failed"
)

// StandSchedulePolicyStatus is a status for StandSchedulePolicy resource.
type StandSchedulePolicyStatus struct {
	// Conditions defines current service state of policy.
	Conditions []PolicyStatusCondition `json:"conditions"`
}

type PolicyStatusCondition struct {
	// Type is the type of the condition.
	Type PolicyStatusConditionType `json:"type"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
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
