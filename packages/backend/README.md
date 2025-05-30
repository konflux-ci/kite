# Issues Backend

## Development setup

### 1. Clone the Repository

```bash
git clone https://github.com/konflux-ci/kite.git
cd kite
```

### 2. Start Minikube

```bash
# Start a minikube cluster
minikube start

# Verify it's running
minikube status
```

### 3. Generate kube-config.yaml

For an easy setup process, Minikube is recommended as a local Kubernetes cluster.

Once it's installed, ensure it's set as the current context:
```bash
kubectl config current-context
minikube
```

If another value is returned, set minikube as the current context:
```bash
kubectl config set-context minikube
Context "minikube" modified.
```
Next, run this script to generate the `kube-config.yaml` file for the backend service:
```
chmod +x scripts/dev/generate-kubeconfig.sh
./scripts/dev/generate-kubeconfig.sh
```

This is used by the service to talk to the cluster, allowing it to perform actions like limiting issues by namespaces.

### 4. Start the Development Environment with <Docker|Podman> Compose

```bash
# Build and start the services
<docker|podman> compose -f compose.yaml up -d --build

# Check if services are running
<docker|podman> compose ps

# Stop services when needed
<docker|podman> compose -f compose.yaml down -v
```

[Air](https://github.com/air-verse/air) is used for hot reloading on your changes.

### 5. Access the Application

- API: http://localhost:8080/health


## Migrations

Migration operations can be done using the `Makefile`.

Make sure you export the ENV variables in `.env.development`:
```bash
export $(grep -v '^#' .env.development | xargs)
```
