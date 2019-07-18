package terraform

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/summerwind/terraform-controller/pkg/exec"
)

var errMessagePrefix = regexp.MustCompile(`Error: ([^\n]+)`)

// Terraform represents terraform command.
type Terraform struct {
	exec.Command
}

// New returns a Terraform.
func New(dir string) *Terraform {
	return &Terraform{
		exec.Command{
			Name:       "terraform",
			Path:       "terraform",
			WorkingDir: dir,
			Debug:      false,
		},
	}
}

// Init executes 'terraform init' on current directory.
func (tf *Terraform) Init() error {
	result, err := tf.Run("init", "-no-color", "-input=false")
	if err != nil {
		tferr := parseError(result.Stderr)
		if tferr != nil {
			return tferr
		}
		return fmt.Errorf("initialization failed: %v", err)
	}

	return nil
}

// Validate executes 'terraform validate' on current directory.
func (tf *Terraform) Validate() error {
	result, err := tf.Run("validate", "-json")
	if result.ExitCode != 1 && err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	r := &ValidationResult{}
	err = json.Unmarshal(result.Stdout, r)
	if err != nil {
		return fmt.Errorf("invalid validation result: %v", err)
	}

	return r.Error()
}

// SelectWorkspace select workspace of terraform.
func (tf *Terraform) SelectWorkspace(workspace string) error {
	if workspace != "" {
		// Always try to create specified workspace.
		tf.Run("workspace", "new", workspace, "-no-color")

		result, err := tf.Run("workspace", "select", workspace, "-no-color")
		if err != nil {
			tferr := parseError(result.Stderr)
			if tferr != nil {
				return tferr
			}
			return fmt.Errorf("workspace '%s' select failed: %v", workspace, err)
		}
	}

	return nil
}

// Plan executes 'terraform plan' with specified vars.
func (tf *Terraform) Plan(vars map[string]string) (bool, error) {
	args := []string{"plan", "-detailed-exitcode", "-no-color", "-input=false"}
	for k, v := range vars {
		args = append(args, "-var", fmt.Sprintf("%s=%s", k, v))
	}

	result, err := tf.Run(args...)
	if result.ExitCode != 2 && err != nil {
		tferr := parseError(result.Stderr)
		if tferr != nil {
			return false, tferr
		}
		return false, fmt.Errorf("plan failed: %v", err)
	}

	// The resource has been changed if exit code is 2.
	return (result.ExitCode == 2), nil
}

// Apply executes 'terraform apply' with specified vars.
func (tf *Terraform) Apply(vars map[string]string) error {
	args := []string{"apply", "-auto-approve", "-no-color", "-input=false"}
	for k, v := range vars {
		args = append(args, "-var", fmt.Sprintf("%s=%s", k, v))
	}

	result, err := tf.Run(args...)
	if err != nil {
		tferr := parseError(result.Stderr)
		if tferr != nil {
			return tferr
		}
		return fmt.Errorf("apply failed: %v", err)
	}

	return nil
}

// Destroy executes 'terraform destroy' with specified vars.
func (tf *Terraform) Destroy(vars map[string]string) error {
	args := []string{"destroy", "-auto-approve", "-no-color"}
	for k, v := range vars {
		args = append(args, "-var", fmt.Sprintf("%s=%s", k, v))
	}

	result, err := tf.Run(args...)
	if err != nil {
		tferr := parseError(result.Stderr)
		if tferr != nil {
			return tferr
		}
		return fmt.Errorf("destroy failed: %v", err)
	}

	return nil
}

// parseError returns a error with terraform error message.
func parseError(buf []byte) error {
	matched := errMessagePrefix.FindSubmatch(buf)
	if len(matched) > 0 {
		return errors.New(string(matched[1]))
	}

	return nil
}
