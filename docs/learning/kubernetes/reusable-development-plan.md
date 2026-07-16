# Reusable Kubernetes Development Deployment

## Status

This document is a design and learning plan. It does not describe a production Kubernetes
deployment.

The target is a reusable local development workflow in `kind` that supports the existing Book
Social environment model:

- `dev`: SQLite;
- `stage`: PostgreSQL;
- `prod`: PostgreSQL in a local prod-like environment.

`prod` in this document means the existing `APP_ENV=prod` application profile running locally. It
does not represent a real production deployment or production data.

## Goals

- Reuse common application manifests across environments.
- Preserve the current `APP_ENV` and database selection model.
- Isolate dev, stage, and prod-like resources with namespaces.
- Support repeatable local deployment and debugging in `kind`.
- Make environment differences explicit and reviewable.
- Keep secrets out of Git.

## Non-Goals

- Operating a public or production cluster.
- High availability PostgreSQL.
- Multiple Book Social application replicas while using SQLite.
- Cloud load balancers, public Ingress, autoscaling, or cluster administration.
- Replacing Docker Compose workflows immediately.

## Design Decisions

### Kustomize for Environment Composition

Use Kustomize rather than copied manifest directories or Helm for the first reusable deployment
iteration. It is available through `kubectl`, preserves ordinary Kubernetes YAML, and supports a
shared base with environment overlays.

Use these commands to render, inspect, and apply an environment:

```bash
kubectl kustomize k8s/overlays/dev
kubectl diff -k k8s/overlays/dev
kubectl apply -k k8s/overlays/dev
```

Helm is a separate learning module after the Kustomize structure is understood.

### Namespaces

Use one namespace per environment:

```text
book-social-dev
book-social-stage
book-social-prod
```

Namespaces allow the same resource names, such as `book-social-api` and `postgres`, to exist in
each environment without collisions.

### Environment Matrix

| Environment | Namespace           | App environment | Database                    | App replicas |
|-------------|---------------------|-----------------|-----------------------------|--------------|
| dev         | `book-social-dev`   | `APP_ENV=dev`   | SQLite with a PVC           | 1            |
| stage       | `book-social-stage` | `APP_ENV=stage` | Local PostgreSQL with a PVC | 1 initially  |
| prod-like   | `book-social-prod`  | `APP_ENV=prod`  | Local PostgreSQL with a PVC | 1 initially  |

The PostgreSQL instances are local development resources. They are isolated by namespace and use
different database names, credentials, and persistent volumes.

### Database Workloads

Keep the current dev SQLite workflow as a single-replica application with its own PVC.

For stage and prod-like environments, use a single-replica PostgreSQL StatefulSet, a ClusterIP
Service named `postgres`, and a PVC for PostgreSQL data. The app DSN can then use the in-cluster
service name, for example:

```text
postgres://book_social:<password>@postgres:5432/book_social_stage?sslmode=disable
```

The password is supplied by a Secret, not committed in the DSN ConfigMap.

### Health Model

The reusable deployment must replace the shared `/healthz` endpoint with:

- `GET /livez` for the process health check;
- `GET /readyz` for application readiness and required database availability.

The liveness probe must use `/livez`. The readiness probe must use `/readyz`.

### Image Strategy

Do not rely on mutable tags such as `book-social:dev` for reusable environment rollouts. Use a
unique image tag, preferably based on a Git commit or build version, and set it in the relevant
overlay.

## Target Manifest Layout

```text
k8s/
  base/
    app/
      deployment.yaml
      service.yaml
      kustomization.yaml
  components/
    postgres/
      service.yaml
      statefulset.yaml
      kustomization.yaml
  overlays/
    dev/
      namespace.yaml
      configmap.yaml
      pvc.yaml
      kustomization.yaml
    stage/
      namespace.yaml
      configmap.yaml
      postgres-secret.example.yaml
      postgres-pvc.yaml
      kustomization.yaml
    prod/
      namespace.yaml
      configmap.yaml
      postgres-secret.example.yaml
      postgres-pvc.yaml
      kustomization.yaml
  kind/
    kind-cluster.yaml
```

The exact directory names may be adjusted during implementation, but the separation of base,
database component, and environment overlays should remain.

## Implementation Stages

### Stage 1: Establish the Kustomize Base

