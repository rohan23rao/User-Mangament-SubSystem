export interface KratosSession {
  id: string;
  active: boolean;
  expires_at: string;
  authenticated_at: string;
  authenticator_assurance_level: string;
  authentication_methods: Array<{
    method: string;
    aal: string;
    completed_at: string;
    provider?: string;
  }>;
  issued_at: string;
  identity: KratosIdentity;
}

export interface KratosIdentity {
  id: string;
  schema_id: string;
  schema_url: string;
  state: string;
  state_changed_at: string;
  traits: {
    email: string;
    name?: {
      first: string;
      last: string;
    };
  };
  verifiable_addresses: Array<{
    id: string;
    value: string;
    verified: boolean;
    via: string;
    status: string;
    created_at: string;
    updated_at: string;
  }>;
  recovery_addresses: Array<{
    id: string;
    value: string;
    via: string;
    created_at: string;
    updated_at: string;
  }>;
  metadata_public?: any;
  created_at: string;
  updated_at: string;
}

export interface LoginFlow {
  id: string;
  type: string;
  expires_at: string;
  issued_at: string;
  request_url: string;
  ui: {
    action: string;
    method: string;
    nodes: Array<{
      type: string;
      group: string;
      attributes: any;
      messages: any[];
      meta: any;
    }>;
  };
  created_at: string;
  updated_at: string;
  refresh?: boolean;
  requested_aal?: string;
}

export interface RegistrationFlow {
  id: string;
  type: string;
  expires_at: string;
  issued_at: string;
  request_url: string;
  ui: {
    action: string;
    method: string;
    nodes: Array<{
      type: string;
      group: string;
      attributes: any;
      messages: any[];
      meta: any;
    }>;
  };
  created_at: string;
  updated_at: string;
}
