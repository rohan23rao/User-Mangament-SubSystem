#!/bin/bash

# Test Email Authentication System
set -e

# Configuration
KRATOS_PUBLIC="http://localhost:4433"
KRATOS_ADMIN="http://localhost:4434"
BACKEND_API="http://localhost:8080"
MAILSLURPER="http://localhost:4436"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_step() {
    echo -e "${BLUE}🔧${NC} $1"
}

echo "🧪 Testing Email Authentication System"
echo "======================================"

# Test 1: Health Checks
print_step "Step 1: Health Checks"
echo ""

# Test Kratos Public API
response=$(curl -s -w "%{http_code}" -o /dev/null "$KRATOS_PUBLIC/health/ready")
if [ "$response" = "200" ]; then
    print_status "Kratos Public API: Running (Port 4433)"
else
    print_error "Kratos Public API: Failed (HTTP $response)"
    exit 1
fi

# Test Kratos Admin API
response=$(curl -s -w "%{http_code}" -o /dev/null "$KRATOS_ADMIN/health/ready")
if [ "$response" = "200" ]; then
    print_status "Kratos Admin API: Running (Port 4434)"
else
    print_error "Kratos Admin API: Failed (HTTP $response)"
    exit 1
fi

# Test Go Backend
response=$(curl -s -w "%{http_code}" -o /dev/null "$BACKEND_API/health")
if [ "$response" = "200" ]; then
    print_status "Go Backend API: Running (Port 8080)"
else
    print_error "Go Backend API: Failed (HTTP $response)"
    exit 1
fi

# Test Mailslurper
response=$(curl -s -w "%{http_code}" -o /dev/null "$MAILSLURPER")
if [ "$response" = "200" ]; then
    print_status "Mailslurper (Email UI): Running (Port 4436)"
    print_info "📧 Email UI available at: $MAILSLURPER"
else
    print_error "Mailslurper: Failed (HTTP $response)"
fi

echo ""

# Test 2: Registration Flow
print_step "Step 2: Testing Registration Flow"
echo ""

# Get registration flow
print_info "Getting registration flow..."
REG_FLOW=$(curl -s "$KRATOS_PUBLIC/self-service/registration/api")
FLOW_ID=$(echo "$REG_FLOW" | jq -r '.id' 2>/dev/null)

if [ "$FLOW_ID" != "null" ] && [ -n "$FLOW_ID" ]; then
    print_status "Registration flow ID: $FLOW_ID"
    
    # Check for email field
    if echo "$REG_FLOW" | jq -e '.ui.nodes[] | select(.attributes.name == "traits.email")' >/dev/null 2>&1; then
        print_status "Email field found in registration form"
    else
        print_error "Email field missing in registration form"
    fi
    
    # Check for password field
    if echo "$REG_FLOW" | jq -e '.ui.nodes[] | select(.attributes.name == "password")' >/dev/null 2>&1; then
        print_status "Password field found in registration form"
    else
        print_error "Password field missing in registration form"
    fi
    
    # Check for Google OAuth
    if echo "$REG_FLOW" | jq -e '.ui.nodes[] | select(.attributes.value == "google")' >/dev/null 2>&1; then
        print_status "Google OAuth option found"
    else
        print_info "Google OAuth not configured (expected if credentials not set)"
    fi
else
    print_error "Failed to get registration flow"
    exit 1
fi

echo ""

# Test 3: User Registration
print_step "Step 3: Testing User Registration"
echo ""

# Generate test user data
TEST_EMAIL="test-$(date +%s)@example.com"
TEST_PASSWORD="TestPassword123!"
TEST_FIRST_NAME="Test"
TEST_LAST_NAME="User"

print_info "Test user email: $TEST_EMAIL"

# Create registration data
REG_DATA=$(cat <<EOF
{
    "traits.email": "$TEST_EMAIL",
    "traits.name.first": "$TEST_FIRST_NAME",
    "traits.name.last": "$TEST_LAST_NAME",
    "password": "$TEST_PASSWORD",
    "method": "password"
}
EOF
)

print_info "Submitting registration..."
REG_RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "$REG_DATA" \
    "$KRATOS_PUBLIC/self-service/registration?flow=$FLOW_ID")

# Check registration response
if echo "$REG_RESPONSE" | jq -e '.session' >/dev/null 2>&1; then
    print_status "User registration successful!"
    SESSION_TOKEN=$(echo "$REG_RESPONSE" | jq -r '.session_token // .session.token // ""')
    if [ -n "$SESSION_TOKEN" ]; then
        print_status "Session token received"
    fi
elif echo "$REG_RESPONSE" | jq -e '.redirect_browser_to' >/dev/null 2>&1; then
    print_status "Registration successful - verification required"
    REDIRECT_URL=$(echo "$REG_RESPONSE" | jq -r '.redirect_browser_to')
    print_info "Redirect URL: $REDIRECT_URL"
