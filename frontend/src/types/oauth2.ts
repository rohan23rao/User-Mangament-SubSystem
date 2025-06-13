// frontend/src/types/oauth2.ts

export interface OAuth2Client {
  id: string;
  client_id: string;
  client_secret?: string; // Only shown once after creation
  user_id: string;
  org_id: string;
  name: string;
  description: string;
  scopes: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  last_used_at?: string;
}

export interface CreateM2MClientRequest {
  name: string;
  description: string;
  org_id: string;
  scopes?: string;
}

export interface TokenRequest {
  client_id: string;
  client_secret: string;
  grant_type?: string;
  scope?: string;
}

export interface TokenResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  scope: string;
  refresh_token?: string;
}

export interface TokenInfo {
  active: boolean;
  client_id: string;
  scope: string;
  expires_at: string;
  issued_at?: string;
  subject?: string;
}

export interface OAuth2ClientsResponse {
  clients: OAuth2Client[];
  count: number;
}

// Predefined scopes for different use cases
export const OAUTH2_SCOPES = {
  DATA_PIPELINE: 'data_pipeline',
  DATA_EXPORT: 'data_export',
  TELEMETRY_INGEST: 'telemetry_ingest',
  READ_ONLY: 'read',
  FULL_ACCESS: 'data_pipeline data_export telemetry_ingest',
} as const;

export const SCOPE_DESCRIPTIONS = {
  [OAUTH2_SCOPES.DATA_PIPELINE]: 'Access to data pipeline operations',
  [OAUTH2_SCOPES.DATA_EXPORT]: 'Export data and generate reports',
  [OAUTH2_SCOPES.TELEMETRY_INGEST]: 'Send telemetry and metrics data',
  [OAUTH2_SCOPES.READ_ONLY]: 'Read-only access to resources',
  [OAUTH2_SCOPES.FULL_ACCESS]: 'Full access to all operations',
} as const;