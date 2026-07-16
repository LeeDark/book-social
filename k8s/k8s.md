# Book Social on kind

## Purpose and Scope

This folder contains a local Kubernetes learning workflow for Book Social using
[kind](https://kind.sigs.k8s.io/).

It is intended for local development and learning only. It is not production deployment
infrastructure and does not make production claims.

The current workflow runs the application with:

- `APP_ENV=dev`;
- a single replica;
- a local SQLite database mounted at `/app/data`;
- a local PersistentVolumeClaim;
- a NodePort exposed on `127.0.0.1:18080`.

The `k8s/` folder is currently ignored by Git. It remains local until Step 10 decides whether the
manifests should become tracked project infrastructure.

## Folder Structure

```text
k8s/
  app/
    configmap.yaml
    deployment.yaml
    pvc.yaml
    service.yaml
  kind/
    kind-cluster.yaml
  k8s.md
```

`k8s/app/` contains Kubernetes resources for Book Social. Apply this folder with `kubectl`.

`k8s/kind/kind-cluster.yaml` configures the kind cluster itself. Pass it to `kind create cluster`.
Do not pass it to `kubectl apply`.

## Prerequisites

Run these commands from the repository root outside the Codex sandbox:

```bash
docker version
kind version
kubectl version --client
```

Check the configured Kubernetes contexts:

```bash
kubectl config get-contexts
```

## Create the Cluster

List existing kind clusters:

```bash
kind get clusters
```

Create the local cluster with the NodePort mapping from `k8s/kind/kind-cluster.yaml`:

```bash
kind create cluster --name book-social --config k8s/kind/kind-cluster.yaml
kubectl config use-context kind-book-social
```

Verify that Kubernetes is reachable:

```bash
kubectl cluster-info
kubectl get nodes
```

The kind configuration maps the host port `127.0.0.1:18080` to port `30080` on the kind node.
The Book Social Service uses the same `nodePort: 30080` and forwards traffic to the application on
port `8080`.

Changes to `k8s/kind/kind-cluster.yaml` require recreating the kind cluster. Recreating the cluster
deletes the local PersistentVolumeClaim and its SQLite data.

## Build and Load the Image

Build the application image using the same tag as Docker Compose and the Deployment:

```bash
make docker/build
```

Load the local image into the kind cluster:

```bash
kind load docker-image book-social:dev --name book-social
```

The kind nodes have a separate image store from the host Docker daemon. Loading the image is
required before a pod using `imagePullPolicy: IfNotPresent` can start.

## Apply the Application Manifests

Apply only the Book Social resources:

```bash
kubectl apply -f k8s/app/
```

This creates or updates:

- `book-social-api-config` ConfigMap;
- `book-social-api-data` PersistentVolumeClaim;
- `book-social-api` Service;
- `book-social-api` Deployment.

## Verify the Deployment

Wait for the Deployment rollout:

```bash
kubectl rollout status deployment/book-social-api
```

Inspect the main resources:

```bash
kubectl get deployment book-social-api
kubectl get pods -l app=book-social-api
kubectl get service book-social-api
kubectl get pvc book-social-api-data
```

The expected state is:

- the Deployment has one available replica;
- the pod is `Running` and `Ready`;
- the PVC is `Bound`;
- the Service is `NodePort` with node port `30080`.

## Open the Application

The kind NodePort mapping exposes the application only on the local machine:

```text
http://127.0.0.1:18080/
http://127.0.0.1:18080/books
http://127.0.0.1:18080/healthz
```

Test the health endpoint and catalog from a terminal:

```bash
curl -i http://127.0.0.1:18080/healthz
curl -i http://127.0.0.1:18080/books
```

`GET /healthz` returns `200 OK`. The Deployment uses it for both readiness and liveness probes.

## Update Workflow

After changing Go code:

```bash
make docker/build
kind load docker-image book-social:dev --name book-social
kubectl rollout restart deployment/book-social-api
kubectl rollout status deployment/book-social-api
```

After changing an application manifest:

```bash
kubectl apply -f k8s/app/
kubectl rollout status deployment/book-social-api
```

After changing `k8s/app/configmap.yaml`, restart the Deployment. Environment variables loaded with
`envFrom` do not change inside existing containers:

```bash
kubectl apply -f k8s/app/configmap.yaml
kubectl rollout restart deployment/book-social-api
kubectl rollout status deployment/book-social-api
```

## Diagnostics

Check resource status and recent events:

```bash
kubectl get pods -l app=book-social-api -o wide
kubectl get pvc
kubectl get events --sort-by=.lastTimestamp
```

Describe a failing pod or claim:

```bash
kubectl describe pod <pod-name>
kubectl describe pvc book-social-api-data
```

Read application logs:

```bash
kubectl logs deployment/book-social-api --tail=100
```

Common failures:

- `ImagePullBackOff`: build `book-social:dev`, then run `kind load docker-image` for the current
  cluster.
- PVC remains `Pending`: inspect `kubectl get storageclass` and `kubectl describe pvc
  book-social-api-data`. The current PVC expects the local `standard` StorageClass.
- `FailedMount`: inspect the pod and PVC descriptions for volume errors.
- `CrashLoopBackOff`: inspect the pod description and application logs.

## Cleanup

Delete only the application resources:

```bash
kubectl delete -f k8s/app/
```

Delete the entire local cluster when it is no longer needed:

```bash
kind delete cluster --name book-social
```

Deleting the cluster removes the local SQLite storage.

## Known Limitations

- This workflow supports local kind learning only.
- SQLite uses one replica. It is not a shared multi-replica database.
- The local PVC is not durable after the kind cluster is deleted.
- The NodePort mapping is local-only and binds to `127.0.0.1`.
- `/healthz` checks that the HTTP application responds. It does not run a separate database query.
- Docker Compose stage and prod PostgreSQL workflows are separate local environments; this kind
  workflow uses SQLite dev only.

## Next Stage: Reusable Development Deployment

Before this becomes a reusable development deployment, the health model must be expanded:

- add `GET /livez` for process liveness;
- add `GET /readyz` for readiness and required dependency checks;
- use `/livez` for the liveness probe and `/readyz` for the readiness probe;
- decide on a shared database strategy before running more than one replica.
