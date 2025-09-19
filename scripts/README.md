# Testing Scripts

This directory contains testing scripts for the IP Geolocation Service.

## Scripts

### `test_3_clients.sh`
Tests the rate limiter with 3 specific clients, each making 50 requests.

**Usage:**
```bash
./scripts/test_3_clients.sh
```

**Features:**
- 3 different clients (192.168.1.100-102)
- 50 requests per client (150 total)
- Color-coded output for each client
- Individual client summaries
- Overall test summar

## Prerequisites

1. **Server must be running:**
   ```bash
   make run-dev
   # or
   make run-prod
   ```

2. **Required tools:**
   - `curl`
   - `jq`
   - `bash`

## How to Run

1. **Start the server:**
   ```bash
   cd /path/to/ip-geolocation-service
   make run-dev    # Development mode
   # or
   make run-prod   # Production mode
   ```

2. **Run a test script:**
   ```bash
   make test-3-clients
   # or
   ./scripts/test_3_clients.sh
   ```

3. **Other useful commands:**
   ```bash
   make help              # Show all available commands
   make test              # Run all tests
   make test-coverage     # Run tests with coverage
   make build             # Build the application
   make clean             # Clean build artifacts
   ```

## Expected Behavior

### Rate Limiter Configuration
- **Burst Size:** 20 tokens per client
- **Rate:** 20 requests per second
- **Inactive Threshold:** 5 minutes

### Test Results
- **First 20 requests per client:** Should succeed (using burst tokens)
- **Additional requests:** May be rate limited depending on timing
- **Token refill:** 1 token per 50ms (20 requests/second)

## Troubleshooting

### Server not running
```
❌ Server is not running! Please start the server first.
```
**Solution:** Start the server with `make run-dev` or `make run-prod`

### Connection failed
```
❌ Request X: CONNECTION FAILED
```
**Solution:** Check if server is running and accessible on port 8080

### Rate limit exceeded
```
🚫 Request X: RATE LIMITED
```
**Expected behavior:** This indicates the rate limiter is working correctly

## Script Output

### Colors
- 🟢 **Green:** Successful requests
- 🔴 **Red:** Rate limited or failed requests
- 🟡 **Yellow:** Warnings or unknown errors
- 🔵 **Blue:** Progress indicators
- 🟣 **Purple:** Summary information

### Progress Indicators
- Shows progress every 10-50 requests
- Displays success/failure counts
- Shows rate limiter state before and after tests
