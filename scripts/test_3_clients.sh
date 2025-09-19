#!/bin/bash

# Test script for 3 clients with 50 requests each
# This will test rate limiting with multiple clients

echo "🚀 Testing 3 Clients with 50 Requests Each"
echo "=========================================="

# Colors for better visibility
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# 3 different client IDs
CLIENTS=(
    "192.168.1.100"
    "192.168.1.101" 
    "192.168.1.102"
)

# Function to make a request with visual indicators
make_request() {
    local client_id=$1
    local request_num=$2
    local client_index=$3
    
    # Make request
    response=$(curl -s -H "X-Forwarded-For: $client_id" \
        "http://localhost:8080/v1/find-country?ip=8.8.8.8" 2>/dev/null)
    
    # Check if request was successful
    if [ $? -eq 0 ]; then
        # Check if response contains country (success) or rate limit error
        if echo "$response" | grep -q "country"; then
            # Extract country from response
            country=$(echo "$response" | jq -r '.country // "Unknown"' 2>/dev/null)
            city=$(echo "$response" | jq -r '.city // "Unknown"' 2>/dev/null)
            
            # Color code based on client
            case $client_index in
                0) color=$GREEN ;;
                1) color=$BLUE ;;
                2) color=$YELLOW ;;
                *) color=$WHITE ;;
            esac
            
            echo -e "${color}✅ Client $((client_index+1)) Request $request_num: ${country}, ${city}${NC}"
            return 0
        else
            # Check if it's a rate limit error
            if echo "$response" | grep -q "Rate limit exceeded"; then
                echo -e "${RED}🚫 Client $((client_index+1)) Request $request_num: RATE LIMITED${NC}"
                return 1
            else
                echo -e "${YELLOW}⚠️  Client $((client_index+1)) Request $request_num: UNKNOWN ERROR${NC}"
                return 1
            fi
        fi
    else
        echo -e "${RED}❌ Client $((client_index+1)) Request $request_num: CONNECTION FAILED${NC}"
        return 1
    fi
}

# Function to check rate limiter state
check_rate_limiter() {
    echo ""
    echo -e "${CYAN}📊 Rate Limiter State:${NC}"
    echo "====================="
    
    # Get rate limiter state
    state=$(curl -s http://localhost:8080/debug/rate-limiter 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        echo "$state" | jq -r '.clients | to_entries | .[] | 
            "\(.key): \(.value.tokens) tokens (\(.value.is_active | if . then "active" else "inactive" end))"' | while read line; do
            if echo "$line" | grep -q "active"; then
                echo -e "${GREEN}$line${NC}"
            else
                echo -e "${RED}$line${NC}"
            fi
        done
    else
        echo -e "${RED}❌ Failed to get rate limiter state${NC}"
    fi
}

# Function to show client summary
show_client_summary() {
    local client_index=$1
    local client_id=$2
    local success=$3
    local failed=$4
    local total=$5
    
    echo ""
    echo -e "${WHITE}📊 Client $((client_index+1)) ($client_id) Summary:${NC}"
    echo "================================="
    echo -e "${GREEN}✅ Successful: $success${NC}"
    echo -e "${RED}❌ Failed: $failed${NC}"
    echo -e "${BLUE}📈 Success rate: $(( (success * 100) / total ))%${NC}"
}

# Check if server is running
echo -e "${CYAN}🔍 Checking if server is running...${NC}"
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo -e "${RED}❌ Server is not running! Please start the server first.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Server is running!${NC}"
echo ""

# Show initial rate limiter state
echo -e "${CYAN}📊 Initial Rate Limiter State:${NC}"
echo "=============================="
check_rate_limiter
echo ""

# Test each client
total_success=0
total_failed=0

for client_index in "${!CLIENTS[@]}"; do
    client_id="${CLIENTS[$client_index]}"
    
    echo ""
    echo -e "${WHITE}🔄 Testing Client $((client_index+1)): $client_id${NC}"
    echo "============================================="
    
    client_success=0
    client_failed=0
    
    # Make 50 requests for this client
    for i in $(seq 1 50); do
        make_request "$client_id" "$i" "$client_index"
        
        # Count results
        if [ $? -eq 0 ]; then
            ((client_success++))
            ((total_success++))
        else
            ((client_failed++))
            ((total_failed++))
        fi
        
        # Show progress every 10 requests
        if [ $((i % 10)) -eq 0 ]; then
            echo -e "${BLUE}📈 Progress: $i/50 requests completed${NC}"
        fi
        
        # Small delay
        sleep 0.05
    done
    
    # Show client summary
    show_client_summary "$client_index" "$client_id" "$client_success" "$client_failed" 50
done

# Show final summary
echo ""
echo -e "${WHITE}🎯 Overall Test Summary:${NC}"
echo "========================"
echo -e "${GREEN}✅ Total successful requests: $total_success${NC}"
echo -e "${RED}❌ Total failed requests: $total_failed${NC}"
echo -e "${BLUE}📊 Total requests: $((total_success + total_failed))${NC}"

if [ $((total_success + total_failed)) -gt 0 ]; then
    overall_success_rate=$(( (total_success * 100) / (total_success + total_failed) ))
    echo -e "${PURPLE}📈 Overall success rate: $overall_success_rate%${NC}"
fi

# Show final rate limiter state
check_rate_limiter

echo ""
echo -e "${PURPLE}🏁 Test completed!${NC}"
