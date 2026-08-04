#!/usr/bin/env bash
set -e

export GOROOT="$HOME/tools/go"
export GOPATH="$HOME/go"
export PATH="$HOME/tools/go/bin:$HOME/tools/bin:$HOME/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

cd /mnt/c/Users/Admin/Desktop/devops

echo "=== 1. Rebuilding API with Prometheus & Notifier ==="
bash build_api.sh

echo "=== 2. Starting API Service ==="
export PORT=8080
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=secret
export DB_NAME=statuspage
export JWT_SECRET=test-secret-key-2026

./api/bin/api > /tmp/api_phase5.log 2>&1 &
API_PID=$!

cleanup() {
    echo "Cleaning up API process $API_PID..."
    kill $API_PID 2>/dev/null || true
}
trap cleanup EXIT

echo "Waiting for API to respond on port 8080..."
for i in {1..30}; do
    if curl -s http://localhost:8080/healthz | grep -q "UP"; then
        echo "API service is UP!"
        break
    fi
    sleep 1
done

echo ""
echo "=== 3. Testing Prometheus /metrics Endpoint ==="
METRICS_OUT=$(curl -s http://localhost:8080/metrics)
if echo "$METRICS_OUT" | grep -q "statuspage_api_requests_total"; then
    echo "SUCCESS: Prometheus custom metric statuspage_api_requests_total found!"
else
    echo "ERROR: Metrics endpoint output missing expected metrics"
    exit 1
fi

echo ""
echo "=== 4. Testing Webhook Alert Registration API ==="
SIGNUP_RES=$(curl -s -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"tenant_name":"AlertCorp","email":"alert-admin@acme.com","password":"securepassword123"}')

TOKEN=$(echo "$SIGNUP_RES" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

HOOK_RES=$(curl -s -X POST http://localhost:8080/api/v1/alerts/webhooks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://httpbin.org/post","secret":"webhooksecret123"}')

echo "Webhook Registration Response: $HOOK_RES"

if echo "$HOOK_RES" | grep -q "httpbin.org"; then
    echo "SUCCESS: Webhook subscription registered successfully!"
else
    echo "ERROR: Failed to register webhook subscription"
    cat /tmp/api_phase5.log
    exit 1
fi

echo ""
echo "=== ALL PHASE 5 OBSERVABILITY & ALERTING TESTS PASSED! ==="
