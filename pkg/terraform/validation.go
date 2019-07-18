package terraform

import "fmt"

// ValidationError represents a validation error of Terraform.
type ValidationError struct {
	Diagnostic
}

// Error returns an error with validation error message.
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s (%s line %d)", ve.Severity, ve.Summary, ve.Range.Filename, ve.Range.Start.Line)
}

// ValidationResult represents a result of 'terraform validate' command.
type ValidationResult struct {
	Valid        bool         `json:"valid"`
	ErrorCount   int          `json:"error_count"`
	WarningCount int          `json:"warning_count"`
	Diagnostics  []Diagnostic `json:"diagnostics"`
}

// Error returns a ValidationError.
func (r *ValidationResult) Error() error {
	if r.Valid {
		return nil
	}

	return &ValidationError{r.Diagnostics[0]}
}

type Diagnostic struct {
	Severity string          `json:"severity"`
	Summary  string          `json:"summary"`
	Detail   string          `json:"detail"`
	Range    DiagnosticRange `json:"range"`
}

type DiagnosticRange struct {
	Filename string             `json:"filename"`
	Start    DiagnosticRangePos `json:"start"`
	End      DiagnosticRangePos `json:"end"`
}

type DiagnosticRangePos struct {
	Line   int `json:"line"`
	Column int `json:"start"`
	Byte   int `json:"byte"`
}
