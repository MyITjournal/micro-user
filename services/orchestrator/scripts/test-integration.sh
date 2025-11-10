#!/bin/bash
set -e

echo "🧪 Integration Test for Orchestrator Service"
echo "==========================================="
echo ""

BASE_URL="http://localhost:8080"
FAILED=0

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Start service if not running
if ! curl -s "$BASE_URL/health" > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Service not running. Starting...${NC}"
    make run &
    SERVICE_PID=$!
    sleep 3
fi

test_endpoint() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local expected_status=$5

    echo -n "Testing: $name ... "

    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" -eq "$expected_status" ]; then
        echo -e "${GREEN}✓ PASS${NC} (HTTP $http_code)"
        return 0
    else
        echo -e "${RED}✗ FAIL${NC} (Expected: $expected_status, Got: $http_code)"
        echo "Response: $body"
        FAILED=$((FAILED + 1))
        return 1
    fi
}

echo "Health Checks:"
test_endpoint "Simple health" GET "/health" "" 200
test_endpoint "Liveness probe" GET "/health/live" "" 200
test_endpoint "Readiness probe" GET "/health/ready" "" 200

echo ""
echo "Notification API:"

# Valid notification request
test_endpoint "Create email notification" POST "/api/v1/notifications" \
    '{"user_id":"usr_123","template_id":"welcome_email","channel":"email","variables":{"user_name":"John Doe","app_name":"MyApp"}}' \
    201

# Valid push notification
test_endpoint "Create push notification" POST "/api/v1/notifications" \
    '{"user_id":"usr_123","template_id":"welcome_email","channel":"push","variables":{"user_name":"Jane Smith","app_name":"MyApp"}}' \
    201

# Invalid request (missing required field)
test_endpoint "Invalid request (missing user_id)" POST "/api/v1/notifications" \
    '{"template_id":"welcome_email","channel":"email"}' \
    400

# Invalid channel
test_endpoint "Invalid channel" POST "/api/v1/notifications" \
    '{"user_id":"usr_123","template_id":"welcome_email","channel":"invalid"}' \
    400

echo ""
echo "==========================================="

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    if [ ! -z "$SERVICE_PID" ]; then
        kill $SERVICE_PID 2>/dev/null || true
    fi
    exit 0
else
    echo -e "${RED}❌ $FAILED test(s) failed${NC}"
    if [ ! -z "$SERVICE_PID" ]; then
        kill $SERVICE_PID 2>/dev/null || true
    fi
    exit 1
fi