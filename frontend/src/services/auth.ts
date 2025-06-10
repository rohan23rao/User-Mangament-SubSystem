import axios, { AxiosInstance } from 'axios';
import { KratosSession, LoginFlow, RegistrationFlow } from '../types/auth';
import { ApiService } from './api';

const KRATOS_PUBLIC_URL = process.env.REACT_APP_KRATOS_PUBLIC_URL || 'http://localhost:4433';

// Create axios instance with proper cookie handling
const kratosClient: AxiosInstance = axios.create({
  baseURL: KRATOS_PUBLIC_URL,
  withCredentials: true, // Essential for cookies
  headers: {
    'Content-Type': 'application/json',
  },
});



export class AuthService {
  static async getSession(): Promise<KratosSession> {
    try {
      const response = await kratosClient.get('/sessions/whoami');
      return response.data;
    } catch (error) {
      console.error('Session error:', error);
      throw new Error('Not authenticated');
    }
  }

  static async getLoginFlow(flowId?: string): Promise<LoginFlow> {
    const url = flowId 
      ? `/self-service/login/flows?id=${flowId}`
      : '/self-service/login/browser';
    
    try {
      const response = await kratosClient.get(url);
      return response.data;
    } catch (error: any) {
      console.error('Login flow error:', error);
      
      // If we get a 400 error, it might be because of invalid session cookies
      // Clear them and try again
      if (error.response?.status === 400) {
        console.log('Clearing invalid session cookies and retrying...');
        this.clearSessionCookies();
        
        try {
          const retryResponse = await kratosClient.get(url);
          return retryResponse.data;
        } catch (retryError) {
          console.error('Retry after clearing cookies also failed:', retryError);
          throw retryError;
        }
      }
      
      throw error;
    }
  }

  static async getRegistrationFlow(flowId?: string): Promise<RegistrationFlow> {
    const url = flowId 
      ? `/self-service/registration/flows?id=${flowId}`
      : '/self-service/registration/browser';
    
    try {
      const response = await kratosClient.get(url);
      return response.data;
    } catch (error: any) {
      console.error('Registration flow error:', error);
      
      // If we get a 400 error, it might be because of invalid session cookies
      // Clear them and try again
      if (error.response?.status === 400) {
        console.log('Clearing invalid session cookies and retrying registration flow...');
        this.clearSessionCookies();
        
        try {
          const retryResponse = await kratosClient.get(url);
          return retryResponse.data;
        } catch (retryError) {
          console.error('Retry after clearing cookies also failed:', retryError);
          throw retryError;
        }
      }
      
      throw error;
    }
  }

  static async submitLoginFlow(flowId: string, body: any): Promise<any> {
    try {
      const response = await kratosClient.post(`/self-service/login?flow=${flowId}`, body);
      return response.data;
    } catch (error) {
      console.error('Login submission error:', error);
      throw error;
    }
  }

  static async submitRegistrationFlow(flowId: string, body: any): Promise<any> {
    try {
      const response = await kratosClient.post(`/self-service/registration?flow=${flowId}`, body);
      return response.data;
    } catch (error) {
      console.error('Registration submission error:', error);
      throw error;
    }
  }

  static async logout(): Promise<void> {
    try {
      // Clear session cookies first
      this.clearSessionCookies();
      
      // Try to get logout flow and execute logout
      try {
        const logoutResponse = await kratosClient.get('/self-service/logout/browser');
        const logoutUrl = logoutResponse.data.logout_url;
        await kratosClient.get(logoutUrl);
      } catch (logoutError) {
        console.warn('Kratos logout failed, but continuing with local cleanup:', logoutError);
      }
      
      // Don't do a hard redirect, let React handle navigation
      console.log('Logout completed successfully');
    } catch (error) {
      console.error('Logout error:', error);
      // Clear cookies anyway
      this.clearSessionCookies();
      throw error;
    }
  }

  static async initiateGoogleLogin(): Promise<string> {
    try {
      // Get login flow first
      const flow = await this.getLoginFlow();
      
      // Find Google OAuth node
      const googleNode = flow.ui.nodes.find(
        node => node.attributes?.name === 'provider' && node.attributes?.value === 'google'
      );
      
      if (!googleNode) {
        throw new Error('Google OAuth not configured');
      }
      
      // Return the action URL for Google OAuth
      return `${flow.ui.action}?flow=${flow.id}`;
    } catch (error) {
      console.error('Google login initiation error:', error);
      throw error;
    }
  }

