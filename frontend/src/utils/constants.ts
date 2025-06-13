// frontend/src/utils/constants.ts
export const API_ENDPOINTS = {
  KRATOS_PUBLIC: process.env.REACT_APP_KRATOS_PUBLIC_URL || 'http://localhost:4433',
  API_BASE: process.env.REACT_APP_API_URL || 'http://localhost:3000',
};

export const ROUTES = {
  LOGIN: '/login',
  REGISTER: '/register',
  DASHBOARD: '/dashboard',
  PROFILE: '/profile',
  USERS: '/users',
  ORGANIZATIONS: '/organizations',
  SETTINGS: '/settings',
  ANALYTICS: '/analytics',
};

export const USER_ROLES = {
  MEMBER: 'member',
  ADMIN: 'admin',
  OWNER: 'owner',
} as const;

export const ORG_TYPES = {
  ORGANIZATION: 'organization',
  TENANT: 'tenant',
} as const;

export const ORG_TYPE_LABELS = {
  [ORG_TYPES.ORGANIZATION]: 'Organization',
  [ORG_TYPES.TENANT]: 'Tenant',
} as const;

export const ROLE_LABELS = {
  [USER_ROLES.OWNER]: 'Owner',
  [USER_ROLES.ADMIN]: 'Admin',
  [USER_ROLES.MEMBER]: 'Member',
} as const;

export const ROLE_DESCRIPTIONS = {
  [USER_ROLES.OWNER]: 'Full control over the organization',
  [USER_ROLES.ADMIN]: 'Can manage members and settings',
  [USER_ROLES.MEMBER]: 'Can view and participate',
} as const;

export const ORG_TYPE_DESCRIPTIONS = {
  [ORG_TYPES.ORGANIZATION]: 'Top-level workspace for your company or team',
  [ORG_TYPES.TENANT]: 'Project or environment within an organization',
} as const;