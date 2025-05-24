#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="http://localhost:8080"
API_KEY="test-api-key-12345"

echo -e "${BLUE}🔍 Certificate Monkey - Tag Search Demo${NC}"
echo -e "${BLUE}=====================================${NC}"
echo ""

# Function to check if server is running
check_server() {
    echo -e "${CYAN}📡 Checking if Certificate Monkey server is running...${NC}"
    if ! curl -s "$API_BASE_URL/health" > /dev/null; then
        echo -e "${RED}❌ Server is not running at $API_BASE_URL${NC}"
        echo -e "${YELLOW}💡 Please start the server with: go run cmd/server/main.go${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Server is running${NC}"
    echo ""
}

# Function to create test certificates with different tags
create_test_certificates() {
    echo -e "${CYAN}🏗️  Creating test certificates with various tags...${NC}"

    # Certificate 1: Production environment
    echo -e "  Creating production certificate..."
    curl -s -X POST "$API_BASE_URL/api/v1/keys" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{
            "common_name": "prod-api.example.com",
            "organization": "ACME Corp",
            "key_type": "RSA2048",
            "tags": {
                "environment": "production",
                "project": "api-gateway",
                "team": "platform",
                "cost-center": "IT-001"
            }
        }' > /dev/null

    # Certificate 2: Development environment
    echo -e "  Creating development certificate..."
    curl -s -X POST "$API_BASE_URL/api/v1/keys" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{
            "common_name": "dev-api.example.com",
            "organization": "ACME Corp",
            "key_type": "ECDSA-P256",
            "tags": {
                "environment": "development",
                "project": "api-gateway",
                "team": "platform",
                "cost-center": "IT-001"
            }
        }' > /dev/null

    # Certificate 3: Different project
    echo -e "  Creating web-server certificate..."
    curl -s -X POST "$API_BASE_URL/api/v1/keys" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{
            "common_name": "web.example.com",
            "organization": "ACME Corp",
            "key_type": "RSA4096",
            "tags": {
                "environment": "production",
                "project": "web-server",
                "team": "frontend",
                "cost-center": "IT-002"
            }
        }' > /dev/null

    # Certificate 4: Staging environment
    echo -e "  Creating staging certificate..."
    curl -s -X POST "$API_BASE_URL/api/v1/keys" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{
            "common_name": "staging.example.com",
            "organization": "ACME Corp",
            "key_type": "ECDSA-P384",
            "tags": {
                "environment": "staging",
                "project": "web-server",
                "team": "qa",
                "temporary": "true"
            }
        }' > /dev/null

    echo -e "${GREEN}✅ Test certificates created${NC}"
    echo ""
}

# Function to perform tag searches
test_tag_searches() {
    echo -e "${CYAN}🔎 Testing tag-based searches...${NC}"
    echo ""

    # Test 1: Search by environment
    echo -e "${YELLOW}Test 1: Search for production certificates${NC}"
    echo -e "Query: ${BLUE}?environment=production${NC}"
    result=$(curl -s "$API_BASE_URL/api/v1/keys?environment=production" \
        -H "X-API-Key: $API_KEY")

    count=$(echo "$result" | grep -o '"common_name"' | wc -l)
    echo -e "Results: ${GREEN}$count certificates found${NC}"
    if [[ $count -gt 0 ]]; then
        echo "$result" | grep -o '"common_name":"[^"]*"' | sed 's/"common_name":"/  - /' | sed 's/"$//'
    fi
    echo ""

    # Test 2: Search by project
    echo -e "${YELLOW}Test 2: Search for api-gateway project${NC}"
    echo -e "Query: ${BLUE}?project=api-gateway${NC}"
    result=$(curl -s "$API_BASE_URL/api/v1/keys?project=api-gateway" \
        -H "X-API-Key: $API_KEY")

    count=$(echo "$result" | grep -o '"common_name"' | wc -l)
    echo -e "Results: ${GREEN}$count certificates found${NC}"
    if [[ $count -gt 0 ]]; then
        echo "$result" | grep -o '"common_name":"[^"]*"' | sed 's/"common_name":"/  - /' | sed 's/"$//'
    fi
    echo ""

    # Test 3: Search by team
    echo -e "${YELLOW}Test 3: Search for platform team certificates${NC}"
    echo -e "Query: ${BLUE}?team=platform${NC}"
    result=$(curl -s "$API_BASE_URL/api/v1/keys?team=platform" \
        -H "X-API-Key: $API_KEY")

    count=$(echo "$result" | grep -o '"common_name"' | wc -l)
    echo -e "Results: ${GREEN}$count certificates found${NC}"
    if [[ $count -gt 0 ]]; then
        echo "$result" | grep -o '"common_name":"[^"]*"' | sed 's/"common_name":"/  - /' | sed 's/"$//'
    fi
    echo ""

    # Test 4: Multiple tag search
    echo -e "${YELLOW}Test 4: Search for production AND web-server project${NC}"
    echo -e "Query: ${BLUE}?environment=production&project=web-server${NC}"
    result=$(curl -s "$API_BASE_URL/api/v1/keys?environment=production&project=web-server" \
        -H "X-API-Key: $API_KEY")

    count=$(echo "$result" | grep -o '"common_name"' | wc -l)
    echo -e "Results: ${GREEN}$count certificates found${NC}"
    if [[ $count -gt 0 ]]; then
        echo "$result" | grep -o '"common_name":"[^"]*"' | sed 's/"common_name":"/  - /' | sed 's/"$//'
    fi
    echo ""

    # Test 5: Combined with other filters
    echo -e "${YELLOW}Test 5: Search for production + RSA keys${NC}"
    echo -e "Query: ${BLUE}?environment=production&key_type=RSA2048${NC}"
    result=$(curl -s "$API_BASE_URL/api/v1/keys?environment=production&key_type=RSA2048" \
        -H "X-API-Key: $API_KEY")

    count=$(echo "$result" | grep -o '"common_name"' | wc -l)
    echo -e "Results: ${GREEN}$count certificates found${NC}"
    if [[ $count -gt 0 ]]; then
        echo "$result" | grep -o '"common_name":"[^"]*"' | sed 's/"common_name":"/  - /' | sed 's/"$//'
    fi
    echo ""

    # Test 6: Custom tag search
    echo -e "${YELLOW}Test 6: Search for temporary certificates${NC}"
    echo -e "Query: ${BLUE}?temporary=true${NC}"
    result=$(curl -s "$API_BASE_URL/api/v1/keys?temporary=true" \
        -H "X-API-Key: $API_KEY")

    count=$(echo "$result" | grep -o '"common_name"' | wc -l)
    echo -e "Results: ${GREEN}$count certificates found${NC}"
    if [[ $count -gt 0 ]]; then
        echo "$result" | grep -o '"common_name":"[^"]*"' | sed 's/"common_name":"/  - /' | sed 's/"$//'
    fi
    echo ""
}

