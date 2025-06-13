// frontend/src/types/user.ts
export interface User {
  id: string;
  email: string;
  email_verified: boolean;
  first_name: string;
  last_name: string;
  time_zone: string;
  ui_mode: string;
  can_create_organizations: boolean;
  traits: {
    email: string;
    name?: {
      first: string;
      last: string;
    };
  };
  verifiable_addresses?: Array<{
    id: string;
    value: string;
    verified: boolean;
    via: string;
    status: string;
    created_at: string;
    updated_at: string;
  }>;
  organizations?: OrgMember[];
  created_at: string;
  updated_at: string;
  last_login?: string;
}

export interface OrgMember {
  org_id: string;
  org_name: string;
  org_type: 'organization' | 'tenant';
  role: 'owner' | 'admin' | 'member';
  joined_at: string;
  parent_name?: string;
}