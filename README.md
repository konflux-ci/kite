# Konflux Issue Tracking Engine (Kite) :kite:

![Go CI Checks](https://github.com/konflux-ci/kite/actions/workflows/go-ci-checks.yaml/badge.svg)

:construction: **Proof of Concept** - APIs and core functionality may change. See [Status](#status) below.

---

## What is Kite?

Kite is a set of components that **detect, create and track issues** that can disrupt applications in Konflux.

Typical issues include:

- Tekton `PipelineRun` failures
- Release pipeline errors
- MintMaker issues
- Cluster-wide problems affecting builds/releases

---

### Why Kite?
- **Single source of truth** for diagnosing cluster incidents
- **Extensible** controllers to watch your own resources
- **Automation-first** via webhooks and API
- **CLI-first** tooling for developers, SREs and power users

---

## Features
- **Issue Tracking**: Centralized, extensible store for build/test failures, release problems, and more.
- **CLI Integration**: Use a standalone CLI tool to query issues in the store.
- **Namespace Scoping**: Issues are namespaces for isolation and security. (**WIP**)
- **Automation**: Webhooks for automatic issue creation/resolution.
- **REST API**: Integrate with external tools (such as the Issues Dashboard **WIP**).
- **Extensible Operator**: Add custom controllers to watch cluster resources and open/resolve issues in Kite.

All these components work together to create and track issues that may disrupt your ability to build and deploy applications in Konflux.

---

## Repository structure

- **Backend**: [`packages/backend`](./packages/backend/): Go server (Gin) + PostgreSQL database.

- **CLI**: [`packages/cli`](./packages/cli/): Go CLI tool (can also run as a `kubectl` plugin.)

- **Operator**: [`packages/operator`](./packages/operator/): Kubernetes operator for resource watchers.

**Included controller(s)**:
- [PipelineRun Controller](./packages/operator/internal/controller/pipelinerun_controller.go): Tracks Tekton `PipelineRun` successes/failures (reference controller implementation)

---

## Getting started
### Integrate with Kite (recommended path)

1. Build a [custom controller](./packages/operator/docs/ControllerDevelopmentGuide.md) that watches your resources and posts state changes
2. Implement a custom [webhook endpoint](./packages/backend/docs/Webhooks.md) with payloads tailored to your events.
  *You can also use the standard API, but webhooks usually make issue creation/resolution simpler.*

### Alternative approach
See the [External Service Integration](./packages/backend/docs/ExternalServiceIntegration.md) docs.

---

## Docs
- [API](./packages/backend/docs/API.md)
- [Webhooks](./packages/backend/docs/Webhooks.md)
- [Controller Development Guide](./packages/operator/docs/ControllerDevelopmentGuide.md)
- [External Service Integration](./packages/backend/docs/ExternalServiceIntegration.md)

---

## Prerequisites

To work with this project, ensure you have the following installed:

- [Docker](https://docs.docker.com/get-docker/) or [Podman](https://podman.io/docs/installation)
- [Go](https://golang.org/doc/install) v1.23 or later
- [Make](https://www.gnu.org/software/make/)
- [Minikube](https://minikube.sigs.k8s.io/docs/start/) – for local development
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) – for local development

---

## Status
This is still a proof-of-concept project.
See the Roadmap below for a high-level progress overview and plans.

### Roadmap
- [x] Backend API service is built and released through Konflux
- [x] Backend API service is running on a public staging cluster
- [x] Bridge Operator is built and released through Konflux
- [ ] **Current** - Bridge Operator is live on a public staging cluster
- [ ] Konflux teams onboard their controllers onto Bridge Operator, generating issues
- [ ] Controllers can be selected using a configuration, rather than loading all of them
