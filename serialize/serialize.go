// Package serialize handles YAML and JSON serialization of BitriseDataModel.
//
// Two tiers of output functions are provided:
//
//   - ToYAML / ToJSON / Print — serialize as-is, no validation
//   - ValidatedToYAML / ValidatedToJSON / ValidatedPrint — run Full validation first,
//     print any warnings to stderr, and return an error if the config is invalid
package serialize

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	bitriseModels "github.com/bitrise-io/bitrise/v2/models"
	"gopkg.in/yaml.v2"

	"github.com/bitrise-io/bitrise-pipeline-sdk/validate"
)

// ToYAML serializes a BitriseDataModel to YAML bytes.
func ToYAML(data bitriseModels.BitriseDataModel) ([]byte, error) {
	return yaml.Marshal(data)
}

// ToJSON serializes a BitriseDataModel to indented JSON bytes.
func ToJSON(data bitriseModels.BitriseDataModel) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

// Print writes the YAML serialization of data to stdout.
// This is the primary output mechanism for dynamic pipeline scripts —
// pipe the output to `bitrise run --config -`.
func Print(data bitriseModels.BitriseDataModel) error {
	out, err := ToYAML(data)
	if err != nil {
		return fmt.Errorf("serialize: %w", err)
	}
	_, err = os.Stdout.Write(out)
	return err
}

// PrintJSON writes the JSON serialization of data to stdout.
func PrintJSON(data bitriseModels.BitriseDataModel) error {
	out, err := ToJSON(data)
	if err != nil {
		return fmt.Errorf("serialize: %w", err)
	}
	_, err = fmt.Fprintln(os.Stdout, string(out))
	return err
}

// Normalize returns a normalized copy of data by applying upstream
// bitrise/v2/models normalization (TriggerMap, containers, step bundles,
// workflows, meta). The original is not modified.
func Normalize(data bitriseModels.BitriseDataModel) (bitriseModels.BitriseDataModel, error) {
	if err := data.Normalize(); err != nil {
		return bitriseModels.BitriseDataModel{}, fmt.Errorf("normalize: %w", err)
	}
	return data, nil
}

// FillMissingDefaults fills default values on a copy of data (e.g. default
// step timeouts, is_always_run defaults) using bitrise/v2/models defaults.
// The original is not modified.
func FillMissingDefaults(data bitriseModels.BitriseDataModel) (bitriseModels.BitriseDataModel, error) {
	if err := data.FillMissingDefaults(); err != nil {
		return bitriseModels.BitriseDataModel{}, fmt.Errorf("fill defaults: %w", err)
	}
	return data, nil
}

// ValidatedToYAML runs Full validation on data, then serializes to YAML.
// Returns an error if validation fails. Warnings are written to stderr.
func ValidatedToYAML(data bitriseModels.BitriseDataModel) ([]byte, error) {
	if err := runValidation(&data); err != nil {
		return nil, err
	}
	return ToYAML(data)
}

// ValidatedToJSON runs Full validation on data, then serializes to JSON.
// Returns an error if validation fails. Warnings are written to stderr.
func ValidatedToJSON(data bitriseModels.BitriseDataModel) ([]byte, error) {
	if err := runValidation(&data); err != nil {
		return nil, err
	}
	return ToJSON(data)
}

// ValidatedPrint runs Full validation, prints warnings to stderr, then writes
// the normalized config as YAML to stdout. Returns an error if the config is invalid.
//
// This is the recommended output function for dynamic pipeline scripts:
//
//	go run ./pipeline-gen/main.go | bitrise run --config -
func ValidatedPrint(data bitriseModels.BitriseDataModel) error {
	if err := runValidation(&data); err != nil {
		return err
	}
	return Print(data)
}

// runValidation performs Full validation on *data (normalized in place on success)
// and writes warnings to stderr. Returns a formatted error listing all problems.
func runValidation(data *bitriseModels.BitriseDataModel) error {
	result, err := validate.Full(*data)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	for _, w := range result.Warnings {
		fmt.Fprintf(os.Stderr, "WARNING: %s\n", w)
	}

	if !result.IsValid() {
		msgs := make([]string, len(result.Errors))
		for i, e := range result.Errors {
			msgs[i] = "  - " + e.Error()
		}
		return fmt.Errorf("config has %d validation error(s):\n%s",
			len(result.Errors), strings.Join(msgs, "\n"))
	}

	// Normalize the original so the caller gets the normalized form.
	if err := data.Normalize(); err != nil {
		return fmt.Errorf("normalize: %w", err)
	}
	return nil
}
