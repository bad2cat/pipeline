/*
Copyright 2019 The Tekton Authors

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
	"fmt"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

// TaskRunSpec defines the desired state of TaskRun
type TaskRunSpec struct {
	// +optional
	ServiceAccountName string `json:"serviceAccountName"`
	// no more than one of the TaskRef and TaskSpec may be specified.
	// +optional
	TaskRef *TaskRef `json:"taskRef,omitempty"`
	// +optional
	TaskSpec *TaskSpec `json:"taskSpec,omitempty"`
	// Used for cancelling a taskrun (and maybe more later on)
	// +optional
	Status TaskRunSpecStatus `json:"status,omitempty"`
	// Time after which the build times out. Defaults to 10 minutes.
	// Specified build timeout should be less than 24h.
	// Refer Go's ParseDuration documentation for expected format: https://golang.org/pkg/time/#ParseDuration
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// PodTemplate holds pod specific configuration
	// +optional
	PodTemplate *PodTemplate `json:"podTemplate,omitempty"`
	// Workspaces is a list of WorkspaceBindings from volumes to workspaces.
	// +optional
	Workspaces []WorkspaceBinding `json:"workspaces,omitempty"`
	// From v1alpha2
	// +optional
	Params []Param `json:"params,omitempty"`
	// +optional
	Resources *v1alpha2.TaskRunResources `json:"resources,omitempty"`
	// Deprecated
	// +optional
	Inputs TaskRunInputs `json:"inputs,omitempty"`
	// +optional
	Outputs TaskRunOutputs `json:"outputs,omitempty"`
}

// TaskRunSpecStatus defines the taskrun spec status the user can provide
type TaskRunSpecStatus = v1alpha2.TaskRunSpecStatus

const (
	// TaskRunSpecStatusCancelled indicates that the user wants to cancel the task,
	// if not already cancelled or terminated
	TaskRunSpecStatusCancelled = v1alpha2.TaskRunSpecStatusCancelled
)

// TaskRunInputs holds the input values that this task was invoked with.
type TaskRunInputs struct {
	// +optional
	Resources []TaskResourceBinding `json:"resources,omitempty"`
	// +optional
	Params []Param `json:"params,omitempty"`
}

// TaskResourceBinding points to the PipelineResource that
// will be used for the Task input or output called Name.
type TaskResourceBinding struct {
	PipelineResourceBinding
	// Paths will probably be removed in #1284, and then PipelineResourceBinding can be used instead.
	// The optional Path field corresponds to a path on disk at which the Resource can be found
	// (used when providing the resource via mounted volume, overriding the default logic to fetch the Resource).
	// +optional
	Paths []string `json:"paths,omitempty"`
}

// TaskRunOutputs holds the output values that this task was invoked with.
type TaskRunOutputs struct {
	// +optional
	Resources []TaskResourceBinding `json:"resources,omitempty"`
}

// TaskRunStatus defines the observed state of TaskRun
type TaskRunStatus = v1alpha2.TaskRunStatus

// TaskRunStatusFields holds the fields of TaskRun's status.  This is defined
// separately and inlined so that other types can readily consume these fields
// via duck typing.
type TaskRunStatusFields = v1alpha2.TaskRunStatusFields

// TaskRunResult used to describe the results of a task
type TaskRunResult = v1alpha2.TaskRunResult

// StepState reports the results of running a step in the Task.
type StepState = v1alpha2.StepState

// SidecarState reports the results of sidecar in the Task.
type SidecarState = v1alpha2.SidecarState

// CloudEventDelivery is the target of a cloud event along with the state of
// delivery.
type CloudEventDelivery = v1alpha2.CloudEventDelivery

// CloudEventCondition is a string that represents the condition of the event.
type CloudEventCondition = v1alpha2.CloudEventCondition

const (
	// CloudEventConditionUnknown means that the condition for the event to be
	// triggered was not met yet, or we don't know the state yet.
	CloudEventConditionUnknown CloudEventCondition = v1alpha2.CloudEventConditionUnknown
	// CloudEventConditionSent means that the event was sent successfully
	CloudEventConditionSent CloudEventCondition = v1alpha2.CloudEventConditionSent
	// CloudEventConditionFailed means that there was one or more attempts to
	// send the event, and none was successful so far.
	CloudEventConditionFailed CloudEventCondition = v1alpha2.CloudEventConditionFailed
)

// CloudEventDeliveryState reports the state of a cloud event to be sent.
type CloudEventDeliveryState = v1alpha2.CloudEventDeliveryState

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TaskRun represents a single execution of a Task. TaskRuns are how the steps
// specified in a Task are executed; they specify the parameters and resources
// used to run the steps in a Task.
//
// +k8s:openapi-gen=true
type TaskRun struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec TaskRunSpec `json:"spec,omitempty"`
	// +optional
	Status TaskRunStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TaskRunList contains a list of TaskRun
type TaskRunList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TaskRun `json:"items"`
}

// GetBuildPodRef for task
func (tr *TaskRun) GetBuildPodRef() corev1.ObjectReference {
	return corev1.ObjectReference{
		APIVersion: "v1",
		Kind:       "Pod",
		Namespace:  tr.Namespace,
		Name:       tr.Name,
	}
}

// GetPipelineRunPVCName for taskrun gets pipelinerun
func (tr *TaskRun) GetPipelineRunPVCName() string {
	if tr == nil {
		return ""
	}
	for _, ref := range tr.GetOwnerReferences() {
		if ref.Kind == pipeline.PipelineRunControllerName {
			return fmt.Sprintf("%s-pvc", ref.Name)
		}
	}
	return ""
}

// HasPipelineRunOwnerReference returns true of TaskRun has
// owner reference of type PipelineRun
func (tr *TaskRun) HasPipelineRunOwnerReference() bool {
	for _, ref := range tr.GetOwnerReferences() {
		if ref.Kind == pipeline.PipelineRunControllerName {
			return true
		}
	}
	return false
}

// IsDone returns true if the TaskRun's status indicates that it is done.
func (tr *TaskRun) IsDone() bool {
	return !tr.Status.GetCondition(apis.ConditionSucceeded).IsUnknown()
}

// HasStarted function check whether taskrun has valid start time set in its status
func (tr *TaskRun) HasStarted() bool {
	return tr.Status.StartTime != nil && !tr.Status.StartTime.IsZero()
}

// IsSuccessful returns true if the TaskRun's status indicates that it is done.
func (tr *TaskRun) IsSuccessful() bool {
	return tr.Status.GetCondition(apis.ConditionSucceeded).IsTrue()
}

// IsCancelled returns true if the TaskRun's spec status is set to Cancelled state
func (tr *TaskRun) IsCancelled() bool {
	return tr.Spec.Status == TaskRunSpecStatusCancelled
}

// GetRunKey return the taskrun key for timeout handler map
func (tr *TaskRun) GetRunKey() string {
	// The address of the pointer is a threadsafe unique identifier for the taskrun
	return fmt.Sprintf("%s/%p", "TaskRun", tr)
}

// IsPartOfPipeline return true if TaskRun is a part of a Pipeline.
// It also return the name of Pipeline and PipelineRun
func (tr *TaskRun) IsPartOfPipeline() (bool, string, string) {
	if tr == nil || len(tr.Labels) == 0 {
		return false, "", ""
	}

	if pl, ok := tr.Labels[pipeline.GroupName+pipeline.PipelineLabelKey]; ok {
		return true, pl, tr.Labels[pipeline.GroupName+pipeline.PipelineRunLabelKey]
	}

	return false, "", ""
}
