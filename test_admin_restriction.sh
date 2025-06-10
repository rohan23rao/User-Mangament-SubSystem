#!/bin/bash

echo "🔒 Testing Admin Restrictions"
echo "============================"

# Test 1: Check who are the current admins
echo ""
echo "📊 Current Admin Users:"
docker-compose exec postgres psql -U userms -d userms -c "
SELECT u.email, uo.role, o.name as org_name 
FROM users u 
JOIN user_organization_links uo ON u.id = uo.user_id 
JOIN organizations o ON o.id = uo.organization_id 
WHERE uo.role = 'admin' 
ORDER BY u.email;"

# Test 2: Check member users 
echo ""
echo "👥 Current Member Users:"
docker-compose exec postgres psql -U userms -d userms -c "
SELECT u.email, uo.role, o.name as org_name 
FROM users u 
JOIN user_organization_links uo ON u.id = uo.user_id 
JOIN organizations o ON o.id = uo.organization_id 
WHERE uo.role = 'member' 
ORDER BY u.email;"

# Test 3: Show recent organization creation logs
echo ""
echo "📋 Recent Organization Creation Attempts:"
docker-compose logs backend | grep -E "(Admin check|organization creation|not authorized)" | tail -10

echo ""
echo "✅ Test completed! Check if restrictions are working properly."
echo "🔍 Look for 'Admin check' logs to see if users are being properly validated." 