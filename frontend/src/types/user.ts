export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  time_zone: string;
  ui_mode: string;
  can_create_organizations: boolean; // ADDED: New permission field
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
  org_type: string;
  role: string;
  joined_at: string;
}