- Move common app Deployment and Service configuration into `k8s/base/app/`.
- Add a base `kustomization.yaml`.
- Keep image, health probes, labels, and container port in the base.
- Render the base with `kubectl kustomize` before creating overlays.

Definition of done: the rendered base contains valid common application resources and no
environment-specific database credentials.

### Stage 2: Create the Dev Overlay

- Add the `book-social-dev` namespace.
- Set `APP_ENV=dev` and the SQLite DSN.
- Add the existing SQLite PVC and mount `/app/data`.
- Keep `replicas: 1`.
- Apply with `kubectl apply -k k8s/overlays/dev`.

Definition of done: dev behavior matches the current local kind workflow.

### Stage 3: Create the Stage PostgreSQL Overlay

- Add the `book-social-stage` namespace.
- Add a PostgreSQL StatefulSet, Service, and PVC.
- Add a ConfigMap with `APP_ENV=stage` and a DSN that uses the `postgres` Service.
- Create the PostgreSQL credential Secret locally without committing its real value.
- Add database initialization or migration as a separate Job.

Definition of done: the stage app reaches `Ready`, connects to local PostgreSQL, and serves the
catalog using the PostgreSQL repository.

### Stage 4: Create the Prod-Like PostgreSQL Overlay

- Add the `book-social-prod` namespace.
- Create an isolated PostgreSQL instance and PVC.
- Set `APP_ENV=prod` and a separate database name and Secret.
- Use a distinct immutable application image tag for rollout practice.

Definition of done: prod-like can run alongside dev and stage without sharing namespaces, data,
or credentials.

### Stage 5: Improve Application Health and Lifecycle

- Add `/livez` and `/readyz` to the Go application.
- Make `/readyz` check the selected database connection.
- Add a startup probe if initialization or migrations require it.
- Add graceful shutdown handling for Kubernetes termination.
- Add modest CPU and memory requests and limits.

Definition of done: a temporary database failure makes the app not ready without causing an
unnecessary liveness restart.

### Stage 6: Deployment Lifecycle and Migration Jobs

- Use unique image tags for each rollout.
- Verify `kubectl rollout status`, `history`, and `undo`.
- Run schema changes through a Kubernetes Job rather than app startup behavior.
- Decide explicitly whether seed data belongs in each local environment.

Definition of done: a failed image or migration can be diagnosed and recovered without manually
editing live cluster resources.

## Secret Handling

Do not commit real PostgreSQL passwords, tokens, or kubeconfig files.

Keep an example Secret manifest that documents required keys but contains placeholders. Create the
actual local Secret from a local environment file or an explicit `kubectl create secret` command.
Document the command, but do not place the resulting real Secret manifest under version control.

## GoLand-Assisted Workflow

Use GoLand as a visual development aid alongside the documented CLI commands.

1. Ensure the Kubernetes plugin is enabled and GoLand detects the `kind` kubeconfig context.
2. Add or select the `kind-book-social` context in the Services tool window.
3. Filter resources by `book-social-dev`, `book-social-stage`, or `book-social-prod`.
4. Use resource details, logs, events, and container shell access to investigate failures.
5. Use port forwarding from the IDE for temporary local inspection.
6. Configure `/bin/sh` as the container shell for Alpine images when needed.

The GoLand workflow must support, not replace, reproducible terminal commands in documentation and
CI. GoLand can display resource details, follow logs, open shells, and forward ports through its
Kubernetes integration. See
the [GoLand Services tool window documentation](https://www.jetbrains.com/help/go/services-tool-window.html?keymap=windows).

## Verification Matrix

For each overlay, verify:

```bash
kubectl kustomize k8s/overlays/<environment>
kubectl diff -k k8s/overlays/<environment>
kubectl apply -k k8s/overlays/<environment>
kubectl get pods -n book-social-<environment>
kubectl get pvc -n book-social-<environment>
kubectl rollout status deployment/book-social-api -n book-social-<environment>
```

Also verify that:

- the expected database is selected by `APP_ENV`;
- the application can reach the database through the in-cluster Service;
- readiness removes an unhealthy app pod from Service traffic;
- a rolling update can be rolled back;
- the three environments do not share data or credentials.

## Future Work

Potential follow-up topics, after the reusable local deployment is stable:

- Helm chart packaging based on the understood Kustomize resources;
- CI rendering and validation for each overlay;
- a local container registry or CI-built image workflow;
- GitOps tools such as Argo CD or Flux;
- cloud deployment design as a separate project decision.