# Function to show all certificates with their tags
show_all_certificates() {
    echo -e "${CYAN}📋 All certificates with tags:${NC}"
    echo ""

    result=$(curl -s "$API_BASE_URL/api/v1/keys" -H "X-API-Key: $API_KEY")

    # Parse and display (basic parsing without jq)
    echo "$result" | grep -E '"(common_name|tags)"' | sed 's/.*"common_name":"/📜 /' | sed 's/".*//' | head -10
    echo ""
}

# Function to demonstrate API usage examples
show_usage_examples() {
    echo -e "${CYAN}💡 Tag Search Usage Examples:${NC}"
    echo ""
    echo -e "${YELLOW}1. Search by single tag:${NC}"
    echo -e "   curl '${API_BASE_URL}/api/v1/keys?environment=production' \\"
    echo -e "     -H 'X-API-Key: your-api-key'"
    echo ""

    echo -e "${YELLOW}2. Search by multiple tags:${NC}"
    echo -e "   curl '${API_BASE_URL}/api/v1/keys?environment=production&team=platform' \\"
    echo -e "     -H 'X-API-Key: your-api-key'"
    echo ""

    echo -e "${YELLOW}3. Combine with other filters:${NC}"
    echo -e "   curl '${API_BASE_URL}/api/v1/keys?environment=production&status=CERT_UPLOADED&key_type=RSA2048' \\"
    echo -e "     -H 'X-API-Key: your-api-key'"
    echo ""

    echo -e "${YELLOW}4. Any custom tag:${NC}"
    echo -e "   curl '${API_BASE_URL}/api/v1/keys?cost-center=IT-001&owner=john.doe' \\"
    echo -e "     -H 'X-API-Key: your-api-key'"
    echo ""

    echo -e "${YELLOW}5. With pagination:${NC}"
    echo -e "   curl '${API_BASE_URL}/api/v1/keys?environment=production&page=1&page_size=10' \\"
    echo -e "     -H 'X-API-Key: your-api-key'"
    echo ""
}

# Function to measure performance
test_performance() {
    echo -e "${CYAN}⚡ Performance Test (current implementation):${NC}"
    echo ""

    start_time=$(date +%s%N)
    result=$(curl -s "$API_BASE_URL/api/v1/keys?environment=production" \
        -H "X-API-Key: $API_KEY")
    end_time=$(date +%s%N)

    duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
    count=$(echo "$result" | grep -o '"common_name"' | wc -l)

    echo -e "Query: ?environment=production"
    echo -e "Results: ${GREEN}$count certificates${NC}"
    echo -e "Time: ${YELLOW}${duration}ms${NC}"
    echo ""
    echo -e "${YELLOW}💡 Note: Performance will degrade as table size grows${NC}"
    echo -e "${YELLOW}   See docs/TAG_SEARCH_OPTIMIZATION.md for scaling strategies${NC}"
    echo ""
}

# Main execution
main() {
    check_server
    create_test_certificates
    test_tag_searches
    show_all_certificates
    test_performance
    show_usage_examples

    echo -e "${GREEN}🎉 Tag search demo completed!${NC}"
    echo -e "${CYAN}📖 For optimization strategies, see: docs/TAG_SEARCH_OPTIMIZATION.md${NC}"
}

# Handle script interruption
trap 'echo -e "\n${RED}❌ Script interrupted${NC}"; exit 1' INT TERM

# Run main function
main
