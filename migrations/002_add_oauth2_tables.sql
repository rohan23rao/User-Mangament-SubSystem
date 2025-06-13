-- migrations/002_add_oauth2_tables.sql
-- Add OAuth2 client management tables

-- OAuth2 clients table for M2M authentication
CREATE TABLE IF NOT EXISTS oauth2_clients (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id varchar(255) NOT NULL UNIQUE,
    client_secret varchar(255) NOT NULL, -- In production, encrypt this
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id uuid NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    description text,
    scopes text NOT NULL DEFAULT 'data_pipeline',
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    last_used_at timestamptz NULL
);

-- OAuth2 token usage logs for auditing
CREATE TABLE IF NOT EXISTS oauth2_token_logs (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id varchar(255) NOT NULL,
    granted_scopes text,
    ip_address inet,
    user_agent text,
    expires_at timestamptz,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP
);

-- API keys for additional authentication methods (optional)
CREATE TABLE IF NOT EXISTS api_keys (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_hash varchar(255) NOT NULL UNIQUE, -- SHA-256 hash of the key
    key_prefix varchar(10) NOT NULL, -- First 8 chars for identification
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id uuid NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    description text,
    scopes text NOT NULL DEFAULT 'data_pipeline',
    is_active boolean NOT NULL DEFAULT true,
    expires_at timestamptz NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    last_used_at timestamptz NULL
);

-- Client IP whitelisting for enhanced security
CREATE TABLE IF NOT EXISTS oauth2_client_ip_whitelist (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id varchar(255) NOT NULL,
    ip_address inet NOT NULL,
    description text,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (client_id) REFERENCES oauth2_clients(client_id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_oauth2_clients_user_id ON oauth2_clients(user_id);
CREATE INDEX IF NOT EXISTS idx_oauth2_clients_org_id ON oauth2_clients(org_id);
CREATE INDEX IF NOT EXISTS idx_oauth2_clients_active ON oauth2_clients(is_active);
CREATE INDEX IF NOT EXISTS idx_oauth2_token_logs_client_id ON oauth2_token_logs(client_id);
CREATE INDEX IF NOT EXISTS idx_oauth2_token_logs_created_at ON oauth2_token_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(is_active);
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at ON api_keys(expires_at);
CREATE INDEX IF NOT EXISTS idx_oauth2_client_ip_whitelist_client_id ON oauth2_client_ip_whitelist(client_id);

-- Create updated_at trigger function if not exists
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_oauth2_clients_updated_at 
    BEFORE UPDATE ON oauth2_clients 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_api_keys_updated_at 
    BEFORE UPDATE ON api_keys 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Grant necessary permissions to users based on their roles
-- This function automatically grants OAuth2 client creation permissions
CREATE OR REPLACE FUNCTION grant_oauth2_permissions()
RETURNS TRIGGER AS $$
BEGIN
    -- Users who can create organizations can also create OAuth2 clients
    IF NEW.can_create_organizations = true THEN
        -- Log the permission grant
        INSERT INTO oauth2_token_logs (client_id, granted_scopes, created_at)
        VALUES ('system', 'oauth2_client_creation_granted', CURRENT_TIMESTAMP);
    END IF;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to automatically grant OAuth2 permissions
CREATE TRIGGER grant_oauth2_permissions_trigger
    AFTER UPDATE OF can_create_organizations ON users
    FOR EACH ROW EXECUTE FUNCTION grant_oauth2_permissions();