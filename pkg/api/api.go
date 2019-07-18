package api

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/summerwind/terraform-controller/pkg/api/v1alpha1"
	"github.com/summerwind/terraform-controller/pkg/git"
	"github.com/summerwind/terraform-controller/pkg/terraform"
)

const (
	ReasonInvalid         v1alpha1.RunStatusReason = "Invalid"
	ReasonGitError        v1alpha1.RunStatusReason = "GitError"
	ReasonTerraformError  v1alpha1.RunStatusReason = "TerraformError"
	ReasonValidationError v1alpha1.RunStatusReason = "ValidationError"
	ReasonPlanFailed      v1alpha1.RunStatusReason = "PlanFailed"
	ReasonApplySucceeded  v1alpha1.RunStatusReason = "ApplySucceeded"
	ReasonApplyFailed     v1alpha1.RunStatusReason = "ApplyFailed"
	ReasonDestroyFailed   v1alpha1.RunStatusReason = "DestroyFailed"
)

var (
	debug bool = (os.Getenv("TF_CONTROLLER_DEBUG") != "")
)

func ReconcileRun(state *v1alpha1.RunState, finalize bool) (*v1alpha1.RunState, error) {
	var (
		err        error
		workingDir string
		checksum   string
	)

	dir, err := ioutil.TempDir("", "terraform-controller")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	workingDir = dir

	r := state.Object
	err = r.Validate()
	if err != nil {
		r.Status.Fail(ReasonInvalid, err.Error())
		return state, nil
	}

	if r.Spec.Source != nil {
		var checkout bool

		if r.Spec.Source.Git != nil {
			g := git.New(dir)
			g.Debug = debug

			commit, err := g.Checkout(r.Spec.Source.Git.URL, r.Spec.Source.Git.Revision)
			if err != nil {
				r.Status.Fail(ReasonGitError, err.Error())
				return state, nil
			}

			checkout = true
			checksum = commit
			log(r, "Checked out source from git: %s@%s", r.Spec.Source.Git.URL, r.Spec.Source.Git.Revision)
		}

		if !checkout {
			r.Status.Fail(ReasonInvalid, "source is not specified")
			return state, nil
		}

		if r.Spec.Source.WorkingDir != "" {
			workingDir = filepath.Join(dir, r.Spec.Source.WorkingDir)
			log(r, "Set '%s' to working directory", r.Spec.Source.WorkingDir)
		}
	} else {
		err := ioutil.WriteFile(filepath.Join(dir, "main.tf"), []byte(r.Spec.Content), 0644)
		if err != nil {
			return nil, err
		}

		checksum = fmt.Sprintf("%x", sha256.Sum256([]byte(r.Spec.Content)))
		log(r, "Generated configuration file with specified content")
	}

	tf := terraform.New(workingDir)
	tf.Debug = debug

	err = tf.Init()
	if err != nil {
		r.Status.Fail(ReasonTerraformError, err.Error())
		return state, nil
	}
	log(r, "Initialized")

	if r.Spec.Workspace != "" {
		err = tf.SelectWorkspace(r.Spec.Workspace)
		if err != nil {
			r.Status.Fail(ReasonTerraformError, err.Error())
			return state, nil
		}
		log(r, "Workspace changed to '%s'", r.Spec.Workspace)
	}

	err = tf.Validate()
	if err != nil {
		r.Status.Fail(ReasonValidationError, err.Error())
		return state, err
	}
	log(r, "Validation succeeded")

	if finalize {
		if r.Spec.Destroy {
			err := tf.Destroy(r.Spec.Vars)
			if err != nil {
				r.Status.Fail(ReasonDestroyFailed, err.Error())
				state.Requeue = true
				return state, nil
			}

			log(r, "Destroy succeeded")
		}
	} else {
		changed, err := tf.Plan(r.Spec.Vars)
		if err != nil {
			r.Status.Fail(ReasonPlanFailed, err.Error())
			return state, nil
		}
		log(r, "Plan succeeded")

		if changed {
			log(r, "Changes has been detected")

			err := tf.Apply(r.Spec.Vars)
			if err != nil {
				r.Status.Fail(ReasonApplyFailed, err.Error())
				return state, nil
			}

			log(r, "Apply succeeded")

			r.Status.Success(ReasonApplySucceeded, "")
			r.Status.LastAppliedTime = metav1.Now()
			r.Status.LastAppliedChecksum = checksum
		} else {
			log(r, "Resource is up to date")
		}
	}

	return state, nil
}

func log(c *v1alpha1.Run, format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "[%s/%s] %s\n", c.Namespace, c.Name, fmt.Sprintf(format, a...))
}
