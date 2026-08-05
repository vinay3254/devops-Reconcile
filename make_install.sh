#!/usr/bin/env bash
set -e

# Set up full PATH with all tools
export GOROOT="$HOME/tools/go"
export GOPATH="$HOME/go"
export PATH="$HOME/tools/go/bin:$HOME/tools/bin:$HOME/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH"

echo "=== Environment ==="
echo "Go: $(go version)"
echo "make: $(make --version | head -1)"
echo "kubectl: $(kubectl version --client --short 2>/dev/null || kubectl version --client)"
echo "kind clusters: $(kind get clusters)"
echo ""

cd /mnt/c/Users/Admin/Desktop/devops/operator

echo "=== make manifests (regenerate CRD YAML from types) ==="
make manifests
echo "MANIFESTS_OK"

echo ""
echo "=== make install (apply CRDs to kind cluster) ==="
make install
echo "CRDs_INSTALLED_OK"

echo ""
echo "=== Verify CRDs installed ==="
kubectl get crd monitors.monitoring.statuspage.io
