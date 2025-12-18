#!/bin/bash

echo "🧪 Testing Concurrent Features"
echo "================================"

echo -e "\n📧 Test 1: Bulk Notifications..."
curl -s -X POST http://localhost:8080/api/demo/bulk-notifications \
  -H "Content-Type: application/json" \
  -d '{
    "recipients": ["test1@example.com", "test2@example.com", "test3@example.com"],
    "subject": "Test",
    "message": "Concurrent test"
  }' | jq .

sleep 1

echo -e "\n\n📊 Test 2: Notification Stats..."
curl -s http://localhost:8080/api/demo/notification-stats | jq .

echo -e "\n\n🪑 Test 3: Parallel Table Availability Check..."
curl -s -X POST http://localhost:8080/api/demo/check-availability \
  -H "Content-Type: application/json" \
  -d '{
    "table_ids": ["550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440002"],
    "start_time": "2025-12-20T18:00:00Z",
    "end_time": "2025-12-20T20:00:00Z"
  }' | jq .

echo -e "\n\n📈 Test 4: Booking Stats..."
curl -s http://localhost:8080/api/demo/booking-stats/550e8400-e29b-41d4-a716-446655440000 | jq .

echo -e "\n\n🔍 Test 5: Parallel Search Across Restaurants..."
curl -s -X POST http://localhost:8080/api/demo/search-tables \
  -H "Content-Type: application/json" \
  -d '{
    "restaurant_ids": [
      "550e8400-e29b-41d4-a716-446655440010",
      "550e8400-e29b-41d4-a716-446655440011",
      "550e8400-e29b-41d4-a716-446655440012"
    ],
    "start_time": "2025-12-20T18:00:00Z",
    "end_time": "2025-12-20T20:00:00Z",
    "guest_count": 4
  }' | jq .

echo -e "\n\n✅ All tests completed!"
echo "👀 Check Docker logs: docker-compose logs -f api"
