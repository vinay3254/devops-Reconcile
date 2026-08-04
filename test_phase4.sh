#!/usr/bin/env bash
set -e

export GOROOT="$HOME/tools/go"
export GOPATH="$HOME/go"
export PATH="$HOME/tools/go/bin:$HOME/tools/bin:$HOME/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

cd /mnt/c/Users/Admin/Desktop/devops

echo "=== 1. Ensuring Helm is installed ==="
if ! command -v helm &>/dev/null; then
    echo "Downloading helm v3.14.0..."
    curl -fsSL https://get.helm.sh/helm-v3.14.0-linux-amd64.tar.gz | tar -xz -C /tmp
    mkdir -p "$HOME/tools/bin"
    mv /tmp/linux-amd64/helm "$HOME/tools/bin/helm"
    echo "Helm installed."
fi
helm version

echo ""
echo "=== 2. Linting Helm Charts ==="
helm lint charts/statuspage-operator
helm lint charts/statuspage-api
helm lint charts/statuspage-frontend
helm lint charts/statuspage-stack

echo ""
echo "=== 3. Building Chart Dependencies & Testing Helm Dry-Run Rendering ==="
helm dependency build charts/statuspage-stack
helm template statuspage-test charts/statuspage-stack > /tmp/stack-rendered.yaml
echo "Successfully rendered $(wc -l < /tmp/stack-rendered.yaml) lines of Kubernetes manifests!"

echo ""
echo "=== 4. Testing Docker Builds ==="
docker build --network=host -t statuspage-api:v1.0.0 ./api
docker build --network=host -t statuspage-frontend:v1.0.0 ./frontend

echo ""
echo "=== ALL PHASE 4 GITOPS & CI/CD VERIFICATIONS PASSED! ==="
