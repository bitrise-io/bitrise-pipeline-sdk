# bitrise-pipeline-sdk

A Go SDK for generating [Bitrise](https://bitrise.io) pipeline configurations programmatically — the Bitrise equivalent of [Buildkite Dynamic Pipelines](https://buildkite.com/docs/pipelines/configure/dynamic-pipelines).

Write your pipeline as a Go program, run it, and pipe the YAML output directly into the Bitrise CLI:

```sh
go run ./pipeline.go | bitrise run --config -
```

## Features

- **Fluent builder API** for workflows, graph pipelines, steps, triggers, containers, and step bundles
- **Typed step builders** for common steps (Xcode, Android, Script, Git Clone, …) with IDE auto-complete
- **Built-in validation** — catches broken `before_run` references, missing container IDs, and more before the config leaves your machine
- **`bitrise-gen` CLI** — run, validate, or scaffold pipeline scripts without writing any boilerplate

## Installation

```sh
go get github.com/bitrise-io/bitrise-pipeline-sdk
```

Install the `bitrise-gen` CLI:

```sh
go install github.com/bitrise-io/bitrise-pipeline-sdk/cmd/bitrise-gen@latest
```

## Quick start

Scaffold a starter pipeline script:

```sh
bitrise-gen scaffold          # writes pipeline.go in the current directory
bitrise-gen scaffold --output=ci/pipeline.go
```

Run it and print the YAML:

```sh
bitrise-gen run ./pipeline.go
```

Validate without printing:

```sh
bitrise-gen validate ./pipeline.go
```

Pipe directly into the Bitrise CLI:

```sh
bitrise-gen run ./pipeline.go | bitrise run --config -
```

## SDK usage

```go
package main

import (
    "log"

    "github.com/bitrise-io/bitrise-pipeline-sdk/graphpipeline"
    "github.com/bitrise-io/bitrise-pipeline-sdk/pipeline"
    "github.com/bitrise-io/bitrise-pipeline-sdk/serialize"
    "github.com/bitrise-io/bitrise-pipeline-sdk/step"
    "github.com/bitrise-io/bitrise-pipeline-sdk/trigger"
    "github.com/bitrise-io/bitrise-pipeline-sdk/workflow"
)

func main() {
    setup := workflow.New().
        AddStep(step.ActivateSshKey()).
        AddStep(step.GitClone()).
        AddStep(step.CachePull())

    test := workflow.New().
        WithBeforeRun("setup").
        AddStep(step.Script().WithContent("go test ./..."))

    deploy := workflow.New().
        WithBeforeRun("setup").
        AddStep(step.Script().WithContent("go build -o bin/app ./cmd/app")).
        AddStep(step.DeployToBitriseIo())

    ci := graphpipeline.New().
        AddWorkflow("setup", graphpipeline.NewWorkflow()).
        AddWorkflow("test", graphpipeline.NewWorkflow().WithDependsOn("setup")).
        AddWorkflow("deploy", graphpipeline.NewWorkflow().WithDependsOn("test").WithAbortOnFail(true))

    cfg := pipeline.New("other").
        AddWorkflow("setup", setup).
        AddWorkflow("test", test).
        AddWorkflow("deploy", deploy).
        AddGraphPipeline("ci", ci).
        AddTrigger(trigger.OnPush("", "ci").WithBranch("main").Build()).
        AddTrigger(trigger.OnTag("deploy", "").WithTag("v*").Build())

    if err := serialize.ValidatedPrint(cfg.Build()); err != nil {
        log.Fatal(err)
    }
}
```

## Packages

| Package | Purpose |
|---|---|
| `pipeline` | Root config builder — entry point for every pipeline script |
| `workflow` | Workflow builder with steps, env vars, before/after run |
| `graphpipeline` | Graph pipeline builder with dependency edges |
| `step` | Generic step builder (`From`) and typed builders (Xcode, Android, Script, …) |
| `trigger` | Push, pull-request, and tag trigger builders |
| `container` | Execution and service container builders |
| `stepbundle` | Step bundle definition and call-site reference builders |
| `withgroup` | `with` group builder for container-scoped steps |
| `serialize` | YAML/JSON serialization and `ValidatedPrint` |
| `validate` | Structural validation without external calls |

## Typed step builders

The SDK ships typed builders for the most common steps. Each builder exposes step-specific input methods that are checked at compile time:

| Function | Step |
|---|---|
| `step.GitClone()` | `git-clone` |
| `step.ActivateSshKey()` | `activate-ssh-key` |
| `step.Script()` | `script` |
| `step.XcodeTest()` | `xcode-test` |
| `step.XcodeArchive()` | `xcode-archive` |
| `step.AndroidBuild()` | `android-build` |
| `step.AndroidUnitTest()` | `android-unit-test` |
| `step.DeployToBitriseIo()` | `deploy-to-bitrise-io` |
| `step.Fastlane()` | `fastlane` |
| `step.CachePull()` | `cache-pull` |
| `step.CachePush()` | `cache-push` |
| `step.Slack()` | `slack` |
| `step.CocoapodsInstall()` | `cocoapods-install` |
| `step.ExportXcarchive()` | `export-xcarchive` |
| `step.FirebaseAppDistribution()` | `firebase-app-distribution` |
| `step.FlutterBuild()` | `flutter-build` |
| `step.FlutterTest()` | `flutter-test` |
| `step.GradleRunner()` | `gradle-runner` |
| `step.Npm()` | `npm` |
| `step.SignApk()` | `sign-apk` |
| `step.Yarn()` | `yarn` |

For any step not listed above use `step.From(id, version)`.

## Adding builders for more steps

`cmd/stepgen` generates typed builders for any step in the [Bitrise steplib](https://github.com/bitrise-io/bitrise-steplib). To add a new step:

1. Add its ID to `stepgen.json`:
   ```json
   {
     "steps": ["your-new-step"]
   }
   ```

2. Regenerate:
   ```sh
   go generate ./step/
   ```

   Or run the generator directly for a specific step:
   ```sh
   go run ./cmd/stepgen your-new-step
   ```

The generator fetches the latest step version from the steplib, parses the step's input definitions, and writes a `step/gen_<step_id>.go` file with a fully typed builder.

When using the GitHub source (default), set `GITHUB_TOKEN` to avoid the 60 req/hour unauthenticated rate limit. The `--jobs` flag (default `10`) controls how many steps are fetched in parallel:

```sh
GITHUB_TOKEN=ghp_... go run ./cmd/stepgen activate-ssh-key git-clone script
```

For a full regeneration of all builders without a local clone, increase concurrency up to the rate limit:

```sh
GITHUB_TOKEN=ghp_... go run ./cmd/stepgen --config=stepgen.json --output=step --jobs=20
```

#### Version fingerprinting

By default the generator reads the `// Step: <id> (<version>)` header from each existing generated file. If the latest version in the steplib matches, the step is skipped with no further fetches — saving two network round-trips per unchanged step. Use `--force` to bypass this and regenerate everything:

```sh
go run ./cmd/stepgen --config=stepgen.json --output=step --force
```

## Updating all step builders at once

To refresh every builder to the latest steplib version in one command, do a sparse clone of the steplib (no full history, only the `steps/` tree) and run the generator against it:

```sh
# Clone the steplib — only downloads step metadata, not full git history
git clone --depth 1 --filter=blob:none --sparse \
  https://github.com/bitrise-io/bitrise-steplib.git /tmp/bitrise-steplib
cd /tmp/bitrise-steplib && git sparse-checkout set steps

# Regenerate all builders from the local clone
cd /path/to/bitrise-pipeline-sdk
go run ./cmd/stepgen --steplib-dir=/tmp/bitrise-steplib --config=stepgen.json --output=step
```

To also pick up brand-new steps that have been added to the steplib since the last run, update `stepgen.json` first:

```sh
# Rebuild the steps list from the cloned steplib, preserving any existing skip list
python3 -c "
import json, os
existing = json.load(open('stepgen.json')) if os.path.exists('stepgen.json') else {}
skip = set(existing.get('skip', []))
steps = sorted(s for s in os.listdir('/tmp/bitrise-steplib/steps') if s not in skip)
existing['steps'] = steps
print(json.dumps(existing, indent=2))
" > stepgen.json

# Then regenerate
go run ./cmd/stepgen --steplib-dir=/tmp/bitrise-steplib --config=stepgen.json --output=step
```

If you already have a clone you want to reuse, pull the latest changes first:

```sh
cd /tmp/bitrise-steplib && git pull
```

## Examples

- [examples/basic](examples/basic/main.go) — Go service with graph pipeline, service container, and triggers
- [examples/ios](examples/ios/main.go) — iOS project: Xcode test + App Store archive
- [examples/android](examples/android/main.go) — Android project: unit tests + AAB release build
- [examples/monorepo](examples/monorepo/main.go) — Monorepo: parallel per-service tests in a graph pipeline

## `bitrise-gen` CLI reference

```
Usage: bitrise-gen <command> [flags]

Commands:
  run       <script.go>   Run a pipeline script and print YAML to stdout
  validate  <script.go>   Validate a pipeline script; exit 1 if invalid
  scaffold  [--output]    Write a starter pipeline script
  help                    Show this message
```

### `run`

Compiles and runs `<script.go>` via `go run`. All output (YAML) is forwarded to stdout so you can pipe it directly to `bitrise run --config -`.

### `validate`

Runs the script, captures the YAML output, and runs the SDK's full validation suite (structural checks + `bitrise/v2` normalisation and validation). Warnings go to stderr; errors cause a non-zero exit.

### `scaffold`

Writes a ready-to-run pipeline script to `--output` (default `pipeline.go`). Fails if the file already exists.

## License

[MIT](LICENSE)
