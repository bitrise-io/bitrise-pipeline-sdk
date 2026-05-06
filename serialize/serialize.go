// Package serialize handles YAML and JSON serialization of BitriseDataModel.
package serialize

import (
	"encoding/json"
	"fmt"
	"os"

	bitriseModels "github.com/bitrise-io/bitrise/v2/models"
	"gopkg.in/yaml.v2"
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
