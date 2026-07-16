# Kubernetes Learning and Implementation Sequence

## Purpose

This sequence combines the learning modules in `learning-path.md` with the implementation stages
in `reusable-development-plan.md`.

The modules describe skills. The stages describe reusable local deployment work. They should not
be completed in strict numeric order. The sequence below follows technical dependencies and keeps
each task small enough to learn from and review.

## Session Pattern

Use this pattern for each step:

```text
30-45 minutes: study the concept
60-90 minutes: implement one focused change
15 minutes: verify with CLI commands
15 minutes: create and diagnose a controlled failure or rollback
5-10 minutes: record the result and remaining questions
```

Use `kubectl` commands as the primary source of truth. Use GoLand to inspect the same resources,
logs, and namespaces from a second interface.

## Recommended Order

| Order | Learn                                            | Implement                                                                                              | GoLand practice                                                                                            |
|-------|--------------------------------------------------|--------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------|
| 0     | Close the current local baseline                 | Commit the current local `kind` workflow and recreate it from scratch using `k8s/k8s.md`               | Add the `kind-book-social` context and locate the Deployment, Service, PVC, and pod logs                   |
| 1     | Module 1: operate and debug                      | Create controlled failures: wrong image, broken configuration, and PVC problem                         | Compare `kubectl describe`, events, and logs with the Services view                                        |
| 2     | Module 2: configuration, namespaces, and Secrets | Learn labels and selectors; create a disposable namespace and a non-sensitive ConfigMap                | Switch namespace filters and inspect ConfigMap values and pod environment                                  |
| 3     | Module 5 and Stage 1                             | Convert the current working manifests into a Kustomize base                                            | Edit YAML in GoLand and render with `kubectl kustomize` in the Terminal                                    |
| 4     | Stage 2                                          | Build the `dev` overlay with `book-social-dev`, SQLite PVC, and one replica                            | Filter to `book-social-dev`; inspect PVC binding and health logs                                           |
| 5     | Module 3: deployment lifecycle                   | Use explicit image tags in dev; perform rollout, history inspection, failed tag, and rollback          | Watch old and new pods during rollout and compare their logs                                               |
| 6     | Stage 3                                          | Build `stage`: namespace, PostgreSQL StatefulSet, Service, PVC, ConfigMap, and local Secret            | Inspect application and PostgreSQL pods separately; use logs and shell access to diagnose database startup |
| 7     | Module 4 and Stage 5                             | Add `/livez`, `/readyz`, database-aware readiness, resource requests and limits, and graceful shutdown | Forward a port, call both endpoints, and observe readiness when the database is unavailable                |
| 8     | Stage 4                                          | Build isolated local `prod-like`: namespace, PostgreSQL PVC, Secret, and image tag                     | Verify that stage and prod-like resources, logs, PVCs, and credentials are separate                        |
| 9     | Stage 6                                          | Add a migration Job and verify failure recovery, rollout, and rollback behavior                        | Inspect Job status and logs, then follow the application rollout in Services                               |
| 10    | Module 6: Helm basics                            | Create a minimal chart from the understood Kustomize resources                                         | Use GoLand for chart and YAML editing; run `helm lint` and `helm template` in the Terminal                 |

## Why This Order

- Module 1 comes first because every later Kubernetes task needs debugging ability.
- Module 5 comes before its numeric position because Kustomize is required for reusable
  `dev`, `stage`, and `prod` overlays.
- Stage 2 creates a safe dev environment before PostgreSQL adds database complexity.
- Module 3 provides rollout and rollback skills before stage and prod-like deployments.
- Stage 5 comes before Stage 4 deliberately. The prod-like overlay should inherit the improved
  `/livez` and `/readyz` health model instead of the temporary shared `/healthz` endpoint.
- Helm comes last because it should package Kubernetes resources that are already understood.

## GoLand Practice Rule

For every implementation step:

1. Apply and verify the change through documented CLI commands.
2. Open the same namespace in the GoLand Services tool window.
3. Inspect the resource that changed.
4. Use GoLand logs or shell access to confirm the CLI diagnosis.
5. Record which interface provided the clearest evidence.

For Alpine-based Book Social containers, configure `/bin/sh` as the GoLand container shell if the
default `/bin/bash` is unavailable.

GoLand can work with kubeconfig contexts, namespace filters, resource details, logs, container
shells, and port forwarding. See
the [GoLand Services tool window documentation](https://www.jetbrains.com/help/go/services-tool-window.html?keymap=windows).

## Immediate Next Step

Start with Order 0. Commit the current local `kind` baseline, then recreate it from scratch using
only the local kind guide before beginning new reusable deployment work.
