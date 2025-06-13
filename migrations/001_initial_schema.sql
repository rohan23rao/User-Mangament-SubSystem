-- Create UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create OrgType enum
CREATE TYPE OrgType AS ENUM(
    'domain',
    'organization', 
    'tenant'
);

-- Create organizations table first (since users references it)
CREATE TABLE IF NOT EXISTS organizations(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    domain_id uuid NULL,
    org_id uuid NULL,
    org_type OrgType NOT NULL,
    name varchar(1024) NOT NULL UNIQUE,
    description text,
    owner_id uuid NULL, -- Will be set after users table exists
    is_default boolean NOT NULL DEFAULT false, -- ADDED: Flag for default organization
    data jsonb DEFAULT '{}',
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz DEFAULT CURRENT_TIMESTAMP
);

-- Create users table
CREATE TABLE IF NOT EXISTS users(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id uuid REFERENCES organizations(id) ON DELETE SET NULL,
    email varchar(1024) NOT NULL UNIQUE,
    first_name varchar(1024) NOT NULL DEFAULT '',
    last_name varchar(1024) NOT NULL DEFAULT '',
    time_zone varchar(255) NOT NULL DEFAULT 'UTC',
    ui_mode varchar(255) NOT NULL DEFAULT 'system',
    can_create_organizations boolean NOT NULL DEFAULT false, -- ADDED: Permission to create orgs
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    last_login timestamptz NULL
);

-- Create user_organization_links table for many-to-many relationships
CREATE TABLE IF NOT EXISTS user_organization_links(
    user_id uuid NOT NULL,
    organization_id uuid NOT NULL,
    role varchar(50) NOT NULL DEFAULT 'member',
    joined_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, organization_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- Add foreign key constraint for organization owner after users table exists
ALTER TABLE organizations 
ADD CONSTRAINT fk_organizations_owner 
FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE SET NULL;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_org_id ON users(org_id);
CREATE INDEX IF NOT EXISTS idx_organizations_name ON organizations(name);
CREATE INDEX IF NOT EXISTS idx_organizations_type ON organizations(org_type);
CREATE INDEX IF NOT EXISTS idx_organizations_default ON organizations(is_default); -- ADDED: Index for default org
CREATE INDEX IF NOT EXISTS idx_users_can_create_orgs ON users(can_create_organizations); -- ADDED: Index for org creation permission
CREATE INDEX IF NOT EXISTS idx_user_org_links_user_id ON user_organization_links(user_id);
CREATE INDEX IF NOT EXISTS idx_user_org_links_org_id ON user_organization_links(organization_id);
CREATE INDEX IF NOT EXISTS idx_user_org_links_role ON user_organization_links(role);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_organizations_updated_at 
    BEFORE UPDATE ON organizations 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ADDED: Function to handle first user and default organization setup
CREATE OR REPLACE FUNCTION handle_first_user_setup()
RETURNS TRIGGER AS $$
DECLARE
    user_count integer;
    default_org_id uuid;
BEGIN
    -- Count existing users
    SELECT COUNT(*) INTO user_count FROM users;
    
    -- If this is the first user
    IF user_count = 1 THEN
        -- Give them permission to create organizations
        UPDATE users SET can_create_organizations = true WHERE id = NEW.id;
        
        -- Create default organization if it doesn't exist
        SELECT id INTO default_org_id FROM organizations WHERE is_default = true LIMIT 1;
        IF default_org_id IS NULL THEN
            INSERT INTO organizations (name, description, org_type, is_default, owner_id)
            VALUES ('Default Organization', 'Default organization for new users', 'organization', true, NEW.id)
            RETURNING id INTO default_org_id;
        ELSE
            -- Update existing default org owner
            UPDATE organizations SET owner_id = NEW.id WHERE id = default_org_id;
        END IF;
        
        -- Add first user as owner of default organization
        INSERT INTO user_organization_links (user_id, organization_id, role)
        VALUES (NEW.id, default_org_id, 'owner')
        ON CONFLICT (user_id, organization_id) DO UPDATE SET role = 'owner';
    ELSE
        -- For all other users, add them to default organization as members
        SELECT id INTO default_org_id FROM organizations WHERE is_default = true LIMIT 1;
        IF default_org_id IS NOT NULL THEN
            INSERT INTO user_organization_links (user_id, organization_id, role)
            VALUES (NEW.id, default_org_id, 'member')
            ON CONFLICT (user_id, organization_id) DO NOTHING;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- ADDED: Trigger to handle first user setup
CREATE TRIGGER handle_first_user_trigger
    AFTER INSERT ON users
    FOR EACH ROW EXECUTE FUNCTION handle_first_user_setup();