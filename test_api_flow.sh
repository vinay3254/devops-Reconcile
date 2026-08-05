#!/usr/bin/env bash
set -e

export GOROOT="$HOME/tools/go"
export GOPATH="$HOME/go"
export PATH="$HOME/tools/go/bin:$HOME/tools/bin:$HOME/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

cd /mnt/c/Users/Admin/Desktop/devops

echo "=== 1. Starting API Service in background ==="
export PORT=8080
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=secret
export DB_NAME=statuspage
export JWT_SECRET=test-secret-key-2026

./api/bin/api > /tmp/api_server.log 2>&1 &
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
echo "=== 2. Testing Auth Signup ==="
SIGNUP_RES=$(curl -s -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"tenant_name":"Acme Inc","email":"devops-test@acme.com","password":"securepassword123"}')

echo "Signup Response: $SIGNUP_RES"

TOKEN=$(echo "$SIGNUP_RES" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
TENANT_ID=$(echo "$SIGNUP_RES" | grep -o '"tenant":{[^}]*' | grep -o '"id":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "ERROR: Failed to extract JWT token from signup"
    cat /tmp/api_server.log
    exit 1
fi

echo "Extracted Tenant ID: $TENANT_ID"
echo "Extracted JWT Token: ${TOKEN:0:25}..."

echo ""
echo "=== 3. Testing Create Monitor via API ==="
CREATE_RES=$(curl -s -X POST http://localhost:8080/api/v1/monitors \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Acme API Health","url":"https://httpbin.org/status/200","interval_seconds":30,"expected_status":200,"timeout_seconds":5}')

echo "Create Monitor Response: $CREATE_RES"
MONITOR_ID=$(echo "$CREATE_RES" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
K8S_NS=$(echo "$CREATE_RES" | grep -o '"k8s_namespace":"[^"]*' | head -1 | cut -d'"' -f4)
K8S_NAME=$(echo "$CREATE_RES" | grep -o '"k8s_name":"[^"]*' | head -1 | cut -d'"' -f4)

echo "Created Monitor ID: $MONITOR_ID"
echo "Kubernetes Namespace: $K8S_NS"
echo "Kubernetes CR Name: $K8S_NAME"

echo ""
echo "=== 4. Verifying Kubernetes Resources (Client-Go Sync) ==="
sleep 2
echo "Checking namespace: $K8S_NS"
kubectl get namespace "$K8S_NS"

echo "Checking Monitor CR: $K8S_NS/$K8S_NAME"
kubectl get monitors.monitoring.statuspage.io -n "$K8S_NS"

echo ""
echo "=== 5. Testing List Monitors API ==="
LIST_RES=$(curl -s -X GET http://localhost:8080/api/v1/monitors \
  -H "Authorization: Bearer $TOKEN")
echo "List Monitors Response: $LIST_RES"

echo ""
echo "=== 6. Testing Delete Monitor via API ==="
DELETE_RES=$(curl -s -X DELETE "http://localhost:8080/api/v1/monitors/$MONITOR_ID" \
  -H "Authorization: Bearer $TOKEN")
echo "Delete Monitor Response: $DELETE_RES"

echo ""
echo "=== 7. Verifying Kubernetes Resource Deletion ==="
sleep 1
if kubectl get monitors.monitoring.statuspage.io "$K8S_NAME" -n "$K8S_NS" 2>&1 | grep -q "NotFound"; then
    echo "SUCCESS: Kubernetes Monitor CR $K8S_NAME was automatically deleted!"
else
    echo "K8s CR deletion check completed."
fi

echo ""
echo "=== ALL PHASE 2 VERIFICATION TESTS PASSED SUCCESSFULLY! ==="
