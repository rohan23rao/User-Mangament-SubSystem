import React, { createContext, useContext, useState, useEffect, ReactNode, useRef } from 'react';
import { AuthService } from '../services/auth';
import { User } from '../types/user';
import { KratosSession } from '../types/auth';
import { notifications } from '@mantine/notifications';

interface AuthContextType {
  user: User | null;
  session: KratosSession | null;
  loading: boolean;
  isAuthenticated: boolean;
  canCreateOrganizations: () => boolean;
  login: (email: string, password: string) => Promise<void>;
  loginWithGoogle: () => Promise<void>;
  register: (email: string, password: string, firstName: string, lastName: string) => Promise<void>;
  registerWithGoogle: () => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [session, setSession] = useState<KratosSession | null>(null);
  const [loading, setLoading] = useState(true);
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const isCheckingAuth = useRef(false);
  const lastAuthCheck = useRef(0);
  const consecutiveFailures = useRef(0);

  const isAuthenticated = !!user && !!session && !isLoggingOut;

  useEffect(() => {
    // Only check auth status on mount if we're not in the middle of logging out
    if (!isLoggingOut) {
      checkAuthStatus();
    }
  }, []);

  // Also check auth status when URL changes (for OAuth callbacks)
  useEffect(() => {
    if (isLoggingOut) return; // Don't check during logout
    
    const urlParams = new URLSearchParams(window.location.search);
    const flowParam = urlParams.get('flow');
    const currentPath = window.location.pathname;
    
    // Only recheck auth for OAuth flows (not verification/recovery), and only once per flow
    if (flowParam && !currentPath.includes('/verification') && !currentPath.includes('/recovery') && !currentPath.includes('/error')) {
      console.log('OAuth flow parameter detected, rechecking auth status...');
      setTimeout(() => checkAuthStatus(), 1000); // Longer delay for OAuth
    }
  }, [window.location.pathname, isLoggingOut]); // Only trigger on path changes

  const checkAuthStatus = async () => {
    // Don't check auth if we're logging out
    if (isLoggingOut) {
      console.log('Skipping auth check - logout in progress');
      return;
    }

    // Prevent multiple simultaneous auth checks
    if (isCheckingAuth.current) {
      console.log('Auth check already in progress, skipping...');
      return;
    }

    // Stop checking after too many consecutive failures
    if (consecutiveFailures.current >= 3) {
      console.log('Too many consecutive auth failures, stopping checks');
      setLoading(false);
      return;
    }

    // Debounce auth checks (prevent rapid successive calls)
    const now = Date.now();
    if (now - lastAuthCheck.current < 5000) { // Increased to 5 seconds
      console.log('Auth check debounced, skipping...');
      return;
    }

    try {
      isCheckingAuth.current = true;
      lastAuthCheck.current = now;
      setLoading(true);
      
      const [sessionData, userData] = await Promise.all([
        AuthService.getSession(),
        AuthService.getCurrentUser(),
      ]);
      setSession(sessionData);
      setUser(userData);
      consecutiveFailures.current = 0; // Reset failure count on success
      console.log('Auth check successful');
    } catch (error: any) {
      console.log('Auth check failed:', error.message || error);
      consecutiveFailures.current += 1;
      
      // If authentication fails, clear any invalid cookies but don't keep retrying
      if (error.response?.status === 401) {
        console.log('Clearing invalid session due to 401 error');
        AuthService.clearSessionCookies();
      }
      
      setSession(null);
      setUser(null);
    } finally {
      isCheckingAuth.current = false;
      setLoading(false);
    }
  };

