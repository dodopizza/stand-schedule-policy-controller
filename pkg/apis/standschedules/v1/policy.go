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

	// Schedule contains schedule spec.
	Schedule ScheduleSpec `json:"schedule"`

	// Resources contains external resources spec.
	Resources ResourcesSpec `json:"resources"`
}

// ScheduleSpec defines startup and shutdown schedules of a policy.
type ScheduleSpec struct {
	// Startup defines schedule in cron format for stand startup.
	Startup string `json:"startup"`

	// StartupOverride defines override for startup schedule as time string
	StartupOverride string `json:"startupOverride,omitempty"`

	// Shutdown defines schedule in cron format for stand shutdown.
	Shutdown string `json:"shutdown"`

	// StartupOverride defines override for shutdown schedule as time string
	ShutdownOverride string `json:"shutdownOverride,omitempty"`
}
