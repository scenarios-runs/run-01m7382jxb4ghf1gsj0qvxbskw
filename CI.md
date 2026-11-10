# Application CI

The workflow runs on pushes to every branch and on pull requests. New runs cancel older runs for the same ref. All jobs use GitHub-hosted Ubuntu 24.04 runners and read-only repository permissions. No custom secrets, registry login, image push, or deployment are needed.

| Job | Behavior | Typical duration |
| --- | --- | --- |
| lint | Set up Go 1.23, cache dependencies and build cache keyed by go.mod, download modules, run go vet ./... | 1–3 minutes |
| test | Independently set up Go and its cache, download modules, run go test ./... | 1–3 minutes |
| build | After lint and test pass, build and load the container, upload an image archive, scan, and smoke-test GET /healthz | 3–8 minutes |

Lint and test run in parallel. Total expected time is roughly 4–11 minutes; cold caches or vulnerability database downloads can increase it. Job timeouts are 10 minutes for lint/test and 20 minutes for build. No service containers are required: tests use httptest and in-memory storage.

The image is tagged ghcr.io/<owner>/<repository>:<commit SHA>, with the repository name lowercased for Docker compatibility. It stays local to the runner and is saved as the image-<commit SHA> artifact for seven days. Later jobs can download it with actions/download-artifact@v4 and run docker load -i image.tar. The archive is uploaded before scanning so it remains available for diagnosis even if scanning fails; its existence does not imply that checks passed.

Two Trivy passes implement the requested threshold: the first reports fixable HIGH and CRITICAL vulnerabilities without failing; the second fails on fixable CRITICAL findings. Both ignore unfixed vulnerabilities. Scanner/tool/database errors can also fail the job. Trivy runs from the versioned aquasec/trivy:0.58.1 container against the saved image archive; both passes share a runner-local database cache. Actions use major version tags. No separate scan job is needed because scan is part of build. There is no deploy job.

## Container

The Dockerfile installs dependencies from go.mod before copying source, compiles a static binary with Go 1.25.13, and copies only that binary into the required distroless Debian 12 runtime. It runs as UID/GID 65532, exposes port 8000, and uses an exec-form command. The runtime has no shell, so no Docker HEALTHCHECK is added; CI probes the existing /healthz endpoint from the runner. PORT is the only application environment variable; it defaults to 8000. Notes are stored in memory and disappear when the process restarts.

## Move into your repository

Copy app/ to the repository root, copy .github/workflows/ci.yml into the root .github/workflows/ directory, and copy CI.md alongside it. Do not copy .steps/. Commit and push to GitHub. If application files already live at the root, adjust workflow working-directory, cache-dependency-path, and Docker build context together. Enable GitHub Actions and allow the referenced actions under repository/organization policy. A fresh public repository otherwise needs no additional configuration.

In branch protection or a ruleset, select lint, test, and build as required checks after the first run. The build check includes the scan. Registry settings need no change for this pipeline because nothing is pushed. If publishing is later enabled, retain GHCR and GITHUB_TOKEN, grant packages: write only to the job that pushes, and restrict publishing to trusted push events. If adding a real deployment target later, configure its environment and credentials as repository/environment secrets and gate deployment on successful checks. No deployment credentials or other custom secrets are referenced by this workflow.

## Changes

- Added app/Dockerfile with cached dependency installation and a minimal non-root runtime; updated its builder to Go 1.25.13 to fix the embedded standard-library vulnerability.
- Added app/.dockerignore to exclude VCS files, local environments, dependencies, build output, caches, .env files, and private-key files from the build context.
- Adapted the supplied ci.none.yml template into .github/workflows/ci.yml, retaining lint, test, and build; removed registry authentication/pushing and omitted deploy.
- Updated this document and .steps/analyze.json assumptions.
- Application source, manifests, and tests are unchanged. The existing health endpoint required no changes.

## Pipeline repair

The published build failed during job setup because the pinned Trivy action referenced aquasecurity/setup-trivy@v0.2.2, which GitHub could not resolve. Replaced both action invocations with direct runs of the versioned Trivy container against the existing image archive. Reporting HIGH/CRITICAL findings, failing on fixable CRITICAL findings, and ignoring unfixed findings remain unchanged. Application files are unchanged.

The subsequent published build reached Trivy and failed on fixable CRITICAL CVE-2025-68121 (crypto/tls certificate validation) in the binary’s embedded Go v1.23.12 standard library. The log lists fixed versions 1.24.13 and 1.25.7. Updated only the Docker builder image to golang:1.25.13-bookworm, a newer patch also listed in the report. The static binary must be rebuilt to replace its vulnerable standard library; changing the distroless runtime alone would not fix it. Lint/test retain Go 1.23 to check the module’s declared minimum. Scan thresholds, application source, tests, module manifest, and runtime remain unchanged.

## Validation

Inspected source, manifest, tests, analysis, and workflow template offline. No packages were installed and no application, Docker build, or CI jobs were run locally. The supplied subsequent GitHub Actions log confirms lint/test passed and identifies the critical finding above; the repaired image and scan must be verified by the next GitHub Actions run. Action availability and remote image contents could not be checked without network access.

## Assumptions

- Use Go 1.23 from go.mod; no patch version is specified.
- Run commands from the application root; start runs the binary produced by the build command.
- Keep the app/ directory in the repository root and place .github/workflows/ci.yml at the repository root.
- The user request forbids pushing images; use a GHCR-formatted local tag and an artifact only, with no registry login or deployment.
- Report fixable HIGH and CRITICAL findings; fail only on fixable CRITICAL findings using a separate scan.
- Use app/go.mod as the Go cache dependency key because this standard-library-only application has no go.sum.
- GitHub-hosted runners have network access for actions, base images, Go toolchains, and Trivy vulnerability databases; local generation is offline.
- The required distroless Debian 12 nonroot tag is the runtime platform pin; this image does not offer a numeric minor-version tag or a shell healthcheck.
- Go 1.23 remains the module minimum and lint/test toolchain; use Go 1.25.13 for the container build to remediate CVE-2025-68121. The fixed-version evidence comes from the supplied Trivy log; image availability and rebuilt scan results require GitHub Actions verification.
- CI duration estimates are approximate and depend on caches and runner availability.

- For the repair, use Trivy 0.58.1 as an explicit scanner image version; remote image availability cannot be verified offline.
