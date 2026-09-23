# Deployment Guide

This document outlines how to deploy Whale Tracker to Kubernetes using Kustomize, either manually for local/initial setup or automatically via GitHub Actions CI/CD.

---

## Prerequisites

- Kubernetes cluster (v1.26+)
- `kubectl` configured with cluster access
- `cert-manager` installed in the cluster (for Let's Encrypt TLS)

---

## 1. Local / Manual Deployment

Use this section to deploy from your machine, set up a new environment, or troubleshoot cluster state.

### Step 1: Stage Runtime Context Files

Copy the example file to create your .env</p>
And fill in your specific values
```bash
cp backend/.env.prod.example backend/.env
```

> [!NOTE]
> Ensure `backend/k8s/env-context/.env` contains your domain host setting (e.g., `HOSTNAME=kernel-panic.mooo.com`) along with database and auth secrets.

```bash
mkdir -p backend/k8s/env-context

# Copy environment configuration
cp backend/.env backend/k8s/env-context/.env

# Copy Keycloak realm import and theme build
cp backend/deploy/keycloak/realm-export.json backend/k8s/env-context/
cp backend/deploy/keycloak/fintech-theme-build/fintech-theme.jar backend/k8s/env-context/
```

Step 2: Create Namespace & Secrets

```bash
# Create namespace
kubectl create namespace whale-tracker-prod --dry-run=client -o yaml | kubectl apply -f -

# Generate runtime application secrets from .env
kubectl create secret generic app-secrets \
  --from-env-file=backend/k8s/env-context/.env \
  -n whale-tracker-prod \
  --dry-run=client -o yaml | kubectl apply -f -
```

Step 3: Deploy Full Stack via Kustomize

Apply all persistent volumes, databases, services, generated ConfigMaps, workloads, and Ingress rules with a single command:
```bash
kubectl apply -k backend/k8s/ -n whale-tracker-prod
```

Step 4: Verify Deployment
```bash
# Monitor pod rollout status
kubectl get pods -n whale-tracker-prod -w

# Verify Ingress routing and Let's Encrypt TLS certificate status
kubectl get ingress,certificate -n whale-tracker-prod
```

### Accessing Grafana

Since Grafana is kept internal to the cluster namespace, use `port-forward` to access the dashboard locally:

```bash
# http://localhost:3000
kubectl port-forward svc/grafana 3000:3000 -n whale-tracker-prod
```

## Automated CI/CD Deployment (GitHub Actions)

Once your environment is set up, GitHub Actions automates manifest validation and deployment on every commit pushed to the main branch.
Repository Setup

Add the following secret in GitHub Settings -> Secrets and variables -> Actions:
```bash
KUBECONFIG — The base64 or raw kubeconfig content with write access to your Kubernetes cluster.
```

Pipeline Workflow

On every push to main, the workflow (.github/workflows/deploy.yml):

    Checks out the repository and configures kubectl.

    Assembles backend/k8s/env-context/.env and stage assets.

    Executes kubectl apply -k backend/k8s/ -n whale-tracker-prod to perform zero-downtime updates.

Troubleshooting
TLS Certificate Issues (cert-manager)

If the SSL certificate remains in a pending state (READY: False):

```bash
# Inspect ACME challenge and order status for specific failures
kubectl describe challenge -n whale-tracker-prod
kubectl describe order -n whale-tracker-prod

# If necessary, force cert-manager to re-issue the certificate
kubectl delete certificate kernel-panic-tls -n whale-tracker-prod
kubectl delete secret kernel-panic-tls -n whale-tracker-prod
kubectl apply -k backend/k8s/ -n whale-tracker-prod
```