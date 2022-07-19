package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// Shutdown defines schedule in cron format for stand shutdown.
	Shutdown string `json:"shutdown"`
}

type ResourcesSpec struct {
	// Azure contains an array of related azure resources.
	Azure []AzureResource `json:"azure"`
}

type AzureResource struct {
	// Type defines one of supported azure resource types.
	Type AzureResourceType `json:"type"`

	// ResourceGroupName defines resource group name for resource.
	ResourceGroupName string `json:"resourceGroupName"`

	// ResourceNameFilter defines regex filter for resource.
	ResourceNameFilter string `json:"resourceNameFilter"`

	// Priority specifies order in which resources will be started or shutdowned.
	Priority int64 `json:"priority"`
}

type AzureResourceType string

const (
	AzureResourceManagedMySQL   AzureResourceType = "mysql"
	AzureResourceVirtualMachine AzureResourceType = "vm"
)

// StandSchedulePolicyStatus is a status for StandSchedulePolicy resource.
type StandSchedulePolicyStatus struct {
	// Conditions defines current service state of policy.
	Conditions []PolicyStatusCondition
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
