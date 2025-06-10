export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  time_zone: string;
  ui_mode: string;
  traits: {
    email: string;
    name?: {
      first: string;
      last: string;
    };
  };
  organizations?: OrgMember[];
  created_at: string;
  updated_at: string;
  last_login?: string;
  verified: boolean;
  recovery_addresses?: RecoveryAddress[];
  verifiable_addresses?: VerifiableAddress[];
}

export interface RecoveryAddress {
  id: string;
  value: string;
  via: string;
}

export interface VerifiableAddress {
  id: string;
  value: string;
  verified: boolean;
  via: string;
  status: string;
}

export interface OrgMember {
  org_id: string;
  org_name: string;
  org_type: string;
  role: string;
  joined_at: string;
}
