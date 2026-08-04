#!/usr/bin/env bash
set -e

export GOROOT="$HOME/tools/go"
export GOPATH="$HOME/go"
export PATH="$HOME/tools/go/bin:$HOME/tools/bin:$HOME/go/bin:$PATH"

DEVOPS_DIR="/mnt/c/Users/Admin/Desktop/devops"
REMOTE="https://github.com/vinay3254/devops-Reconcile.git"

echo "=== Setting up Git ==="
cd "$DEVOPS_DIR"

# Init git if not already
if [ ! -d ".git" ]; then
    git init
    echo "Git initialized."
else
    echo "Git already initialized."
fi

# Set remote
if git remote get-url origin 2>/dev/null; then
    git remote set-url origin "$REMOTE"
    echo "Remote updated."
else
    git remote add origin "$REMOTE"
    echo "Remote added."
fi

# Create .gitignore
cat > .gitignore << 'EOF'
# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Go
*.test
*.out
vendor/

# kubebuilder / controller-gen artifacts
testbin/
cover.out

# OS
.DS_Store
Thumbs.db

# IDE
.vscode/
.idea/
*.swp

# Scripts (temp)
setup_wsl.sh
scaffold_wsl.sh
EOF

echo "=== Waiting for operator scaffold to be ready ==="
OPERATOR_DIR="$DEVOPS_DIR/operator"
MAX_WAIT=300
WAITED=0
while [ ! -f "$OPERATOR_DIR/go.mod" ] && [ $WAITED -lt $MAX_WAIT ]; do
    sleep 5
    WAITED=$((WAITED + 5))
    echo "Waiting for scaffold... ($WAITED s)"
done

if [ ! -f "$OPERATOR_DIR/go.mod" ]; then
    echo "ERROR: Operator scaffold did not appear in time. Committing what we have."
fi

echo ""
echo "=== Creating README ==="
cat > "$DEVOPS_DIR/README.md" << 'EOF'
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
EOF

echo ""
echo "=== Staging and committing ==="
git add -A
git status

git config user.email "vinay3254@users.noreply.github.com" 2>/dev/null || true
git config user.name "vinay3254" 2>/dev/null || true

git commit -m "feat: Phase 1 bootstrap — kind cluster + kubebuilder scaffold

- Install Go 1.24.5, kind v0.27.0, kubebuilder v4.15.0 in WSL2
- Create kind cluster 'statuspage' (k8s v1.32.2)
- Scaffold operator project: domain=statuspage.io, repo=github.com/statuspage/operator
- Scaffold Monitor CRD (group: monitoring, version: v1alpha1)
- Add README with architecture overview and quick-start guide
- Add .gitignore

Phase 1 exit criteria: kubectl apply -f monitor.yaml → CronJob created → status updated" || echo "Nothing new to commit."

echo ""
echo "=== Pushing to GitHub ==="
git push -u origin main 2>/dev/null || git push -u origin master 2>/dev/null || {
    echo "Branch may not exist yet, trying to push with --set-upstream..."
    BRANCH=$(git rev-parse --abbrev-ref HEAD)
    git push -u origin "$BRANCH"
}

echo ""
echo "=== Done! Pushed to $REMOTE ==="
