#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🧪 Certificate Monkey - Test Suite Runner"
echo "========================================"

# Run all tests and capture results
TEST_PACKAGES=(
    "./internal/models"
    "./internal/config"
    "./internal/crypto"
    "./internal/api/middleware"
    "./internal/api/routes"
)

PASSED=0
FAILED=0
TOTAL=0

echo ""
echo "📋 Running test packages..."

for package in "${TEST_PACKAGES[@]}"; do
    echo ""
    echo "Testing $package..."
    echo "-------------------"

    if go test -v "$package"; then
        echo -e "${GREEN}✅ $package: PASSED${NC}"
        ((PASSED++))
    else
        echo -e "${RED}❌ $package: FAILED${NC}"
        ((FAILED++))
    fi
    ((TOTAL++))
done

echo ""
echo "📊 Test Summary"
echo "==============="
echo -e "Total packages: $TOTAL"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "\n🎉 ${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "\n⚠️  ${YELLOW}Some tests failed. Please check the output above.${NC}"
    exit 1
fi
