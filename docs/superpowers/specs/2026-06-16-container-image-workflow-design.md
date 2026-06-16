# Container Image Build & Push Workflow

**Date:** 2026-06-16  
**Status:** Approved

## Goal

Add a GitHub Actions workflow that cross-compiles all five CLI binaries into a single multi-arch Docker image and pushes it to DockerHub on every push to `main`. The image is designed for local development on Apple Silicon and eventual k3s cluster deployment, where different binaries will run as distinct workloads (some continuously, some on schedule).

## New Files

- `Dockerfile` — repo root
- `.github/workflows/docker.yml` — separate from `go.yml`; different concerns

## Dockerfile

Alpine base for debuggability (exec into running containers in k3s). Binaries are statically linked (`CGO_ENABLED=0`), so no libc dependency. Docker buildx injects `TARGETARCH` (`amd64` or `arm64`) during a multi-arch build; the Dockerfile uses it to select the correct pre-compiled binary.

```dockerfile
FROM alpine:latest
ARG TARGETARCH
COPY bin/stockbatch-linux-${TARGETARCH}    /usr/local/bin/stockbatch
COPY bin/filterjson-linux-${TARGETARCH}    /usr/local/bin/filterjson
COPY bin/stockclient-linux-${TARGETARCH}   /usr/local/bin/stockclient
COPY bin/currentreturn-linux-${TARGETARCH} /usr/local/bin/currentreturn
COPY bin/targetreturn-linux-${TARGETARCH}  /usr/local/bin/targetreturn
```

No `ENTRYPOINT` — the binary name is specified as the container command, enabling runtime override without `--entrypoint`.

## Workflow: `.github/workflows/docker.yml`

**Trigger:** `push` to `main` branch only. Not on PRs (avoids pushing untested images) and not on tags (semver releases are a future concern).

**Platforms:** `linux/amd64`, `linux/arm64`. Covers k3s nodes and Apple Silicon local development (Docker Desktop on Mac runs Linux arm64 containers natively).

**Build strategy:** Pre-compile Go binaries on the runner using native cross-compilation (`GOOS`/`GOARCH`), then use `docker/build-push-action` to assemble and push the multi-arch image. No QEMU involved — Go handles cross-compilation natively, which is significantly faster than emulation.

**Job steps (single job):**

1. `actions/checkout@v6`
2. `actions/setup-go@v6` with `go-version-file: go.mod`
3. Cross-compile all five binaries for `linux/amd64` and `linux/arm64` into `bin/`, named `<binary>-linux-<arch>`
4. `docker/metadata-action@v5` — generates tags: `latest` and short SHA (e.g., `sliverman69/portfoliotools:a3f9b12`)
5. `docker/setup-buildx-action@v3`
6. `docker/login-action@v3` — authenticates to DockerHub using `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` secrets
7. `docker/build-push-action@v6` — builds multi-arch manifest and pushes both platform images atomically

## Image Tags

| Tag | Example | Purpose |
|-----|---------|---------|
| `latest` | `sliverman69/portfoliotools:latest` | Always points to most recent main build |
| short SHA | `sliverman69/portfoliotools:a3f9b12` | Pinnable, traceable; use in k3s manifests |

Semver tags (e.g., `v1.0.0`) are out of scope for now; add by extending the `docker/metadata-action` tags block when ready.

## Secrets Required

| Secret | Value |
|--------|-------|
| `DOCKERHUB_USERNAME` | `sliverman69` |
| `DOCKERHUB_TOKEN` | DockerHub access token (already configured) |

`DOCKERHUB_USERNAME` still needs to be added as a repository secret (the token is set but the username is a separate secret).

## Runtime Usage

Credentials (`~/.stockclientconfig.json`) are never baked into the image. Mount at runtime:

```sh
docker run --rm \
  -v ~/.stockclientconfig.json:/root/.stockclientconfig.json \
  sliverman69/portfoliotools stockclient -ticker AAPL
```

For k3s: mount the config as a Secret and project it into the pod at `/root/.stockclientconfig.json`. Each workload (CronJob or Deployment) specifies its binary as the container `command`.

## Relationship to Existing `go.yml`

The existing `go.yml` build job cross-compiles for `linux/amd64`, `darwin/amd64`, and `windows/amd64` and uploads binaries as GitHub Actions artifacts. That job serves direct binary downloads. The new `docker.yml` is independent — it recompiles specifically for Linux targets with `CGO_ENABLED=0` for container use. The two workflows share no steps and can evolve independently.