  const login = async (email: string, password: string): Promise<void> => {
    try {
      setLoading(true);
      const flow = await AuthService.createLoginFlow();
      const result = await AuthService.submitLogin(flow.id, email, password);
      
      console.log('Login submission result:', result);
      
      // For browser flows, session is established via cookies
      // Always refresh user data after successful submission
      await refreshUser();
      notifications.show({
        title: 'Login Successful',
        message: 'Welcome back!',
        color: 'green',
      });
    } catch (error: any) {
      console.error('Login error:', error);
      const message = error.response?.data?.ui?.messages?.[0]?.text || 'Login failed';
      notifications.show({
        title: 'Login Failed',
        message,
        color: 'red',
      });
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const loginWithGoogle = async (): Promise<void> => {
    try {
      setLoading(true);
      const flow = await AuthService.createLoginFlow();
      
      // Create a form and submit it to trigger Google OAuth
      const form = document.createElement('form');
      form.method = 'POST';
      form.action = flow.ui.action;
      
      const providerInput = document.createElement('input');
      providerInput.type = 'hidden';
      providerInput.name = 'provider';
      providerInput.value = 'google';
      form.appendChild(providerInput);
      
      document.body.appendChild(form);
      form.submit();
    } catch (error: any) {
      notifications.show({
        title: 'Google Login Failed',
        message: 'Failed to initiate Google login',
        color: 'red',
      });
      setLoading(false);
      throw error;
    }
  };

  const register = async (
    email: string, 
    password: string, 
    firstName: string, 
    lastName: string
  ): Promise<void> => {
    try {
      setLoading(true);
      const flow = await AuthService.createRegistrationFlow();
      const result = await AuthService.submitRegistration(flow.id, email, password, firstName, lastName);
      
      console.log('Registration submission result:', result);
      
      // Check if verification is required
      if (result?.continue_with?.some((item: any) => item.action === 'show_verification_ui')) {
        // Verification required - don't refresh user data yet
        notifications.show({
          title: 'Registration Successful',
          message: 'Please check your email for verification instructions.',
          color: 'green',
        });
      } else {
        // No verification required - establish session
        await refreshUser();
        notifications.show({
          title: 'Registration Successful',
          message: 'Welcome to the platform!',
          color: 'green',
        });
      }
    } catch (error: any) {
      console.error('Registration error:', error);
      const message = error.response?.data?.ui?.messages?.[0]?.text || 'Registration failed';
      notifications.show({
        title: 'Registration Failed',
        message,
        color: 'red',
      });
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const registerWithGoogle = async (): Promise<void> => {
    try {
      setLoading(true);
      const flow = await AuthService.createRegistrationFlow();
      
      // Create a form and submit it to trigger Google OAuth
      const form = document.createElement('form');
      form.method = 'POST';
      form.action = flow.ui.action;
      
      const providerInput = document.createElement('input');
      providerInput.type = 'hidden';
      providerInput.name = 'provider';
      providerInput.value = 'google';
      form.appendChild(providerInput);
      
      document.body.appendChild(form);
      form.submit();
    } catch (error: any) {
      notifications.show({
        title: 'Google Registration Failed',
        message: 'Failed to initiate Google registration',
        color: 'red',
      });
      setLoading(false);
      throw error;
    }
  };

  const logout = async (): Promise<void> => {
    try {
      console.log('Starting logout process...');
      setIsLoggingOut(true);
      setLoading(true);
      
      // Clear local state first
      setUser(null);
      setSession(null);
      
      // Reset auth failure counter
      consecutiveFailures.current = 0;
      
      // Then try to logout from Kratos
      await AuthService.logout();
      
      notifications.show({
        title: 'Logged Out',
        message: 'You have been successfully logged out',
        color: 'blue',
      });
      
      // Navigate to login page
      window.location.href = '/login';
    } catch (error: any) {
      console.error('Logout error:', error);
      notifications.show({
        title: 'Logout Failed',
        message: 'Failed to logout properly, but local session cleared',
        color: 'orange',
      });
      
      // Still navigate to login even if logout failed
      window.location.href = '/login';
    } finally {
      setLoading(false);
      // Don't reset isLoggingOut here since we're navigating away
    }
  };

  const refreshUser = async (): Promise<void> => {
    if (isLoggingOut) {
      console.log('Skipping refresh user - logout in progress');
      return;
    }
    
    try {
      const [sessionData, userData] = await Promise.all([
        AuthService.getSession(),
        AuthService.getCurrentUser(),
      ]);
      setSession(sessionData);
      setUser(userData);
      consecutiveFailures.current = 0; // Reset failure count on success
    } catch (error) {
      console.log('Refresh user failed:', error);
      consecutiveFailures.current += 1;
      setSession(null);
      setUser(null);
    }
  };

  const canCreateOrganizations = (): boolean => {
    if (!user) return false;
    
    // User can create organizations if they are already an admin of at least one organization
    // OR if they are the only user in the system (bootstrap scenario)
    const hasAdminRole = user.organizations?.some(org => org.role === 'admin') || false;
    
    // For bootstrap scenario, we check if user has no organizations
    // This indicates they might be the first user who should be able to create the first org
    const hasNoOrganizations = !user.organizations || user.organizations.length === 0;
    
    return hasAdminRole || hasNoOrganizations;
  };

  const value: AuthContextType = {
    user,
    session,
    loading,
    isAuthenticated,
    canCreateOrganizations,
    login,
    loginWithGoogle,
    register,
    registerWithGoogle,
    logout,
    refreshUser,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
