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
};

export const USER_ROLES = {
  MEMBER: 'member',
  ADMIN: 'admin',
  OWNER: 'owner',
} as const;

export const ORG_TYPES = {
  DOMAIN: 'domain',
  ORGANIZATION: 'organization',
  TENANT: 'tenant',
} as const;
