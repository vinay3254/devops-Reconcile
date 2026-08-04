# StatusPage Operator — Phase 1

A Kubernetes-native uptime monitoring SaaS built around a custom Operator.

## Architecture

- **Operator** (`operator/`) — Go + kubebuilder. Manages `Monitor` CRDs → spawns CronJobs for HTTP checks.
- **CRD**: `Monitor` (group: `monitoring.statuspage.io/v1alpha1`)

## Local Dev Setup

### Prerequisites
- WSL2 (Ubuntu)
- Docker Desktop with WSL2 integration enabled
- Tools installed via `setup_wsl.sh`: Go 1.24.5, kind v0.27.0, kubebuilder v4.15.0

### Quick Start (inside WSL2)
```bash
# Create kind cluster
kind create cluster --name statuspage

# Install CRDs
cd operator && make install

# Run operator locally
make run

# Apply a test monitor
kubectl apply -f config/samples/monitoring_v1alpha1_monitor.yaml

# Check status
kubectl get monitors
```

## Phase Progress

- [x] Phase 1 — Core CRD + Operator (in progress)
- [ ] Phase 2 — API + Multi-Tenant Data Model
- [ ] Phase 3 — Frontend Dashboard
- [ ] Phase 4 — GitOps + CI/CD
- [ ] Phase 5 — Observability + Alerting
- [ ] Phase 6 — Production Hardening
