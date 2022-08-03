/*
Copyright Dodo Engineering

Authored by The Infrastructure Platform Team.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AnnotationPrefix = "standschedule." + GroupName
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=standschedulepolicies,scope="Cluster",shortName=sspol
// +kubebuilder:printcolumn:name="StartupScheduledTime",type="date",JSONPath=".status.startup.scheduledTime"
// +kubebuilder:printcolumn:name="StartupStatus",type="string",JSONPath=".status.startup.status"
// +kubebuilder:printcolumn:name="StartupStatusTime",type="date",JSONPath=".status.startup.statusTime"
// +kubebuilder:printcolumn:name="ShutdownScheduledTime",type="date",JSONPath=".status.shutdown.scheduledTime"
// +kubebuilder:printcolumn:name="ShutdownStatus",type="string",JSONPath=".status.shutdown.status"
// +kubebuilder:printcolumn:name="ShutdownStatusTime",type="dateTime",JSONPath=".status.shutdown.statusTime"

// StandSchedulePolicy declares policy for stand startup/shutdown schedules
type StandSchedulePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec declares schedule behavior.
	Spec StandSchedulePolicySpec `json:"spec"`

	// Status contains schedule runtime data.
	// +optional
	Status StandSchedulePolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StandSchedulePolicyList is a list of StandSchedulePolicy resources
type StandSchedulePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []StandSchedulePolicy `json:"items"`
}

// StandSchedulePolicySpec is a spec for StandSchedulePolicy resource.
type StandSchedulePolicySpec struct {
	// TargetNamespaceFilter defines regex filter to match namespaces to process.
	TargetNamespaceFilter string `json:"targetNamespaceFilter"`

	// Schedules contains schedules spec.
	Schedules SchedulesSpec `json:"schedules"`

	// Resources contains external resources spec.
	// +optional
	Resources ResourcesSpec `json:"resources,omitempty"`
}

// SchedulesSpec defines supported schedules for policy.
type SchedulesSpec struct {
	// Startup defines schedule for startup.
	Startup CronSchedule `json:"startup"`

	// Shutdown defines schedule for shutdown.
	Shutdown CronSchedule `json:"shutdown"`
}

// CronSchedule defines schedule (as cron) and optional override (as time string).
type CronSchedule struct {
	// Cron is a cron format schedule.
	Cron string `json:"cron"`

	// Override is an override as time string (formatted as FRC3339)
	// +optional
	Override string `json:"override,omitempty"`
}
