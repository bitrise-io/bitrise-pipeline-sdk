// Code generation entry point for typed step builders.
//
// Run from the repository root:
//
//	go generate ./step/
//
// # Source modes
//
// By default stepgen fetches step metadata from the GitHub API.  Set
// GITHUB_TOKEN to avoid the unauthenticated rate limit (60 req/hour):
//
//	GITHUB_TOKEN=<token> go generate ./step/
//
// For offline / CI use, point stepgen at a local steplib clone via the
// STEPLIB_DIR environment variable — no network calls, no rate limits:
//
//	# one-time clone (sparse, < 50 MB)
//	git clone --depth 1 --filter=blob:none --sparse \
//	  https://github.com/bitrise-io/bitrise-steplib.git /tmp/steplib
//	cd /tmp/steplib && git sparse-checkout set steps
//
//	# generate with the local clone
//	STEPLIB_DIR=/tmp/steplib go generate ./step/
//
// You can also pass the flag explicitly (useful for one-off runs):
//
//	go run ./cmd/stepgen --steplib-dir=/tmp/steplib --config=stepgen.json --output=step

//go:generate go run ../cmd/stepgen --config=../stepgen.json --output=.
//go:generate go run ../cmd/readmegen --step-dir=. --readme=../README.md

package step
