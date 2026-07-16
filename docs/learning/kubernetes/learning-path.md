# Kubernetes Learning Path

## Purpose

This path develops practical Kubernetes knowledge for a Go backend developer working on Book
Social. The goal is to understand how a containerized application is configured, deployed,
observed, and debugged in Kubernetes.

This is not a plan for operating a production cluster. It deliberately focuses on application
workloads and the developer responsibilities around them.

## Starting Point

The local `kind` workflow is the completed baseline. It currently includes:

- a local `kind` cluster;
- a Book Social Deployment and NodePort Service;
- ConfigMap-based configuration;
- a SQLite PersistentVolumeClaim with one application replica;
- `/healthz` readiness and liveness probes;
- local image build and `kind load docker-image` workflow;
- basic inspection with `kubectl get`, `describe`, `logs`, and events.

See [the local kind guide](../k8s/k8s.md) for the current commands and manifest layout.

## Learning Principles

- Learn each Kubernetes concept through Book Social before adding another abstraction.
- Use `kubectl` commands as the primary source of truth.
- Make a controlled failure, inspect it, then fix it.
- Keep local `kind` practice distinct from production infrastructure claims.
- Record a concise outcome and verification command for each completed module.

## Module 1: Operate and Debug a Workload

### Goals

- Read Deployment, Pod, Service, and PVC status.
- Use logs, events, `describe`, and `exec` to identify a failed startup.
- Distinguish common failure states.

### Practice

- Load an image with the wrong tag and diagnose `ImagePullBackOff`.
- Use an unavailable StorageClass or invalid PVC reference and diagnose `Pending` or
  `FailedMount`.
- Change an application setting to cause a failed startup and inspect `CrashLoopBackOff`.
- Restore the working configuration after each exercise.

### Definition of Done

Given a non-ready pod, identify whether the cause is image availability, configuration, storage,
or application startup, using Kubernetes evidence rather than guesswork.

## Module 2: Configuration, Namespaces, and Secrets

### Goals

- Understand labels, selectors, namespaces, ConfigMaps, and Secrets.
- Keep non-sensitive configuration separate from sensitive values.
- Understand in-cluster Service DNS.

### Practice

- Create `book-social-dev`, `book-social-stage`, and `book-social-prod` namespaces.
- Inject `APP_ENV` and `APP_DB_DSN` from a ConfigMap.
- Create a local PostgreSQL credential Secret without committing its real value to Git.
- Confirm configuration in a running pod with `kubectl exec` and `printenv`.
- Resolve a PostgreSQL Service from an application pod in the same namespace.

### Definition of Done

Explain why ConfigMaps and Secrets are different, why a Secret is not automatically secure merely
because it is base64 encoded, and how a pod receives its configuration.

## Module 3: Deployment Lifecycle

### Goals

- Understand the relationship between Deployment, ReplicaSet, and Pod.
- Observe rolling updates and rollback.
- Use explicit image tags rather than relying on a mutable local tag.

### Practice

- Deploy a new image tag and observe the rollout.
- Inspect rollout history.
- Deliberately deploy a non-working image tag, then roll back.
- Compare readiness behavior during a rollout with and without a healthy endpoint.

### Definition of Done

Use `kubectl rollout status`, `history`, and `undo` confidently, and explain how readiness affects
the availability of a new pod during a rolling update.

## Module 4: Application Lifecycle and Resources

### Goals

- Separate process liveness from dependency readiness.
- Understand startup, liveness, and readiness probes.
- Learn the basic purpose of CPU and memory requests and limits.
- Understand pod termination from the application's perspective.

### Practice

- Replace the current shared `/healthz` check with `/livez` and `/readyz`.
- Make `/readyz` check the required database dependency without making `/livez` depend on it.
- Add modest CPU and memory requests and limits.
- Add graceful shutdown handling in the Go application and observe a pod termination.

### Definition of Done

Explain why a failed readiness check removes a pod from Service traffic, while a failed liveness
check may restart the container.

## Module 5: Reusable Manifest Structure

### Goals

- Eliminate copied manifests across environments.
- Render and inspect the exact configuration before applying it.
- Keep shared application configuration separate from environment-specific differences.

### Practice

- Convert the current local manifests into a Kustomize `base`.
- Create `dev`, `stage`, and `prod` overlays.
- Run `kubectl kustomize` and `kubectl diff -k` before applying an overlay.

### Definition of Done

Explain which resources belong in the base and which values belong in an overlay. Apply an
environment with `kubectl apply -k`.

## Module 6: Helm Basics

Helm follows manual YAML and Kustomize work. It is a packaging and templating layer, not a
replacement for understanding Kubernetes resources.

### Goals

- Understand chart structure, values files, templates, releases, upgrades, and rollback.
- Render a chart locally before installing it.

### Practice

- Create a minimal Book Social chart from the understood Kubernetes resources.
- Create separate values files for local dev, stage, and prod-like environments.
- Run `helm lint` and `helm template` before a local install.

### Definition of Done

Read the rendered YAML and explain every generated Deployment, Service, ConfigMap, and Secret
reference. Do not use Helm merely to hide unexplained templates.

## GoLand Workflow

GoLand is a useful inspection and debugging interface, but it does not replace the CLI workflow.

- Add the `kind` kubeconfig context in the Services tool window.
- Filter resources by namespace.
- Use resource details as a visual equivalent of `kubectl describe`.
- Follow pod logs and open a container console when debugging.
- Use IDE port forwarding for temporary inspection when needed.
- Configure `/bin/sh` as the container shell for Alpine-based Book Social images if the default
  `/bin/bash` is unavailable.

The Kubernetes integration requires the Kubernetes plugin and a detected kubeconfig. GoLand can
inspect resources, follow logs, open a shell, and forward ports. See the [GoLand Services tool
window documentation](https://www.jetbrains.com/help/go/services-tool-window.html?keymap=windows).

## Deferred Topics

The following topics are intentionally outside this learning path until there is a concrete
project need:

- production cluster administration;
- cloud provider infrastructure;
- Terraform or OpenTofu;
- complex Ingress or Gateway API configuration;
- autoscaling, node pools, and multi-cluster networking;
- complex Helm charts and GitOps controllers.

## Completion Outcome

After completing this path, the developer should be able to deploy and debug a containerized Go
application in a local Kubernetes cluster, explain its configuration and health model, and safely
work with a basic multi-environment manifest structure.
