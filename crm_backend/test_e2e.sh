#!/bin/bash

BASE_URL="http://localhost:8082/api/v1"

echo "=== E2E TESTING STARTING ==="

# 1. Login
echo "Logging in..."
TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@admin.com", "password":"Admin123!"}' | grep -oP '"token":"\K[^"]+')

if [ -z "$TOKEN" ]; then
    echo "Login failed!"
    exit 1
fi
echo "Login successful. Token acquired."

AUTH_HEADER="Authorization: Bearer $TOKEN"

# 2. Create Lead
echo "Creating lead..."
LEAD_RES=$(curl -s -X POST "$BASE_URL/leads" \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Jane", "last_name":"Doe", "email":"jane@test.com", "phone":"+123456789", "program_id":1}')
LEAD_ID=$(echo $LEAD_RES | grep -oP '"id":\K\d+')
echo "Lead created with ID: $LEAD_ID"

# 3. List Leads
echo "Listing leads..."
curl -s -X GET "$BASE_URL/leads" -H "$AUTH_HEADER" | head -n 1

# 4. Update Status (Normal -> Consultation)
echo "Updating status to 2 (Consultation)..."
curl -s -X PATCH "$BASE_URL/leads/$LEAD_ID/status" \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d '{"status_id":2}'

# 5. Update Status (Refusal validation - Expect 400)
echo "Testing refusal validation (expecting 400)..."
curl -s -I -X PATCH "$BASE_URL/leads/$LEAD_ID/status" \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d '{"status_id":7}' | head -n 1

# 6. Update Status (Refusal with reason)
echo "Updating status to 7 with reason..."
curl -s -X PATCH "$BASE_URL/leads/$LEAD_ID/status" \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d '{"status_id":7, "refusal_reason":"Too expensive"}'

# 7. Add Interaction
echo "Adding interaction..."
curl -s -X POST "$BASE_URL/interactions" \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d "{\"lead_id\":$LEAD_ID, \"type\":\"call\", \"content\":\"Follow up call\"}"

# 8. Get History
echo "Viewing history..."
curl -s -X GET "$BASE_URL/interactions/lead/$LEAD_ID" -H "$AUTH_HEADER"

# 9. Analytics Dashboard
echo "Fetching dashboard metrics..."
curl -s -X GET "$BASE_URL/analytics/dashboard" -H "$AUTH_HEADER"

# 10. Merge Leads
echo "Creating another lead to merge..."
LEAD2_RES=$(curl -s -X POST "$BASE_URL/leads" \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Jane DUPE", "email":"jane@test.com"}')
LEAD2_ID=$(echo $LEAD2_RES | grep -oP '"id":\K\d+')
echo "Duplicate lead created with ID: $LEAD2_ID"

echo "Merging lead $LEAD2_ID into $LEAD_ID..."
curl -s -X POST "$BASE_URL/leads/merge" \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d "{\"target_lead_id\":$LEAD_ID, \"source_lead_id\":$LEAD2_ID}"

echo "=== E2E TESTING COMPLETED ==="
