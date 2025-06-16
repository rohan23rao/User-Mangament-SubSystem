#!/bin/bash
# This script updates your handler files to replace mux.Vars with r.PathValue

echo "Updating handlers to use net/http path parameters..."

# Update internal/handlers/organizations.go
sed -i 's/github.com\/gorilla\/mux//g' internal/handlers/organizations.go
sed -i 's/vars := mux\.Vars(r)/\/\/ Path parameters extracted with r.PathValue/g' internal/handlers/organizations.go
sed -i 's/orgID := vars\["id"\]/orgID := r.PathValue("id")/g' internal/handlers/organizations.go
sed -i 's/userID := vars\["user_id"\]/userID := r.PathValue("user_id")/g' internal/handlers/organizations.go
sed -i 's/clientId := vars\["clientId"\]/clientId := r.PathValue("clientId")/g' internal/handlers/organizations.go

# Update internal/handlers/verification.go
sed -i 's/github.com\/gorilla\/mux//g' internal/handlers/verification.go
sed -i 's/vars := mux\.Vars(r)/\/\/ Path parameters extracted with r.PathValue/g' internal/handlers/verification.go
sed -i 's/userID := vars\["id"\]/userID := r.PathValue("id")/g' internal/handlers/verification.go

# Update internal/handlers/oauth2.go
sed -i 's/github.com\/gorilla\/mux//g' internal/handlers/oauth2.go
sed -i 's/vars := mux\.Vars(r)/\/\/ Path parameters extracted with r.PathValue/g' internal/handlers/oauth2.go
sed -i 's/clientID := vars\["clientId"\]/clientID := r.PathValue("clientId")/g' internal/handlers/oauth2.go

# Update internal/handlers/users.go (if it has mux.Vars)
sed -i 's/github.com\/gorilla\/mux//g' internal/handlers/users.go
sed -i 's/vars := mux\.Vars(r)/\/\/ Path parameters extracted with r.PathValue/g' internal/handlers/users.go
sed -i 's/userID := vars\["id"\]/userID := r.PathValue("id")/g' internal/handlers/users.go

echo "Handler files updated successfully!"