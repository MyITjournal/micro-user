#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ORCHESTRATOR_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
PROJECT_ROOT="$(cd "$ORCHESTRATOR_DIR/../.." && pwd)"

echo -e "${BLUE}🧪 Integration Test Runner for Orchestrator Service${NC}"
echo "=========================================="
echo ""

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed or not in PATH${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! command -v docker compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed or not in PATH${NC}"
    exit 1
fi

# Determine docker-compose command
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    DOCKER_COMPOSE="docker compose"
fi

# Function to cleanup on exit
cleanup() {
    echo ""
    echo -e "${YELLOW}🧹 Cleaning up test infrastructure...${NC}"
    
    # Stop test services
    cd "$PROJECT_ROOT"
    if [ -f "docker-compose.test.yml" ]; then
        $DOCKER_COMPOSE -f docker-compose.test.yml down -v 2>/dev/null || true
    fi
    
    echo -e "${GREEN}✓ Cleanup complete${NC}"
}

# Trap to ensure cleanup on exit
trap cleanup EXIT INT TERM

# Check if test infrastructure should be started
START_INFRA=true
if [ "$1" == "--no-setup" ] || [ "$1" == "-n" ]; then
    START_INFRA=false
    echo -e "${YELLOW}⚠️  Skipping infrastructure setup (using existing services)${NC}"
fi

# Start test infrastructure if needed
if [ "$START_INFRA" = true ]; then
    echo -e "${BLUE}🚀 Setting up test infrastructure...${NC}"
    
    # Create docker-compose.test.yml if it doesn't exist
    cd "$PROJECT_ROOT"
    if [ ! -f "docker-compose.test.yml" ]; then
        echo -e "${YELLOW}Creating docker-compose.test.yml...${NC}"
        cat > docker-compose.test.yml << 'EOF'
version: '3.8'

services:
  postgres-test:
    image: postgres:15-alpine
    container_name: orchestrator-postgres-test
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: orchestrator_test
    ports:
      - "5433:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5
    tmpfs:
      - /var/lib/postgresql/data
    networks:
      - test-network

  redis-test:
    image: redis:7-alpine
    container_name: orchestrator-redis-test
    ports:
      - "6380:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
    networks:
      - test-network

networks:
  test-network:
    driver: bridge
EOF
        echo -e "${GREEN}✓ Created docker-compose.test.yml${NC}"
    fi
    
    # Start services
    echo -e "${BLUE}Starting PostgreSQL and Redis...${NC}"
    $DOCKER_COMPOSE -f docker-compose.test.yml up -d
    
    # Wait for services to be ready
    echo -e "${YELLOW}⏳ Waiting for services to be ready...${NC}"
    
    # Wait for PostgreSQL
    MAX_WAIT=30
    WAIT_COUNT=0
    while ! docker exec orchestrator-postgres-test pg_isready -U postgres > /dev/null 2>&1; do
        if [ $WAIT_COUNT -ge $MAX_WAIT ]; then
            echo -e "${RED}❌ PostgreSQL failed to start within ${MAX_WAIT}s${NC}"
            exit 1
        fi
        sleep 1
        WAIT_COUNT=$((WAIT_COUNT + 1))
        echo -n "."
    done
    echo ""
    
    # Wait for Redis
    WAIT_COUNT=0
    while ! docker exec orchestrator-redis-test redis-cli ping > /dev/null 2>&1; do
        if [ $WAIT_COUNT -ge $MAX_WAIT ]; then
            echo -e "${RED}❌ Redis failed to start within ${MAX_WAIT}s${NC}"
            exit 1
        fi
        sleep 1
        WAIT_COUNT=$((WAIT_COUNT + 1))
        echo -n "."
    done
    echo ""
    
    echo -e "${GREEN}✓ Test infrastructure ready${NC}"
    echo ""
    
    # Set environment variables for test services
    export TEST_DB_HOST=localhost
    export TEST_DB_PORT=5433
    export TEST_DB_USER=postgres
    export TEST_DB_PASSWORD=postgres
    export TEST_DB_NAME=orchestrator_test
    
    export TEST_REDIS_HOST=localhost
    export TEST_REDIS_PORT=6380
    export TEST_REDIS_DB=1
else
    # Use environment variables if set, otherwise use defaults
    export TEST_DB_HOST=${TEST_DB_HOST:-localhost}
    export TEST_DB_PORT=${TEST_DB_PORT:-5432}
    export TEST_DB_USER=${TEST_DB_USER:-postgres}
    export TEST_DB_PASSWORD=${TEST_DB_PASSWORD:-postgres}
    export TEST_DB_NAME=${TEST_DB_NAME:-orchestrator_test}
    
    export TEST_REDIS_HOST=${TEST_REDIS_HOST:-localhost}
    export TEST_REDIS_PORT=${TEST_REDIS_PORT:-6379}
    export TEST_REDIS_DB=${TEST_REDIS_DB:-1}
    
    echo -e "${YELLOW}Using existing services:${NC}"
    echo "  PostgreSQL: ${TEST_DB_HOST}:${TEST_DB_PORT}"
    echo "  Redis: ${TEST_REDIS_HOST}:${TEST_REDIS_PORT}"
    echo ""
fi

# Kafka is optional (uses mock by default)
export TEST_KAFKA_BROKERS=${TEST_KAFKA_BROKERS:-""}

# Change to orchestrator directory
cd "$ORCHESTRATOR_DIR"

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed or not in PATH${NC}"
    exit 1
fi

# Run the integration tests
echo -e "${BLUE}🧪 Running Go integration tests...${NC}"
echo ""

# Determine which tests to run
TEST_PATTERN="./tests/integration/..."
if [ "$1" != "--no-setup" ] && [ "$1" != "-n" ]; then
    # If first arg is not a flag, treat it as a test pattern
    if [ -n "$1" ] && [[ ! "$1" =~ ^- ]]; then
        TEST_PATTERN="$1"
    fi
fi

# Run tests with verbose output
if go test -v "$TEST_PATTERN" -count=1; then
    echo ""
    echo -e "${GREEN}✅ All integration tests passed!${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}❌ Some integration tests failed${NC}"
    exit 1
fi

