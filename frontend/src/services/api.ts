// frontend/src/services/api.ts
import axios from 'axios';
import { User } from '../types/user';
import { Organization, CreateOrgRequest, InviteUserRequest, UpdateMemberRoleRequest, Member } from '../types/organization';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:3000';

axios.defaults.withCredentials = true;
axios.defaults.baseURL = API_URL;

export class ApiService {
  // User endpoints
  static async getCurrentUser(): Promise<User> {
    const response = await axios.get('/api/whoami');
    return response.data;
  }

  static async getUsers(): Promise<User[]> {
    const response = await axios.get('/api/users');
    return response.data;
  }

  static async getUser(id: string): Promise<User> {
    const response = await axios.get(`/api/users/${id}`);
    return response.data;
  }

  static async updateUserProfile(data: Partial<User>): Promise<User> {
    const response = await axios.put('/api/users/profile', data);
    return response.data;
  }

  // Organization endpoints
  static async getOrganizations(): Promise<Organization[]> {
    const response = await axios.get('/api/organizations');
    return response.data;
  }

  static async getOrganization(id: string): Promise<Organization> {
    const response = await axios.get(`/api/organizations/${id}`);
    return response.data;
  }

  static async getOrganizationWithTenants(id: string): Promise<Organization> {
    const response = await axios.get(`/api/organizations/${id}/tenants`);
    return response.data;
  }

  static async createOrganization(data: CreateOrgRequest): Promise<Organization> {
    const response = await axios.post('/api/organizations', data);
    return response.data;
  }

  static async updateOrganization(id: string, data: Partial<CreateOrgRequest>): Promise<Organization> {
    const response = await axios.put(`/api/organizations/${id}`, data);
    return response.data;
  }

  static async deleteOrganization(id: string): Promise<void> {
    await axios.delete(`/api/organizations/${id}`);
  }

  // Member management endpoints
  static async getOrganizationMembers(organizationId: string): Promise<Member[]> {
    const response = await axios.get(`/api/organizations/${organizationId}/members`);
    return response.data;
  }

  static async addOrganizationMember(organizationId: string, data: InviteUserRequest): Promise<void> {
    await axios.post(`/api/organizations/${organizationId}/members`, data);
  }

  static async removeOrganizationMember(organizationId: string, userId: string): Promise<void> {
    await axios.delete(`/api/organizations/${organizationId}/members/${userId}`);
  }

  static async updateMemberRole(organizationId: string, userId: string, data: UpdateMemberRoleRequest): Promise<Member> {
    const response = await axios.put(`/api/organizations/${organizationId}/members/${userId}`, data);
    return response.data;
  }

  // Utility endpoints
  static async healthCheck(): Promise<any> {
    const response = await axios.get('/health');
    return response.data;
  }

  static async debugAuth(): Promise<any> {
    const response = await axios.get('/debug/auth');
    return response.data;
  }
}

// Response interceptor for error handling
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Only redirect to login if we're not already on auth pages
      if (!window.location.pathname.includes('/login') && 
          !window.location.pathname.includes('/register') &&
          !window.location.pathname.includes('/verification')) {
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

// Request interceptor for debugging
axios.interceptors.request.use(
  (config) => {
    console.log(`API Request: ${config.method?.toUpperCase()} ${config.url}`);
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);