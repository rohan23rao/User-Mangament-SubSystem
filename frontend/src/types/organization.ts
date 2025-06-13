// frontend/src/types/organization.ts
export interface Organization {
  id: string;
  parent_id?: string;
  org_type: 'organization' | 'tenant';
  name: string;
  description: string;
  owner_id?: string;
  is_default?: boolean;
  data: Record<string, any>;
  created_at: string;
  updated_at: string;
  
  // Hierarchy fields
  parent_name?: string;
  children?: Organization[];
  members?: Member[];
  member_count?: number;
}

export interface Member {
  user_id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: 'owner' | 'admin' | 'member';
  joined_at: string;
}

export interface OrgMember {
  org_id: string;
  org_name: string;
  org_type: 'organization' | 'tenant';
  role: 'owner' | 'admin' | 'member';
  joined_at: string;
  parent_name?: string;
}

export interface CreateOrgRequest {
  name: string;
  description: string;
  org_type: 'organization' | 'tenant';
  parent_id?: string; // Required for tenants
  data?: Record<string, any>;
}

export interface InviteUserRequest {
  email: string;
  role: 'member' | 'admin';
}

export interface UpdateMemberRoleRequest {
  role: 'member' | 'admin' | 'owner';
}

export type UserRole = 'member' | 'admin' | 'owner';
export type OrgType = 'organization' | 'tenant';