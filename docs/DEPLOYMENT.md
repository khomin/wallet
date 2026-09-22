# Deployment Guide

This document outlines the steps required to deploy Whale Tracker to Kubernetes and maintain environment secrets.

## Prerequisites

- Kubernetes cluster (v1.26+)
- `kubectl` configured with cluster access
- `cert-manager` installed on the cluster
- GitHub Actions environment secrets setup (`KUBECONFIG`)

## Environment & Secrets Setup

1. Copy the production environment template:
   ```bash
   cp backend/.env.prod.example backend/.env
   ```

2. Create the app secret in the production namespace:
   ```bash
   kubectl create namespace whale-tracker-prod --dry-run=client -o yaml | kubectl apply -f -

   kubectl create secret generic app-secrets \
     --from-env-file=backend/.env \
     -n whale-tracker-prod \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

3. Create Keycloak ConfigMaps:
   ```bash
   kubectl create configmap keycloak-realm-import \
     --from-file=realm-export.json=./backend/deploy/keycloak/realm-export.json \
     -n whale-tracker-prod \
     --dry-run=client -o yaml | kubectl apply -f -

   kubectl create configmap keycloak-theme-jar \
     --from-file=./backend/deploy/keycloak/fintech-theme-build/fintech-theme.jar \
     -n whale-tracker-prod \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

## Kubernetes Deployment

Apply persistent volumes, infrastructure components, and workloads:

```bash
kubectl apply -f backend/k8s/postgres-pvc.yaml -n whale-tracker-prod
kubectl apply -f backend/k8s/databases.yaml -n whale-tracker-prod
kubectl apply -k ./backend/k8s -n whale-tracker-prod
```

## GitHub Actions Automated Deployment

CI/CD is configured via `.github/workflows/deploy.yml`. 

Ensure the repository secret `KUBECONFIG` is set in GitHub Settings -> Secrets & Variables -> Actions. On every push to `main`, the workflow validates the manifests and deploys them to the cluster.

## Troubleshooting

### TLS Certificate Issues (`cert-manager`)

If a certificate gets stuck in a pending state:

```bash
# Delete stuck certificate and secret
kubectl delete certificate kernel-panic-tls -n whale-tracker-prod
kubectl delete secret kernel-panic-tls -n whale-tracker-prod

# Re-trigger certificate issuance by re-applying ingress
kubectl apply -f backend/k8s/ingress.yaml -n whale-tracker-prod
```