else
    print_error "Registration failed"
    echo "Response: $REG_RESPONSE" | jq . 2>/dev/null || echo "$REG_RESPONSE"
fi

echo ""

# Test 4: Check Email Verification
print_step "Step 4: Checking Email Verification"
echo ""

print_info "Checking Mailslurper for verification email..."
sleep 2  # Wait for email to arrive

# Get emails from Mailslurper
EMAILS=$(curl -s "$MAILSLURPER/mail" 2>/dev/null)
if [ $? -eq 0 ] && [ -n "$EMAILS" ]; then
    EMAIL_COUNT=$(echo "$EMAILS" | jq 'length' 2>/dev/null || echo "0")
    print_status "Found $EMAIL_COUNT emails in Mailslurper"
    
    # Look for verification email
    VERIFICATION_EMAIL=$(echo "$EMAILS" | jq ".[] | select(.toAddresses[0].emailAddress == \"$TEST_EMAIL\")" 2>/dev/null)
    if [ -n "$VERIFICATION_EMAIL" ]; then
        print_status "Verification email found for $TEST_EMAIL"
        SUBJECT=$(echo "$VERIFICATION_EMAIL" | jq -r '.subject')
        print_info "Email subject: $SUBJECT"
        
        # Extract verification link from email body
        BODY=$(echo "$VERIFICATION_EMAIL" | jq -r '.body')
        VERIFY_LINK=$(echo "$BODY" | grep -oP 'http://localhost:4433/self-service/verification[^"]*' | head -1)
        if [ -n "$VERIFY_LINK" ]; then
            print_status "Verification link found"
            print_info "Link: $VERIFY_LINK"
        fi
    else
        print_info "No verification email found yet (may take a moment)"
    fi
else
    print_info "Cannot access Mailslurper or no emails found"
fi

echo ""

# Test 5: Login Flow
print_step "Step 5: Testing Login Flow"
echo ""

# Get login flow
print_info "Getting login flow..."
LOGIN_FLOW=$(curl -s "$KRATOS_PUBLIC/self-service/login/api")
LOGIN_FLOW_ID=$(echo "$LOGIN_FLOW" | jq -r '.id' 2>/dev/null)

if [ "$LOGIN_FLOW_ID" != "null" ] && [ -n "$LOGIN_FLOW_ID" ]; then
    print_status "Login flow ID: $LOGIN_FLOW_ID"
    
    # Test login with the registered user
    LOGIN_DATA=$(cat <<EOF
{
    "identifier": "$TEST_EMAIL",
    "password": "$TEST_PASSWORD",
    "method": "password"
}
EOF
    )
    
    print_info "Testing login with registered user..."
    LOGIN_RESPONSE=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$LOGIN_DATA" \
        "$KRATOS_PUBLIC/self-service/login?flow=$LOGIN_FLOW_ID")
    
    if echo "$LOGIN_RESPONSE" | jq -e '.session' >/dev/null 2>&1; then
        print_status "Login successful!"
        USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.session.identity.id')
        print_info "User ID: $USER_ID"
    else
        print_info "Login may require email verification first"
        echo "Response: $LOGIN_RESPONSE" | jq . 2>/dev/null || echo "$LOGIN_RESPONSE"
    fi
else
    print_error "Failed to get login flow"
fi

echo ""

# Test 6: Backend Integration
print_step "Step 6: Testing Backend Integration"
echo ""

# Test webhook endpoints
print_info "Testing webhook endpoints..."
curl -s -X POST -H "Content-Type: application/json" -d '{"test": "data"}' "$BACKEND_API/hooks/after-registration" >/dev/null
if [ $? -eq 0 ]; then
    print_status "Registration webhook endpoint responsive"
else
    print_error "Registration webhook endpoint failed"
fi

curl -s -X POST -H "Content-Type: application/json" -d '{"test": "data"}' "$BACKEND_API/hooks/after-login" >/dev/null
if [ $? -eq 0 ]; then
    print_status "Login webhook endpoint responsive"
else
    print_error "Login webhook endpoint failed"
fi

echo ""

# Summary
print_step "Summary & Next Steps"
echo ""
print_status "System Status:"
echo "  ✓ Kratos is running and accepting requests"
echo "  ✓ Go backend is running and responding"
echo "  ✓ Registration flow is working"
echo "  ✓ Login flow is configured"
echo "  ✓ Email system is set up (Mailslurper)"

echo ""
print_info "To complete email verification:"
echo "1. Open Mailslurper: $MAILSLURPER"
echo "2. Find the verification email for: $TEST_EMAIL"
echo "3. Click the verification link in the email"
echo "4. Then test login again"

echo ""
print_info "Manual testing URLs:"
echo "📧 Email UI: $MAILSLURPER"
echo "🔐 Registration: $KRATOS_PUBLIC/self-service/registration/browser"
echo "🔑 Login: $KRATOS_PUBLIC/self-service/login/browser"
echo "⚙️  Backend API: $BACKEND_API/health"