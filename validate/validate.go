// Package validate checks a BitriseDataModel for structural inconsistencies.
// It surfaces problems that would cause a build to fail at runtime — missing workflow
// references, broken dependency chains, etc. — before the config is uploaded.
//
// Two validation levels are available:
//
//   - Config: fast SDK-level structural checks (missing refs, broken depends_on, etc.)
//   - Full: Config + upstream bitrise/v2 Normalize + Validate (cycle detection, step format,
//     DAG validation, deprecated-step warnings, and more)
package validate

import (
	"fmt"
	"strings"

	bitriseModels "github.com/bitrise-io/bitrise/v2/models"
)

// Error represents a single validation problem.
type Error struct {
	// Location is a human-readable path to the offending element (e.g. "workflow.primary.before_run[0]").
	Location string
	// Message describes what is wrong.
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Location, e.Message)
}

// Result holds the combined output of a Full validation run.
type Result struct {
	// Errors are problems that will prevent the config from running correctly.
	// Includes both SDK structural errors and errors from upstream bitrise/v2 validation.
	Errors []Error
	// Warnings are non-fatal advisories from upstream bitrise/v2 validation
	// (e.g. deprecated steps, unused outputs).
	Warnings []string
}

// IsValid reports whether the config has no errors (warnings are allowed).
func (r Result) IsValid() bool { return len(r.Errors) == 0 }

// Config runs SDK-level structural checks and returns any detected errors.
// It is fast and has no side effects on data.
// For the most thorough validation, use Full instead.
func Config(data bitriseModels.BitriseDataModel) []Error {
	var errs []Error

	errs = append(errs, validateWorkflowRefs(data)...)
	errs = append(errs, validateGraphPipelineDeps(data)...)
	errs = append(errs, validateContainerRefs(data)...)
	errs = append(errs, validateStepBundleRefs(data)...)

	return errs
}

// Full runs the complete validation pipeline on a copy of data:
//
//  1. SDK structural checks (Config)
//  2. Upstream bitrise/v2/models Normalize (normalizes TriggerMap, containers, step bundles, workflows)
//  3. Upstream bitrise/v2/models Validate (cycle detection, step format, DAG validation, trigger map, etc.)
//
// The returned Result combines SDK errors with upstream errors and warnings.
// The returned error is non-nil only for unexpected failures during normalization, not for
// validation failures — check Result.IsValid() to determine whether the config is valid.
func Full(data bitriseModels.BitriseDataModel) (Result, error) {
	result := Result{
		Errors: Config(data),
	}

	if err := data.Normalize(); err != nil {
		return result, fmt.Errorf("normalize: %w", err)
	}

	warnings, err := data.Validate()
	result.Warnings = warnings
	if err != nil {
		result.Errors = append(result.Errors, Error{
			Location: "config",
			Message:  err.Error(),
		})
	}

	return result, nil
}

// validateWorkflowRefs checks that every workflow ID referenced in before_run / after_run
// points to a workflow that exists in the config.
func validateWorkflowRefs(data bitriseModels.BitriseDataModel) []Error {
	var errs []Error
	for id, wf := range data.Workflows {
		for i, ref := range wf.BeforeRun {
			if _, ok := data.Workflows[ref]; !ok {
				errs = append(errs, Error{
					Location: fmt.Sprintf("workflow.%s.before_run[%d]", id, i),
					Message:  fmt.Sprintf("references unknown workflow %q", ref),
				})
			}
		}
		for i, ref := range wf.AfterRun {
			if _, ok := data.Workflows[ref]; !ok {
				errs = append(errs, Error{
					Location: fmt.Sprintf("workflow.%s.after_run[%d]", id, i),
					Message:  fmt.Sprintf("references unknown workflow %q", ref),
				})
			}
		}
	}
	return errs
}

// validateGraphPipelineDeps checks that every workflow ID listed in a graph pipeline
// workflow's depends_on exists within the same pipeline.
func validateGraphPipelineDeps(data bitriseModels.BitriseDataModel) []Error {
	var errs []Error
	for pipelineID, pipeline := range data.Pipelines {
		for wfID, wf := range pipeline.Workflows {
			for i, dep := range wf.DependsOn {
				if _, ok := pipeline.Workflows[dep]; !ok {
					errs = append(errs, Error{
						Location: fmt.Sprintf("pipeline.%s.workflows.%s.depends_on[%d]", pipelineID, wfID, i),
						Message:  fmt.Sprintf("references unknown workflow %q within the pipeline", dep),
					})
				}
			}
		}
	}
	return errs
}

// validateContainerRefs checks that container IDs referenced in "with" groups exist.
func validateContainerRefs(data bitriseModels.BitriseDataModel) []Error {
	var errs []Error
	for wfID, wf := range data.Workflows {
		for stepIdx, stepItem := range wf.Steps {
			withRaw, ok := stepItem[bitriseModels.StepListItemWithKey]
			if !ok {
				continue
			}
			withModel, ok := withRaw.(bitriseModels.WithModel)
			if !ok {
				continue
			}
			loc := fmt.Sprintf("workflow.%s.steps[%d].with", wfID, stepIdx)
			if withModel.ContainerID != "" {
				if _, exists := data.Containers[withModel.ContainerID]; !exists {
					errs = append(errs, Error{
						Location: loc + ".container",
						Message:  fmt.Sprintf("references unknown container %q", withModel.ContainerID),
					})
				}
			}
			for i, svcID := range withModel.ServiceIDs {
				if _, exists := data.Containers[svcID]; !exists {
					errs = append(errs, Error{
						Location: fmt.Sprintf("%s.services[%d]", loc, i),
						Message:  fmt.Sprintf("references unknown container %q", svcID),
					})
				}
			}
		}
	}
	return errs
}

// validateStepBundleRefs checks that step bundle IDs referenced in workflows are defined
// in the top-level step_bundles map.
func validateStepBundleRefs(data bitriseModels.BitriseDataModel) []Error {
	var errs []Error
	for wfID, wf := range data.Workflows {
		for stepIdx, stepItem := range wf.Steps {
			for key := range stepItem {
				if !strings.HasPrefix(key, bitriseModels.StepListItemStepBundleKeyPrefix) {
					continue
				}
				bundleID := strings.TrimPrefix(key, bitriseModels.StepListItemStepBundleKeyPrefix)
				if _, ok := data.StepBundles[bundleID]; !ok {
					errs = append(errs, Error{
						Location: fmt.Sprintf("workflow.%s.steps[%d]", wfID, stepIdx),
						Message:  fmt.Sprintf("references undefined step bundle %q", bundleID),
					})
				}
			}
		}
	}
	return errs
}