  static redirectToKratosLogin(): void {
    window.location.href = `${KRATOS_PUBLIC_URL}/self-service/login/browser`;
  }

  static redirectToKratosRegistration(): void {
    window.location.href = `${KRATOS_PUBLIC_URL}/self-service/registration/browser`;
  }

  static clearSessionCookies(): void {
    // Clear Kratos session cookies
    const cookiesToClear = [
      'ory_kratos_session',
      'ory_kratos_continuity',
      'csrf_token_82b119fa58a0a1cb6faa9738c1d0dbbf04fcc89a657b7beb31fcde400ced48ab'
    ];
    
    cookiesToClear.forEach(cookieName => {
      // Clear for current domain
      document.cookie = `${cookieName}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;`;
      // Also try to clear for localhost domain
      document.cookie = `${cookieName}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/; domain=localhost;`;
    });
    
    console.log('Cleared Kratos session cookies');
  }

  // Add missing methods that are called in useAuth.tsx
  static async getCurrentUser() {
    return ApiService.getCurrentUser();
  }

  static async createLoginFlow(): Promise<LoginFlow> {
    return this.getLoginFlow();
  }

  static async createRegistrationFlow(): Promise<RegistrationFlow> {
    return this.getRegistrationFlow();
  }

  static async submitLogin(flowId: string, email: string, password: string): Promise<any> {
    try {
      // First get the login flow to extract CSRF token and other required fields
      const flow = await this.getLoginFlow(flowId);
      
      // Find the CSRF token from the flow nodes
      const csrfTokenNode = flow.ui.nodes.find(
        node => node.attributes?.name === 'csrf_token'
      );
      
      const body: any = {
        method: 'password',
        identifier: email,
        password: password,
      };
      
      // Include CSRF token if found
      if (csrfTokenNode?.attributes?.value) {
        body.csrf_token = csrfTokenNode.attributes.value;
        console.log('Including CSRF token in login submission');
      } else {
        console.warn('No CSRF token found in login flow');
      }
      
      return this.submitLoginFlow(flowId, body);
    } catch (error) {
      console.error('Submit login error:', error);
      throw error;
    }
  }

  static async submitRegistration(flowId: string, email: string, password: string, firstName?: string, lastName?: string): Promise<any> {
    try {
      // First get the registration flow to extract CSRF token and other required fields
      const flow = await this.getRegistrationFlow(flowId);
      
      // Find the CSRF token from the flow nodes
      const csrfTokenNode = flow.ui.nodes.find(
        node => node.attributes?.name === 'csrf_token'
      );
      
      const body: any = {
        method: 'password',
        traits: {
          email: email,
          name: {
            first: firstName || '',
            last: lastName || '',
          },
        },
        password: password,
      };
      
      // Include CSRF token if found
      if (csrfTokenNode?.attributes?.value) {
        body.csrf_token = csrfTokenNode.attributes.value;
        console.log('Including CSRF token in registration submission');
      } else {
        console.warn('No CSRF token found in registration flow');
      }
      
      return this.submitRegistrationFlow(flowId, body);
    } catch (error) {
      console.error('Submit registration error:', error);
      throw error;
    }
  }

  static async getGoogleAuthUrl(flowId: string, flowType: 'login' | 'registration'): Promise<string> {
    try {
      // Get the flow to find the Google OAuth URL
      const flow = flowType === 'login' 
        ? await this.getLoginFlow(flowId)
        : await this.getRegistrationFlow(flowId);
      
      // Find the Google OAuth node in the flow
      const googleNode = flow.ui.nodes.find(
        node => node.attributes?.name === 'provider' && node.attributes?.value === 'google'
      );
      
      if (!googleNode) {
        throw new Error('Google OAuth not configured in flow');
      }
      
      // For OAuth, we need to submit a form to the action URL with provider=google
      // Instead of constructing a URL, we'll trigger the OAuth flow by submitting the form
      return flow.ui.action;
    } catch (error) {
      console.error('Error getting Google auth URL:', error);
      throw error;
    }
  }
}