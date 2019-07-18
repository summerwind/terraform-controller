package v1alpha1

import (
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Run struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunSpec   `json:"spec,omitempty"`
	Status RunStatus `json:"status,omitempty"`
}

func (r *Run) Validate() error {
	return r.Spec.Validate()
}

type RunSpec struct {
	Workspace string            `json:"workspace"`
	Vars      map[string]string `json:"vars"`
	Content   string            `json:"content"`
	Source    *RunSpecSource    `json:"source"`
	Destroy   bool              `json:"destroy"`
}

func (r *RunSpec) Validate() error {
	if r.Content == "" && r.Source == nil {
		return errors.New("one of content or source must be specified")
	}

	if r.Source != nil {
		return r.Source.Validate()
	}

	return nil
}

type RunSpecSource struct {
	Git        *RunSpecSourceGit `json:"git"`
	WorkingDir string            `json:"workingDir"`
}

func (r *RunSpecSource) Validate() error {
	if r.Git == nil {
		return errors.New("at least one source type must be specified")
	}

	return r.Git.Validate()
}

type RunSpecSourceGit struct {
	URL      string `json:"url"`
	Revision string `json:"revision"`
}

func (r *RunSpecSourceGit) Validate() error {
	return nil
}

type RunStatus struct {
	Phase               RunStatusPhase
	Reason              RunStatusReason
	Message             string
	LastAppliedTime     metav1.Time
	LastAppliedChecksum string
}

func (r *RunStatus) Success(reason RunStatusReason, msg string) {
	r.Phase = RunStatusPhaseSucceeded
	r.Reason = reason
	r.Message = msg
}

func (r *RunStatus) Fail(reason RunStatusReason, msg string) {
	r.Phase = RunStatusPhaseFailed
	r.Reason = reason
	r.Message = msg
}

type RunStatusPhase string

const (
	RunStatusPhasePending   RunStatusPhase = "Pending"
	RunStatusPhaseSucceeded RunStatusPhase = "Succeeded"
	RunStatusPhaseFailed    RunStatusPhase = "Failed"
)

type RunStatusReason string

type RunState struct {
	Object       *Run `json:"object"`
	Requeue      bool `json:"requeue,omitempty"`
	RequeueAfter int  `json:"requeueAfter,omitempty"`
}
