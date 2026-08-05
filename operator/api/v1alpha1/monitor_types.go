/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// MonitorSpec defines the desired state of Monitor
type MonitorSpec struct {
	// URL is the target endpoint to monitor
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^https?://.*`
	URL string `json:"url"`

	// IntervalSeconds is how often to run the check (minimum 30s)
	// +kubebuilder:validation:Minimum=30
	// +kubebuilder:default=60
	IntervalSeconds int32 `json:"intervalSeconds,omitempty"`

	// ExpectedStatus is the HTTP status code considered healthy
	// +kubebuilder:validation:Minimum=100
	// +kubebuilder:validation:Maximum=599
	// +kubebuilder:default=200
	ExpectedStatus int32 `json:"expectedStatus,omitempty"`

	// TimeoutSeconds is the HTTP request timeout
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=30
	// +kubebuilder:default=10
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty"`
}

// MonitorResult represents the outcome of a single check
// +kubebuilder:validation:Enum=UP;DOWN;UNKNOWN
type MonitorResult string

const (
	MonitorResultUp      MonitorResult = "UP"
	MonitorResultDown    MonitorResult = "DOWN"
	MonitorResultUnknown MonitorResult = "UNKNOWN"
)

// MonitorStatus defines the observed state of Monitor
type MonitorStatus struct {
	// LastCheck is the timestamp of the most recent check
	// +optional
	LastCheck *metav1.Time `json:"lastCheck,omitempty"`

	// LastResult is the outcome of the last check: UP, DOWN, or UNKNOWN
	// +optional
	LastResult MonitorResult `json:"lastResult,omitempty"`

	// LastStatusCode is the HTTP status code returned by the last check
	// +optional
	LastStatusCode int32 `json:"lastStatusCode,omitempty"`

	// LastError contains the error message if the last check failed
	// +optional
	LastError string `json:"lastError,omitempty"`

	// UptimePercent is the rolling 24h uptime percentage, stored as a string e.g. "99.50"
	// +optional
	UptimePercent string `json:"uptimePercent,omitempty"`

	// CheckJobName is the name of the CronJob managing this monitor's checks
	// +optional
	CheckJobName string `json:"checkJobName,omitempty"`

	// Conditions represent the latest available observations of the Monitor's state
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.spec.url`
// +kubebuilder:printcolumn:name="Interval",type=integer,JSONPath=`.spec.intervalSeconds`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.lastResult`
// +kubebuilder:printcolumn:name="Last Check",type=date,JSONPath=`.status.lastCheck`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Monitor is the Schema for the monitors API
type Monitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MonitorSpec   `json:"spec,omitempty"`
	Status MonitorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MonitorList contains a list of Monitor
type MonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Monitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(SchemeGroupVersion, &Monitor{}, &MonitorList{})
		return nil
	})
}
