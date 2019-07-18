package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/summerwind/terraform-controller/pkg/api"
	"github.com/summerwind/terraform-controller/pkg/api/v1alpha1"
)

func main() {
	log.SetOutput(os.Stderr)

	cmd := &cobra.Command{
		Use:           "terraform-controller",
		Short:         "Manage custom resource for Terraform",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmdRun := &cobra.Command{
		Use:   "run",
		Short: "Manage Run resource",
	}
	cmd.AddCommand(cmdRun)

	cmdRunReconcile := &cobra.Command{
		Use:   "reconcile",
		Short: "Reconcile resource",
		RunE:  reconcileRun,
	}
	cmdRun.AddCommand(cmdRunReconcile)

	cmdRunFinalize := &cobra.Command{
		Use:   "finalize",
		Short: "Finalize resource",
		RunE:  finalizeRun,
	}
	cmdRun.AddCommand(cmdRunFinalize)

	cmdRunValidate := &cobra.Command{
		Use:   "validate",
		Short: "Validate resource",
		RunE:  validateRun,
	}
	cmdRun.AddCommand(cmdRunValidate)

	err := cmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}

func reconcileRun(cmd *cobra.Command, args []string) error {
	state := &v1alpha1.RunState{}

	err := json.NewDecoder(os.Stdin).Decode(state)
	if err != nil {
		return err
	}

	state, err = api.ReconcileRun(state, false)
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(state)
}

func finalizeRun(cmd *cobra.Command, args []string) error {
	state := &v1alpha1.RunState{}

	err := json.NewDecoder(os.Stdin).Decode(state)
	if err != nil {
		return err
	}

	state, err = api.ReconcileRun(state, true)
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(state)
}

func validateRun(cmd *cobra.Command, args []string) error {
	req := &admissionv1beta1.AdmissionRequest{}

	err := json.NewDecoder(os.Stdin).Decode(req)
	if err != nil {
		return err
	}

	c := &v1alpha1.Run{}
	err = json.Unmarshal(req.Object.Raw, c)
	if err != nil {
		return err
	}

	res := &admissionv1beta1.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
	}

	err = c.Validate()
	if err != nil {
		res.Allowed = false
		res.Result = &metav1.Status{
			Status: "Failure",
			Reason: metav1.StatusReason(err.Error()),
		}
	}

	return json.NewEncoder(os.Stdout).Encode(res)
}
