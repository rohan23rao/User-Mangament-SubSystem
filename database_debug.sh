#!/bin/bash

echo "🔧 Database Debug and Fix Commands"
echo "================================="

# Check current directory permissions
echo "1. Checking directory permissions:"
ls -la . | grep -E "(app\.db|\.)"

# Check if database file exists and its permissions
echo "2. Checking database file:"
if [ -f app.db ]; then
    echo "Database file exists:"
    ls -la app.db
    echo "File size: $(stat -c%s app.db 2>/dev/null || stat -f%z app.db 2>/dev/null) bytes"
else
    echo "Database file does not exist"
fi

# Test write permissions
echo "3. Testing write permissions:"
if touch test_write_permission 2>/dev/null; then
    echo "✅ Write permissions OK"
    rm test_write_permission
else
    echo "❌ No write permissions in current directory"
    echo "Current directory: $(pwd)"
    echo "User: $(whoami)"
fi

# Check SQLite installation
echo "4. Checking SQLite:"
if command -v sqlite3 &> /dev/null; then
    echo "✅ SQLite3 is installed: $(sqlite3 --version)"
else
    echo "❌ SQLite3 not found"
fi

# Manual database creation test
echo "5. Manual database creation test:"
sqlite3 test.db "CREATE TABLE test (id INTEGER); INSERT INTO test VALUES (1); SELECT * FROM test; DROP TABLE test;" && rm test.db
if [ $? -eq 0 ]; then
    echo "✅ SQLite manual test passed"
else
    echo "❌ SQLite manual test failed"
fi

# Check Go SQLite driver
echo "6. Testing Go SQLite driver:"
cat > test_db.go << 'EOF'
package main

import (
    "database/sql"
    "fmt"
    "os"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    // Remove any existing test db
    os.Remove("test_go.db")
    
    db, err := sql.Open("sqlite3", "./test_go.db")
    if err != nil {
        fmt.Printf("Error opening database: %v\n", err)
        return
    }
    defer db.Close()

    _, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
    if err != nil {
        fmt.Printf("Error creating table: %v\n", err)
        return
    }

    _, err = db.Exec("INSERT INTO test (name) VALUES (?)", "test_value")
    if err != nil {
        fmt.Printf("Error inserting data: %v\n", err)
        return
    }

    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
    if err != nil {
        fmt.Printf("Error querying data: %v\n", err)
        return
    }

    fmt.Printf("✅ Go SQLite test passed. Records: %d\n", count)
    
    // Clean up
    os.Remove("test_go.db")
}
EOF

go run test_db.go
rm test_db.go

echo ""
echo "7. Force clean database recreation:"
echo "rm -f app.db"
echo "go run main.go"
echo ""

echo "8. Alternative: Create database manually:"
echo "sqlite3 app.db < create_tables.sql"

# Create SQL file for manual database creation
cat > create_tables.sql << 'EOF'
CREATE TABLE IF NOT EXISTS organizations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    owner_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS organization_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    org_id INTEGER,
    user_id TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'member',
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations (id),
    UNIQUE(org_id, user_id)
);

CREATE TABLE IF NOT EXISTS user_profiles (
    user_id TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    first_name TEXT,
    last_name TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_org_members_user_id ON organization_members(user_id);
CREATE INDEX IF NOT EXISTS idx_org_members_org_id ON organization_members(org_id);
CREATE INDEX IF NOT EXISTS idx_organizations_owner_id ON organizations(owner_id);
EOF

echo "SQL file created: create_tables.sql"