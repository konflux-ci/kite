# Konflux Issue Tracking Engine - Kite :kite:

![Go CI Checks](https://github.com/konflux-ci/kite/actions/workflows/go-ci-checks.yaml/badge.svg)

:construction: **Under Construction – Currently a Proof of Concept** :construction:

## About

Kite is a **centralized information store** for tracking issues that may disrupt your ability to build and deploy applications in Konflux.

These issues might include:

- PipelineRun failures
- Release errors
- MintMaker issues
- Cluster-wide problems

Kite **does not** actively monitor cluster resources.
It relies on external tools or workflows to report issues.

If a system can send requests to Kite, it can create and resolve issues.

This makes Kite flexible and easy to integrate into your existing workflows.

## Features

- **Issue Tracking**: Track build/test failures, release problems, and more in a centralized, extendable service.
- **CLI Integration**: Access and manage issues from your terminal or as a `kubectl` plugin.
- **Namespace Isolation**: Issues are scoped to Kubernetes namespaces for better security.
- **Automation-Friendly**: Supports webhooks for automatic issue creation and resolution.
- **API Access**: RESTful API for integration with external tools.

## Components

This monorepo is structured around two primary components:

- `packages/backend`: A Go-based `gin-gonic` server with a PostgreSQL database.
- `packages/cli`: A Go-based CLI tool that can run standalone or as a `kubectl` plugin.

## Integration

This service is flexible enough to accommodate various workflows and scenarios.

Please see the [API docs](./packages/backend/docs/API.md) for more information.

## Prerequisites

To work with this project, ensure you have the following installed:

- [Docker](https://docs.docker.com/get-docker/) or [Podman](https://podman.io/docs/installation)
- [Go](https://golang.org/doc/install) v1.23 or later
- [Make](https://www.gnu.org/software/make/)
- [Minikube](https://minikube.sigs.k8s.io/docs/start/) – for local development
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) – for local development

