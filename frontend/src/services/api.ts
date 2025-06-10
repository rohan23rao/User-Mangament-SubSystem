import axios from 'axios';
import { notifications } from '@mantine/notifications';
import { User } from '../types/user';
import { Organization, CreateOrgRequest, InviteUserRequest, UpdateMemberRoleRequest, Member } from '../types/organization';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:3000';

// Create axios instance with proper configuration
const api = axios.create({
  baseURL: API_URL,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

export class ApiService {
  // User endpoints
  static async getCurrentUser(): Promise<User> {
    const response = await api.get('/api/whoami');
    return response.data;
  }

  static async getUsers(): Promise<User[]> {
    const response = await api.get('/api/users');
    return response.data;
  }

  static async getUser(id: string): Promise<User> {
    const response = await api.get(`/api/users/${id}`);
    return response.data;
  }

  // Organization endpoints
  static async getOrganizations(): Promise<Organization[]> {
    const response = await api.get('/api/organizations');
    return response.data;
  }

  static async getOrganization(id: string): Promise<Organization> {
    const response = await api.get(`/api/organizations/${id}`);
    return response.data;
  }

  static async createOrganization(data: CreateOrgRequest): Promise<Organization> {
    const response = await api.post('/api/organizations', data);
    return response.data;
  }

  static async updateOrganization(id: string, data: Partial<CreateOrgRequest>): Promise<Organization> {
    const response = await api.put(`/api/organizations/${id}`, data);
    return response.data;
  }

  static async deleteOrganization(id: string): Promise<void> {
    await api.delete(`/api/organizations/${id}`);
  }

  // Organization member endpoints
  static async getOrganizationMembers(organizationId: string): Promise<Member[]> {
    const response = await api.get(`/api/organizations/${organizationId}/members`);
    return response.data;
  }

  static async addOrganizationMember(organizationId: string, data: InviteUserRequest): Promise<void> {
    await api.post(`/api/organizations/${organizationId}/members`, data);
  }

  static async removeOrganizationMember(organizationId: string, userId: string): Promise<void> {
    await api.delete(`/api/organizations/${organizationId}/members/${userId}`);
  }

  static async updateMemberRole(organizationId: string, userId: string, data: UpdateMemberRoleRequest): Promise<Member> {
    const response = await api.put(`/api/organizations/${organizationId}/members/${userId}/role`, data);
    return response.data;
  }

  // Health check
  static async healthCheck(): Promise<any> {
    const response = await api.get('/health');
    return response.data;
  }
}

// Error handling interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Handle verification requirement
    if (error.response?.status === 403 && error.response?.data?.code === 'EMAIL_NOT_VERIFIED') {
      console.log('Email verification required - redirecting to verification page');
      window.location.href = '/verification';
      return Promise.reject(error);
    }
    
    // Handle admin permission requirement
    if (error.response?.status === 403 && error.response?.data?.code === 'ADMIN_REQUIRED') {
      console.log('Admin permission required - showing error notification');
      notifications.show({
        title: 'Permission Denied',
        message: 'Only administrators can create organizations',
        color: 'red',
      });
      return Promise.reject(error);
    }
    
    // Don't auto-redirect on 401 - let the auth context handle it
    if (error.response?.status === 401) {
      console.log('API returned 401 - authentication required');
    }
    console.error('API Error:', error.response?.data || error.message);
    return Promise.reject(error);
  }
);
