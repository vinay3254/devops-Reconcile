#!/usr/bin/env bash
# Runs the operator in foreground — keep this terminal open
export GOROOT="$HOME/tools/go"
export GOPATH="$HOME/go"
export PATH="$HOME/tools/go/bin:$HOME/tools/bin:$HOME/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH"

cd /mnt/c/Users/Admin/Desktop/devops/operator

echo "=== Starting StatusPage Operator ==="
echo "Context: $(kubectl config current-context)"
echo "Watching for Monitor CRs..."
echo ""

make run
