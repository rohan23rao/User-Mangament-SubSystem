// Base organization interface matching your API
export interface Organization {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
  owner_id?: string;
  org_type?: string;
  members?: Member[];
}

// Extended organization interface if you need additional frontend-specific fields
export interface OrganizationWithMetadata extends Organization {
  user_count?: number;
  admin_count?: number;
  is_current_user_admin?: boolean;
}

// Member interface - matches backend Member struct
export interface Member {
  user_id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: 'admin' | 'member';
  joined_at: string;
}

// For organization creation
export interface CreateOrganizationRequest {
  name: string;
  description: string;
  org_type: string;
  domain_id?: string;
  org_id?: string;
  data?: {[key: string]: any};
}

// For organization updates
export interface UpdateOrganizationRequest {
  name?: string;
  description?: string;
  org_type?: string;
  domain_id?: string;
  org_id?: string;
  data?: {[key: string]: any};
}

// For creating organization requests (alias for consistency)
export interface CreateOrgRequest extends CreateOrganizationRequest {}

// For inviting users to organizations
export interface InviteUserRequest {
  email: string;
  role: 'admin' | 'member';
}

// For updating member roles
export interface UpdateMemberRoleRequest {
  role: 'admin' | 'member';
}