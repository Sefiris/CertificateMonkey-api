#!/bin/bash

# Certificate Monkey - Swagger UI Demo Script

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m'

PORT="${PORT:-8080}"
HOST="${HOST:-localhost}"

echo -e "${BLUE}🚀 Certificate Monkey - Swagger UI Demo${NC}"
echo -e "${BLUE}====================================${NC}"
echo ""

echo -e "${CYAN}📝 Setting up environment...${NC}"

# Set minimal environment variables for demo
export SERVER_HOST="$HOST"
export SERVER_PORT="$PORT"
export AWS_REGION="us-east-1"
export DYNAMODB_TABLE="certificate-monkey-demo"
export KMS_KEY_ID="arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
export API_KEYS="demo-api-key-12345,swagger-test-key"

echo -e "${GREEN}✅ Environment configured:${NC}"
echo -e "   Server: http://$HOST:$PORT"
echo -e "   Swagger UI: http://$HOST:$PORT/swagger/index.html"
echo -e "   API Keys: demo-api-key-12345, swagger-test-key"
echo ""

echo -e "${CYAN}🔧 Building application...${NC}"
if ! go build -o certificate-monkey cmd/server/main.go; then
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Build successful${NC}"
echo ""

echo -e "${CYAN}🌐 Starting Certificate Monkey server...${NC}"
echo -e "${YELLOW}💡 Press Ctrl+C to stop the server${NC}"
echo ""

echo -e "${CYAN}📖 Available endpoints:${NC}"
echo -e "   Health: ${BLUE}http://$HOST:$PORT/health${NC}"
echo -e "   Swagger UI: ${BLUE}http://$HOST:$PORT/swagger/index.html${NC}"
echo -e "   API Base: ${BLUE}http://$HOST:$PORT/api/v1${NC}"
echo ""

echo -e "${CYAN}🔑 Authentication:${NC}"
echo -e "   Header: X-API-Key: demo-api-key-12345"
echo -e "   Bearer: Authorization: Bearer demo-api-key-12345"
echo ""

echo -e "${CYAN}📋 Quick Test Commands:${NC}"
echo -e "   Health Check:"
echo -e "   ${YELLOW}curl http://$HOST:$PORT/health${NC}"
echo ""
echo -e "   List Certificates:"
echo -e "   ${YELLOW}curl -H 'X-API-Key: demo-api-key-12345' http://$HOST:$PORT/api/v1/keys${NC}"
echo ""

echo -e "${GREEN}🚀 Starting server...${NC}"
echo ""

# Start the server
./certificate-monkey
