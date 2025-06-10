#!/bin/bash

# Complete Mantine v7 Migration Script
# This script fixes all compatibility issues between Mantine v6 and v7

cd ~/umange/late-claude/frontend

echo "🔧 Starting Mantine v7 migration..."
echo "📝 Creating all component files with v7 compatibility..."

# 1. Fix App.tsx - Main entry point
cat > src/App.tsx << 'EOF'
import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { MantineProvider, createTheme } from '@mantine/core';
import { useLocalStorage } from '@mantine/hooks';
import { Notifications } from '@mantine/notifications';
import { ModalsProvider } from '@mantine/modals';

import { AuthProvider } from './hooks/useAuth';
import { AppShell } from './components/layout/AppShell';
import { LoginPage } from './pages/LoginPage';
import { RegisterPage } from './pages/RegisterPage';
import { Dashboard } from './pages/Dashboard';
import { UsersPage } from './pages/UsersPage';
import { OrganizationsPage } from './pages/OrganizationsPage';
import { OrganizationDetailsPage } from './pages/OrganizationDetailsPage';
import { ProfilePage } from './pages/ProfilePage';
import { ProtectedRoute } from './components/auth/ProtectedRoute';

import '@mantine/core/styles.css';
import '@mantine/notifications/styles.css';
import '@mantine/dates/styles.css';

const theme = createTheme({
  primaryColor: 'blue',
  fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif',
  headings: {
    fontFamily: 'Greycliff CF, Inter, -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif',
    fontWeight: '700',
    textWrap: 'wrap',
    sizes: {
      h1: { fontSize: '2rem' },
      h2: { fontSize: '1.5rem' },
      h3: { fontSize: '1.25rem' },
      h4: { fontSize: '1.125rem' },
      h5: { fontSize: '1rem' },
      h6: { fontSize: '0.875rem' },
    },
  },
});

function App() {
  const [colorScheme, setColorScheme] = useLocalStorage<'light' | 'dark'>({
    key: 'mantine-color-scheme',
    defaultValue: 'light',
  });

  const toggleColorScheme = () => setColorScheme(colorScheme === 'dark' ? 'light' : 'dark');

  return (
    <MantineProvider theme={theme} forceColorScheme={colorScheme}>
      <ModalsProvider>
        <Notifications />
        <AuthProvider>
          <Router>
            <Routes>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />
              <Route
                path="/*"
                element={
                  <ProtectedRoute>
                    <AppShell toggleColorScheme={toggleColorScheme}>
                      <Routes>
                        <Route path="/" element={<Navigate to="/dashboard" replace />} />
                        <Route path="/dashboard" element={<Dashboard />} />
                        <Route path="/profile" element={<ProfilePage />} />
                        <Route path="/users" element={<UsersPage />} />
                        <Route path="/organizations" element={<OrganizationsPage />} />
                        <Route path="/organizations/:id" element={<OrganizationDetailsPage />} />
                      </Routes>
                    </AppShell>
                  </ProtectedRoute>
                }
              />
            </Routes>
          </Router>
        </AuthProvider>
      </ModalsProvider>
    </MantineProvider>
  );
}

export default App;
EOF

# 2. Fix ProtectedRoute.tsx
cat > src/components/auth/ProtectedRoute.tsx << 'EOF'
import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { LoadingOverlay, Center } from '@mantine/core';
import { useAuth } from '../../hooks/useAuth';

interface ProtectedRouteProps {
  children: React.ReactNode;
  adminOnly?: boolean;
}

export function ProtectedRoute({ children, adminOnly = false }: ProtectedRouteProps) {
  const { isAuthenticated, loading, user } = useAuth();
  const location = useLocation();

  if (loading) {
    return (
      <Center style={{ height: '100vh' }}>
        <LoadingOverlay visible={true} />
      </Center>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  if (adminOnly) {
    const isAdmin = user?.organizations?.some(org => org.role === 'admin') || false;
    if (!isAdmin) {
      return <Navigate to="/dashboard" replace />;
    }
  }

  return <>{children}</>;
}
EOF

# 3. Fix AppShell.tsx - Updated for v7 API
cat > src/components/layout/AppShell.tsx << 'EOF'
import React, { useState } from 'react';
import {
  AppShell as MantineAppShell,
  Text,
  Burger,
  Group,
  ActionIcon,
  Menu,
  Avatar,
  UnstyledButton,
  Box,
  useMantineColorScheme,
} from '@mantine/core';
import {
  IconSun,
  IconMoonStars,
  IconChevronDown,
  IconUser,
  IconSettings,
  IconLogout,
} from '@tabler/icons-react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { Navigation } from './Navigation';

interface AppShellLayoutProps {
  children: React.ReactNode;
  toggleColorScheme: () => void;
}

export function AppShell({ children, toggleColorScheme }: AppShellLayoutProps) {
  const [opened, setOpened] = useState(false);
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const { colorScheme } = useMantineColorScheme();

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  const getUserInitials = () => {
    if (!user) return '';
    return `${user.first_name?.[0] || ''}${user.last_name?.[0] || ''}`.toUpperCase();
  };

  const getUserDisplayName = () => {
    if (!user) return 'User';
    return `${user.first_name} ${user.last_name}`.trim() || user.email;
  };

  return (
    <MantineAppShell
      header={{ height: 60 }}
      navbar={{ width: 250, breakpoint: 'sm', collapsed: { mobile: !opened } }}
      padding="md"
    >
      <MantineAppShell.Header>
        <Group h="100%" px="md" justify="space-between">
          <Group>
            <Burger opened={opened} onClick={() => setOpened(!opened)} hiddenFrom="sm" size="sm" />
            <Text size="xl" fw={700}>
              UserMS
            </Text>
          </Group>

          <Group>
            <ActionIcon
              variant="default"
              onClick={toggleColorScheme}
              size="lg"
              title="Toggle color scheme"
            >
              {colorScheme === 'dark' ? <IconSun size="1.2rem" /> : <IconMoonStars size="1.2rem" />}
            </ActionIcon>

            <Menu shadow="md" width={200} position="bottom-end">
              <Menu.Target>
                <UnstyledButton>
                  <Group gap={7}>
                    <Avatar size={30} radius="xl" color="blue">
                      {getUserInitials()}
                    </Avatar>
                    <Box style={{ flex: 1 }}>
                      <Text size="sm" fw={500}>
                        {getUserDisplayName()}
                      </Text>
                      <Text c="dimmed" size="xs">
                        {user?.email}
                      </Text>
                    </Box>
                    <IconChevronDown size="1rem" />
                  </Group>
                </UnstyledButton>
              </Menu.Target>

              <Menu.Dropdown>
                <Menu.Label>Account</Menu.Label>
                <Menu.Item leftSection={<IconUser size="0.9rem" />} onClick={() => navigate('/profile')}>
                  Profile
                </Menu.Item>
                <Menu.Item leftSection={<IconSettings size="0.9rem" />} onClick={() => navigate('/settings')}>
                  Settings
                </Menu.Item>
                <Menu.Divider />
                <Menu.Item
                  leftSection={<IconLogout size="0.9rem" />}
                  onClick={handleLogout}
                  color="red"
                >
                  Logout
                </Menu.Item>
              </Menu.Dropdown>
            </Menu>
          </Group>
        </Group>
      </MantineAppShell.Header>

      <MantineAppShell.Navbar p="md">
        <Navigation onNavigate={() => setOpened(false)} />
      </MantineAppShell.Navbar>

      <MantineAppShell.Main>{children}</MantineAppShell.Main>
    </MantineAppShell>
  );
}
EOF

# 4. Fix Navigation.tsx
cat > src/components/layout/Navigation.tsx << 'EOF'
import React from 'react';
import { Group, Text, ThemeIcon, UnstyledButton, Badge } from '@mantine/core';
import {
  IconDashboard,
  IconUsers,
  IconBuilding,
  IconUser,
  IconSettings,
  IconChartBar,
} from '@tabler/icons-react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';

interface NavigationProps {
  onNavigate?: () => void;
}

interface NavItem {
  icon: React.ReactNode;
  label: string;
  path: string;
  badge?: string;
  adminOnly?: boolean;
}

export function Navigation({ onNavigate }: NavigationProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { user } = useAuth();

  const isAdmin = user?.organizations?.some(org => org.role === 'admin') || false;

  const navItems: NavItem[] = [
    {
      icon: <IconDashboard size="1rem" />,
      label: 'Dashboard',
      path: '/dashboard',
    },
    {
      icon: <IconUser size="1rem" />,
      label: 'Profile',
      path: '/profile',
    },
    {
      icon: <IconBuilding size="1rem" />,
      label: 'Organizations',
      path: '/organizations',
      badge: user?.organizations?.length.toString(),
    },
    {
      icon: <IconUsers size="1rem" />,
      label: 'Users',
      path: '/users',
      adminOnly: true,
    },
    {
      icon: <IconChartBar size="1rem" />,
      label: 'Analytics',
      path: '/analytics',
      adminOnly: true,
    },
    {
      icon: <IconSettings size="1rem" />,
      label: 'Settings',
      path: '/settings',
    },
  ];

  const handleNavigation = (path: string) => {
    navigate(path);
    onNavigate?.();
  };

  const filteredNavItems = navItems.filter(item => !item.adminOnly || isAdmin);

  return (
    <div>
      <Group mb="md">
        <ThemeIcon variant="gradient" size={30}>
          <IconDashboard size="1rem" />
        </ThemeIcon>
        <Text fw={700} size="lg">
          Navigation
        </Text>
      </Group>

      {filteredNavItems.map((item) => (
        <UnstyledButton
          key={item.path}
          onClick={() => handleNavigation(item.path)}
          style={{
            display: 'block',
            width: '100%',
            padding: 'var(--mantine-spacing-xs)',
            borderRadius: 'var(--mantine-radius-sm)',
            backgroundColor: location.pathname === item.path ? 'var(--mantine-color-gray-1)' : 'transparent',
          }}
        >
          <Group>
            <ThemeIcon 
              variant={location.pathname === item.path ? 'filled' : 'light'} 
              size={30}
            >
              {item.icon}
            </ThemeIcon>
            <Text size="sm" fw={location.pathname === item.path ? 600 : 400}>
              {item.label}
            </Text>
            {item.badge && (
              <Badge size="sm" variant="filled" color="blue">
                {item.badge}
              </Badge>
            )}
          </Group>
        </UnstyledButton>
      ))}
    </div>
  );
}
EOF

# 5. Fix LoginPage.tsx
cat > src/pages/LoginPage.tsx << 'EOF'
import React, { useState, useEffect } from 'react';
import { useNavigate, Link, useLocation } from 'react-router-dom';
import {
  Container,
  Paper,
  TextInput,
  PasswordInput,
  Button,
  Title,
  Text,
  Anchor,
  Stack,
  Divider,
  Center,
  Alert,
  LoadingOverlay,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { IconMail, IconLock, IconBrandGoogle, IconInfoCircle } from '@tabler/icons-react';
import { useAuth } from '../hooks/useAuth';

interface LoginFormValues {
  email: string;
  password: string;
}

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { login, loginWithGoogle, loading, isAuthenticated } = useAuth();
  const [googleLoading, setGoogleLoading] = useState(false);

  useEffect(() => {
    if (isAuthenticated) {
      const from = location.state?.from?.pathname || '/dashboard';
      navigate(from, { replace: true });
    }
  }, [isAuthenticated, navigate, location.state]);

  const form = useForm<LoginFormValues>({
    initialValues: {
      email: '',
      password: '',
    },
    validate: {
      email: (value) => (/^\S+@\S+$/.test(value) ? null : 'Invalid email'),
      password: (value) => (value.length < 6 ? 'Password must be at least 6 characters' : null),
    },
  });

  const handleSubmit = async (values: LoginFormValues) => {
    try {
      await login(values.email, values.password);
      navigate('/dashboard');
    } catch (error) {
      // Error is handled in the login function
    }
  };

  const handleGoogleLogin = async () => {
    try {
      setGoogleLoading(true);
      await loginWithGoogle();
    } catch (error) {
      setGoogleLoading(false);
    }
  };

  return (
    <Container size={420} my={40}>
      <Title ta="center" fw={900}>
        Welcome back!
      </Title>
      <Text c="dimmed" size="sm" ta="center" mt={5}>
        Do not have an account yet?{' '}
        <Anchor size="sm" component={Link} to="/register">
          Create account
        </Anchor>
      </Text>

      <Paper withBorder shadow="md" p={30} mt={30} radius="md" style={{ position: 'relative' }}>
        <LoadingOverlay visible={loading} />
        
        {location.state?.from && (
          <Alert icon={<IconInfoCircle size="1rem" />} mb="md" color="blue">
            Please log in to access {location.state.from.pathname}
          </Alert>
        )}

        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack>
            <TextInput
              label="Email"
              placeholder="your@email.com"
              leftSection={<IconMail size="1rem" />}
              {...form.getInputProps('email')}
              disabled={loading}
            />

            <PasswordInput
              label="Password"
              placeholder="Your password"
              leftSection={<IconLock size="1rem" />}
              {...form.getInputProps('password')}
              disabled={loading}
            />

            <Button type="submit" fullWidth loading={loading}>
              Sign in
            </Button>
          </Stack>
        </form>

        <Divider label="Or continue with" labelPosition="center" my="lg" />

        <Button
          variant="light"
          fullWidth
          leftSection={<IconBrandGoogle size="1rem" />}
          onClick={handleGoogleLogin}
          loading={googleLoading}
          disabled={loading}
          color="red"
        >
          Continue with Google
        </Button>

        <Center mt="md">
          <Text size="sm" c="dimmed">
            Forgot your password?{' '}
            <Anchor size="sm" href="#" onClick={(e) => e.preventDefault()}>
              Reset password
            </Anchor>
          </Text>
        </Center>
      </Paper>
    </Container>
  );
}
EOF

# 6. Fix RegisterPage.tsx
cat > src/pages/RegisterPage.tsx << 'EOF'
import React, { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import {
  Container,
  Paper,
  TextInput,
  PasswordInput,
  Button,
  Title,
  Text,
  Anchor,
  Stack,
  Divider,
  Group,
  LoadingOverlay,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { IconMail, IconLock, IconUser, IconBrandGoogle } from '@tabler/icons-react';
import { useAuth } from '../hooks/useAuth';

interface RegisterFormValues {
  firstName: string;
  lastName: string;
  email: string;
  password: string;
  confirmPassword: string;
}

export function RegisterPage() {
  const navigate = useNavigate();
  const { register, registerWithGoogle, loading, isAuthenticated } = useAuth();
  const [googleLoading, setGoogleLoading] = useState(false);

  useEffect(() => {
    if (isAuthenticated) {
      navigate('/dashboard', { replace: true });
    }
  }, [isAuthenticated, navigate]);

  const form = useForm<RegisterFormValues>({
    initialValues: {
      firstName: '',
      lastName: '',
      email: '',
      password: '',
      confirmPassword: '',
    },
    validate: {
      firstName: (value) => (value.length < 2 ? 'First name must be at least 2 characters' : null),
      lastName: (value) => (value.length < 2 ? 'Last name must be at least 2 characters' : null),
      email: (value) => (/^\S+@\S+$/.test(value) ? null : 'Invalid email'),
      password: (value) => (value.length < 6 ? 'Password must be at least 6 characters' : null),
      confirmPassword: (value, values) =>
        value !== values.password ? 'Passwords do not match' : null,
    },
  });

  const handleSubmit = async (values: RegisterFormValues) => {
    try {
      await register(values.email, values.password, values.firstName, values.lastName);
      navigate('/dashboard');
    } catch (error) {
      // Error is handled in the register function
    }
  };

  const handleGoogleRegister = async () => {
    try {
      setGoogleLoading(true);
      await registerWithGoogle();
    } catch (error) {
      setGoogleLoading(false);
    }
  };

  return (
    <Container size={420} my={40}>
      <Title ta="center" fw={900}>
        Create your account
      </Title>
      <Text c="dimmed" size="sm" ta="center" mt={5}>
        Already have an account?{' '}
        <Anchor size="sm" component={Link} to="/login">
          Sign in
        </Anchor>
      </Text>

      <Paper withBorder shadow="md" p={30} mt={30} radius="md" style={{ position: 'relative' }}>
        <LoadingOverlay visible={loading} />

        <Button
          variant="light"
          fullWidth
          leftSection={<IconBrandGoogle size="1rem" />}
          onClick={handleGoogleRegister}
          loading={googleLoading}
          disabled={loading}
          color="red"
          mb="md"
        >
          Continue with Google
        </Button>

        <Divider label="Or register with email" labelPosition="center" my="lg" />

        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack>
            <Group grow>
              <TextInput
                label="First name"
                placeholder="John"
                leftSection={<IconUser size="1rem" />}
                {...form.getInputProps('firstName')}
                disabled={loading}
              />
              <TextInput
                label="Last name"
                placeholder="Doe"
                leftSection={<IconUser size="1rem" />}
                {...form.getInputProps('lastName')}
                disabled={loading}
              />
            </Group>

            <TextInput
              label="Email"
              placeholder="your@email.com"
              leftSection={<IconMail size="1rem" />}
              {...form.getInputProps('email')}
              disabled={loading}
            />

            <PasswordInput
              label="Password"
              placeholder="Your password"
              leftSection={<IconLock size="1rem" />}
              {...form.getInputProps('password')}
              disabled={loading}
            />

            <PasswordInput
              label="Confirm password"
              placeholder="Confirm your password"
              leftSection={<IconLock size="1rem" />}
              {...form.getInputProps('confirmPassword')}
              disabled={loading}
            />

            <Button type="submit" fullWidth loading={loading}>
              Create account
            </Button>
          </Stack>
        </form>

        <Text size="xs" c="dimmed" ta="center" mt="md">
          By creating an account, you agree to our{' '}
          <Anchor size="xs" href="#" onClick={(e) => e.preventDefault()}>
            Terms of Service
          </Anchor>{' '}
          and{' '}
          <Anchor size="xs" href="#" onClick={(e) => e.preventDefault()}>
            Privacy Policy
          </Anchor>
        </Text>
      </Paper>
    </Container>
  );
}
EOF

# 7. Create services/auth.ts
cat > src/services/auth.ts << 'EOF'
import axios from 'axios';
import { KratosSession, LoginFlow, RegistrationFlow } from '../types/auth';
import { User } from '../types/user';

const KRATOS_PUBLIC_URL = process.env.REACT_APP_KRATOS_PUBLIC_URL || 'http://localhost:4433';
const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:3000';

axios.defaults.withCredentials = true;

export class AuthService {
  static async createLoginFlow(): Promise<LoginFlow> {
    const response = await axios.get(`${KRATOS_PUBLIC_URL}/self-service/login/api`);
    return response.data;
  }

  static async createRegistrationFlow(): Promise<RegistrationFlow> {
    const response = await axios.get(`${KRATOS_PUBLIC_URL}/self-service/registration/api`);
    return response.data;
  }

  static async submitLogin(flowId: string, email: string, password: string): Promise<any> {
    const response = await axios.post(
      `${KRATOS_PUBLIC_URL}/self-service/login?flow=${flowId}`,
      {
        method: 'password',
        password_identifier: email,
        password: password,
      },
      {
        headers: {
          'Content-Type': 'application/json',
        },
      }
    );
    return response.data;
  }

  static async submitRegistration(
    flowId: string, 
    email: string, 
    password: string, 
    firstName: string, 
    lastName: string
  ): Promise<any> {
    const response = await axios.post(
      `${KRATOS_PUBLIC_URL}/self-service/registration?flow=${flowId}`,
      {
        method: 'password',
        password: password,
        traits: {
          email: email,
          name: {
            first: firstName,
            last: lastName,
          },
        },
      },
      {
        headers: {
          'Content-Type': 'application/json',
        },
      }
    );
    return response.data;
  }

  static async getGoogleAuthUrl(flowId: string, flowType: 'login' | 'registration'): Promise<string> {
    const endpoint = flowType === 'login' ? 'login' : 'registration';
    
    const response = await axios.post(
      `${KRATOS_PUBLIC_URL}/self-service/${endpoint}?flow=${flowId}`,
      {
        method: 'oidc',
        provider: 'google',
      },
      {
        headers: {
          'Content-Type': 'application/json',
        },
        maxRedirects: 0,
        validateStatus: (status) => status === 302 || status === 200,
      }
    );

    if (response.status === 302) {
      return response.headers.location;
    }

    throw new Error('Failed to get Google OAuth URL');
  }

  static async getSession(): Promise<KratosSession> {
    const response = await axios.get(`${KRATOS_PUBLIC_URL}/sessions/whoami`);
    return response.data;
  }

  static async getCurrentUser(): Promise<User> {
    const response = await axios.get(`${API_URL}/api/whoami`);
    return response.data;
  }

  static async logout(): Promise<void> {
    await axios.post(`${API_URL}/auth/logout`);
  }

  static async isAuthenticated(): Promise<boolean> {
    try {
      await this.getSession();
      return true;
    } catch (error) {
      return false;
    }
  }
}

axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
EOF

# 8. Create services/api.ts
cat > src/services/api.ts << 'EOF'
import axios from 'axios';
import { User } from '../types/user';
import { Organization, CreateOrgRequest, InviteUserRequest, UpdateMemberRoleRequest, Member } from '../types/organization';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:3000';

axios.defaults.withCredentials = true;
axios.defaults.baseURL = API_URL;

export class ApiService {
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

  static async getOrganizations(): Promise<Organization[]> {
    const response = await axios.get('/api/organizations');
    return response.data;
  }

  static async getOrganization(id: string): Promise<Organization> {
    const response = await axios.get(`/api/organizations/${id}`);
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
    const response = await axios.put(`/api/organizations/${organizationId}/members/${userId}/role`, data);
    return response.data;
  }

  static async healthCheck(): Promise<any> {
    const response = await axios.get('/health');
    return response.data;
  }
}

axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
EOF

# 9. Create hooks/useAuth.tsx
cat > src/hooks/useAuth.tsx << 'EOF'
import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { AuthService } from '../services/auth';
import { User } from '../types/user';
import { KratosSession } from '../types/auth';
import { notifications } from '@mantine/notifications';

interface AuthContextType {
  user: User | null;
  session: KratosSession | null;
  loading: boolean;
  isAuthenticated: boolean;
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

  const isAuthenticated = !!user && !!session;

  useEffect(() => {
    checkAuthStatus();
  }, []);

  const checkAuthStatus = async () => {
    try {
      setLoading(true);
      const [sessionData, userData] = await Promise.all([
        AuthService.getSession(),
        AuthService.getCurrentUser(),
      ]);
      setSession(sessionData);
      setUser(userData);
    } catch (error) {
      setSession(null);
      setUser(null);
    } finally {
      setLoading(false);
    }
  };

  const login = async (email: string, password: string): Promise<void> => {
    try {
      setLoading(true);
      const flow = await AuthService.createLoginFlow();
      const result = await AuthService.submitLogin(flow.id, email, password);
      
      if (result.session_token) {
        await refreshUser();
        notifications.show({
          title: 'Login Successful',
          message: 'Welcome back!',
          color: 'green',
        });
      }
    } catch (error: any) {
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
      const googleUrl = await AuthService.getGoogleAuthUrl(flow.id, 'login');
      window.location.href = googleUrl;
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
      
      if (result.session_token) {
        await refreshUser();
        notifications.show({
          title: 'Registration Successful',
          message: 'Welcome to the platform!',
          color: 'green',
        });
      }
    } catch (error: any) {
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
      const googleUrl = await AuthService.getGoogleAuthUrl(flow.id, 'registration');
      window.location.href = googleUrl;
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
      setLoading(true);
      await AuthService.logout();
      setUser(null);
      setSession(null);
      notifications.show({
        title: 'Logged Out',
        message: 'You have been successfully logged out',
        color: 'blue',
      });
    } catch (error: any) {
      notifications.show({
        title: 'Logout Failed',
        message: 'Failed to logout properly',
        color: 'red',
      });
    } finally {
      setLoading(false);
    }
  };

  const refreshUser = async (): Promise<void> => {
    try {
      const [sessionData, userData] = await Promise.all([
        AuthService.getSession(),
        AuthService.getCurrentUser(),
      ]);
      setSession(sessionData);
      setUser(userData);
    } catch (error) {
      setSession(null);
      setUser(null);
    }
  };

  const value: AuthContextType = {
    user,
    session,
    loading,
    isAuthenticated,
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
EOF

# 10. Create Dashboard page with v7 compatibility
cat > src/pages/Dashboard.tsx << 'EOF'
import React from 'react';
import {
  Container,
  Title,
  Text,
  Grid,
  Card,
  Group,
  ThemeIcon,
  Progress,
  Badge,
  Stack,
  Button,
  SimpleGrid,
  Paper,
  Avatar,
  ActionIcon,
} from '@mantine/core';
import {
  IconUsers,
  IconBuilding,
  IconUserCheck,
  IconTrendingUp,
  IconPlus,
  IconEye,
  IconSettings,
} from '@tabler/icons-react';
import { useAuth } from '../hooks/useAuth';
import { useNavigate } from 'react-router-dom';

interface StatsCardProps {
  title: string;
  value: string | number;
  description: string;
  icon: React.ReactNode;
  color: string;
  progress?: number;
}

function StatsCard({ title, value, description, icon, color, progress }: StatsCardProps) {
  return (
    <Card withBorder p="xl" radius="md">
      <Group justify="space-between">
        <div>
          <Text c="dimmed" fw={700} size="xs" tt="uppercase">
            {title}
          </Text>
          <Text fw={700} size="xl">
            {value}
          </Text>
          <Text c="dimmed" size="xs">
            {description}
          </Text>
        </div>
        <ThemeIcon color={color} size={38} radius="md">
          {icon}
        </ThemeIcon>
      </Group>
      {progress !== undefined && (
        <Progress value={progress} mt="md" size="sm" color={color} />
      )}
    </Card>
  );
}

export function Dashboard() {
  const { user } = useAuth();
  const navigate = useNavigate();

  const userOrganizations = user?.organizations || [];
  const isAdmin = userOrganizations.some(org => org.role === 'admin');

  const stats = [
    {
      title: 'Organizations',
      value: userOrganizations.length,
      description: 'Active memberships',
      icon: <IconBuilding size="1.4rem" />,
      color: 'blue',
    },
    {
      title: 'Admin Roles',
      value: userOrganizations.filter(org => org.role === 'admin').length,
      description: 'Organizations you manage',
      icon: <IconUserCheck size="1.4rem" />,
      color: 'green',
    },
    {
      title: 'Member Since',
      value: new Date(user?.created_at || '').getFullYear(),
      description: 'Account creation year',
      icon: <IconUsers size="1.4rem" />,
      color: 'violet',
    },
    {
      title: 'Last Login',
      value: user?.last_login ? 'Recent' : 'Today',
      description: 'Activity status',
      icon: <IconTrendingUp size="1.4rem" />,
      color: 'orange',
    },
  ];

  const recentOrganizations = userOrganizations.slice(0, 3);

  return (
    <Container size="xl" py="xl">
      <Group justify="space-between" mb="xl">
        <div>
          <Title order={1} mb={4}>
            Welcome back, {user?.first_name}! 👋
          </Title>
          <Text c="dimmed" size="lg">
            Here's what's happening with your organizations today
          </Text>
        </div>
        <Button
          leftSection={<IconPlus size="1rem" />}
          onClick={() => navigate('/organizations')}
        >
          Create Organization
        </Button>
      </Group>

      <SimpleGrid
        cols={{ base: 1, xs: 2, md: 4 }}
        mb="xl"
      >
        {stats.map((stat) => (
          <StatsCard key={stat.title} {...stat} />
        ))}
      </SimpleGrid>

      <Grid>
        <Grid.Col span={{ base: 12, md: 8 }}>
          <Paper withBorder p="md" radius="md" mb="md">
            <Group justify="space-between" mb="md">
              <Title order={3}>Your Organizations</Title>
              <Button
                variant="subtle"
                size="sm"
                rightSection={<IconEye size="1rem" />}
                onClick={() => navigate('/organizations')}
              >
                View All
              </Button>
            </Group>

            {recentOrganizations.length > 0 ? (
              <Stack gap="sm">
                {recentOrganizations.map((org) => (
                  <Card key={org.org_id} withBorder radius="sm" p="md">
                    <Group justify="space-between">
                      <Group>
                        <Avatar color="blue" radius="xl">
                          {org.org_name.charAt(0).toUpperCase()}
                        </Avatar>
                        <div>
                          <Text fw={500} size="sm">
                            {org.org_name}
                          </Text>
                          <Text c="dimmed" size="xs">
                            {org.org_type} • Joined {new Date(org.joined_at).toLocaleDateString()}
                          </Text>
                        </div>
                      </Group>
                      <Group gap="xs">
                        <Badge
                          color={org.role === 'admin' ? 'green' : 'blue'}
                          variant="light"
                          size="sm"
                        >
                          {org.role}
                        </Badge>
                        <ActionIcon
                          size="sm"
                          variant="subtle"
                          onClick={() => navigate(`/organizations/${org.org_id}`)}
                        >
                          <IconEye size="1rem" />
                        </ActionIcon>
                      </Group>
                    </Group>
                  </Card>
                ))}
              </Stack>
            ) : (
              <Paper p="xl" radius="md" withBorder>
                <Stack align="center" gap="sm">
                  <ThemeIcon size={60} radius="xl" color="blue" variant="light">
                    <IconBuilding size="2rem" />
                  </ThemeIcon>
                  <Text fw={500} ta="center">
                    No organizations yet
                  </Text>
                  <Text c="dimmed" size="sm" ta="center">
                    Create your first organization or ask to be invited to one
                  </Text>
                  <Button
                    leftSection={<IconPlus size="1rem" />}
                    onClick={() => navigate('/organizations')}
                  >
                    Create Organization
                  </Button>
                </Stack>
              </Paper>
            )}
          </Paper>

          <Paper withBorder p="md" radius="md">
            <Title order={3} mb="md">
              Quick Actions
            </Title>
            <SimpleGrid cols={2} spacing="sm">
              <Button
                variant="light"
                leftSection={<IconBuilding size="1rem" />}
                onClick={() => navigate('/organizations')}
                fullWidth
              >
                Manage Organizations
              </Button>
              <Button
                variant="light"
                leftSection={<IconSettings size="1rem" />}
                onClick={() => navigate('/profile')}
                fullWidth
              >
                Edit Profile
              </Button>
              {isAdmin && (
                <>
                  <Button
                    variant="light"
                    leftSection={<IconUsers size="1rem" />}
                    onClick={() => navigate('/users')}
                    fullWidth
                  >
                    Manage Users
                  </Button>
                  <Button
                    variant="light"
                    leftSection={<IconTrendingUp size="1rem" />}
                    onClick={() => navigate('/analytics')}
                    fullWidth
                  >
                    View Analytics
                  </Button>
                </>
              )}
            </SimpleGrid>
          </Paper>
        </Grid.Col>

        <Grid.Col span={{ base: 12, md: 4 }}>
          <Paper withBorder p="md" radius="md" mb="md">
            <Group mb="md">
              <Avatar size={60} color="blue">
                {user?.first_name?.charAt(0)}{user?.last_name?.charAt(0)}
              </Avatar>
              <div style={{ flex: 1 }}>
                <Text fw={500}>
                  {user?.first_name} {user?.last_name}
                </Text>
                <Text c="dimmed" size="sm">
                  {user?.email}
                </Text>
                <Badge color="green" size="sm" mt={4}>
                  Active
                </Badge>
              </div>
            </Group>
            <Stack gap="xs">
              <Group justify="space-between">
                <Text size="sm" c="dimmed">
                  Time Zone
                </Text>
                <Text size="sm">{user?.time_zone}</Text>
              </Group>
              <Group justify="space-between">
                <Text size="sm" c="dimmed">
                  UI Mode
                </Text>
                <Text size="sm">{user?.ui_mode}</Text>
              </Group>
              <Group justify="space-between">
                <Text size="sm" c="dimmed">
                  Member Since
                </Text>
                <Text size="sm">
                  {new Date(user?.created_at || '').toLocaleDateString()}
                </Text>
              </Group>
            </Stack>
            <Button
              variant="light"
              fullWidth
              mt="md"
              onClick={() => navigate('/profile')}
            >
              Edit Profile
            </Button>
          </Paper>

          <Paper withBorder p="md" radius="md">
            <Title order={4} mb="md">
              Recent Activity
            </Title>
            <Stack gap="sm">
              <Group>
                <ThemeIcon size="sm" color="blue" variant="light">
                  <IconUserCheck size="0.8rem" />
                </ThemeIcon>
                <Text size="sm">Account created</Text>
              </Group>
              {user?.last_login && (
                <Group>
                  <ThemeIcon size="sm" color="green" variant="light">
                    <IconTrendingUp size="0.8rem" />
                  </ThemeIcon>
                  <Text size="sm">
                    Last login: {new Date(user.last_login).toLocaleDateString()}
                  </Text>
                </Group>
              )}
              {userOrganizations.length > 0 && (
                <Group>
                  <ThemeIcon size="sm" color="violet" variant="light">
                    <IconBuilding size="0.8rem" />
                  </ThemeIcon>
                  <Text size="sm">
                    Member of {userOrganizations.length} organization{userOrganizations.length !== 1 ? 's' : ''}
                  </Text>
                </Group>
              )}
            </Stack>
          </Paper>
        </Grid.Col>
      </Grid>
    </Container>
  );
}
EOF

# 11. Create OrganizationsPage with v7 compatibility
cat > src/pages/OrganizationsPage.tsx << 'EOF'
import React, { useState, useEffect } from 'react';
import {
  Container,
  Title,
  Button,
  Group,
  Card,
  Text,
  Badge,
  Grid,
  Stack,
  Modal,
  TextInput,
  Textarea,
  Select,
  LoadingOverlay,
  ActionIcon,
  Avatar,
  Paper,
  ThemeIcon,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { notifications } from '@mantine/notifications';
import {
  IconPlus,
  IconBuilding,
  IconUsers,
  IconEye,
  IconEdit,
  IconTrash,
  IconCrown,
} from '@tabler/icons-react';
import { useNavigate } from 'react-router-dom';
import { ApiService } from '../services/api';
import { Organization, CreateOrgRequest } from '../types/organization';
import { useAuth } from '../hooks/useAuth';
import { modals } from '@mantine/modals';

export function OrganizationsPage() {
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [loading, setLoading] = useState(true);
  const [createModalOpened, setCreateModalOpened] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const navigate = useNavigate();
  const { user } = useAuth();

  const form = useForm<CreateOrgRequest>({
    initialValues: {
      name: '',
      description: '',
      org_type: 'organization',
      data: {},
    },
    validate: {
      name: (value) => (value.length < 2 ? 'Name must be at least 2 characters' : null),
      description: (value) => (value.length < 10 ? 'Description must be at least 10 characters' : null),
    },
  });

  useEffect(() => {
    loadOrganizations();
  }, []);

  const loadOrganizations = async () => {
    try {
      setLoading(true);
      const data = await ApiService.getOrganizations();
      setOrganizations(data);
    } catch (error) {
      notifications.show({
        title: 'Error',
        message: 'Failed to load organizations',
        color: 'red',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleCreateOrganization = async (values: CreateOrgRequest) => {
    try {
      setSubmitting(true);
      await ApiService.createOrganization(values);
      notifications.show({
        title: 'Success',
        message: 'Organization created successfully',
        color: 'green',
      });
      setCreateModalOpened(false);
      form.reset();
      loadOrganizations();
    } catch (error: any) {
      notifications.show({
        title: 'Error',
        message: error.response?.data?.message || 'Failed to create organization',
        color: 'red',
      });
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteOrganization = (org: Organization) => {
    modals.openConfirmModal({
      title: 'Delete Organization',
      children: (
        <Text size="sm">
          Are you sure you want to delete <strong>{org.name}</strong>? This action cannot be undone.
        </Text>
      ),
      labels: { confirm: 'Delete', cancel: 'Cancel' },
      confirmProps: { color: 'red' },
      onConfirm: async () => {
        try {
          await ApiService.deleteOrganization(org.id);
          notifications.show({
            title: 'Success',
            message: 'Organization deleted successfully',
            color: 'green',
          });
          loadOrganizations();
        } catch (error: any) {
          notifications.show({
            title: 'Error',
            message: error.response?.data?.message || 'Failed to delete organization',
            color: 'red',
          });
        }
      },
    });
  };

  const getUserRole = (orgId: string) => {
    const userOrg = user?.organizations?.find(org => org.org_id === orgId);
    return userOrg?.role || 'member';
  };

  const isOwner = (org: Organization) => {
    return org.owner_id === user?.id;
  };

  const isAdmin = (orgId: string) => {
    const role = getUserRole(orgId);
    return role === 'admin' || isOwner(organizations.find(o => o.id === orgId)!);
  };

  return (
    <Container size="xl" py="xl">
      <Group justify="space-between" mb="xl">
        <div>
          <Title order={1}>Organizations</Title>
          <Text c="dimmed" size="lg">
            Manage your organizations and memberships
          </Text>
        </div>
        <Button
          leftSection={<IconPlus size="1rem" />}
          onClick={() => setCreateModalOpened(true)}
        >
          Create Organization
        </Button>
      </Group>

      {loading ? (
        <Paper p="xl" radius="md" withBorder>
          <LoadingOverlay visible={true} />
        </Paper>
      ) : organizations.length === 0 ? (
        <Paper p="xl" radius="md" withBorder>
          <Stack align="center" gap="md">
            <ThemeIcon size={80} radius="xl" color="blue" variant="light">
              <IconBuilding size="3rem" />
            </ThemeIcon>
            <Title order={3} ta="center">
              No organizations yet
            </Title>
            <Text c="dimmed" ta="center" size="lg">
              Create your first organization to get started
            </Text>
            <Button
              leftSection={<IconPlus size="1rem" />}
              size="lg"
              onClick={() => setCreateModalOpened(true)}
            >
              Create Organization
            </Button>
          </Stack>
        </Paper>
      ) : (
        <Grid>
          {organizations.map((org) => (
            <Grid.Col key={org.id} span={{ base: 12, md: 6, lg: 4 }}>
              <Card withBorder radius="md" p="md" style={{ height: '100%' }}>
                <Card.Section withBorder inheritPadding py="xs">
                  <Group justify="space-between">
                    <Group gap="xs">
                      <Avatar color="blue" radius="sm">
                        {org.name.charAt(0).toUpperCase()}
                      </Avatar>
                      <div>
                        <Text fw={500} size="sm">
                          {org.name}
                        </Text>
                        <Text c="dimmed" size="xs">
                          {org.org_type}
                        </Text>
                      </div>
                    </Group>
                    {isOwner(org) && (
                      <ThemeIcon size="sm" color="yellow" variant="light">
                        <IconCrown size="0.8rem" />
                      </ThemeIcon>
                    )}
                  </Group>
                </Card.Section>

                <Stack gap="xs" mt="md" style={{ flex: 1 }}>
                  <Text size="sm" c="dimmed" lineClamp={3}>
                    {org.description || 'No description provided'}
                  </Text>

                  <Group gap="xs">
                    <Badge
                      color={getUserRole(org.id) === 'admin' ? 'green' : 'blue'}
                      variant="light"
                      size="sm"
                    >
                      {getUserRole(org.id)}
                    </Badge>
                    <Badge variant="outline" size="sm">
                      {org.members?.length || 0} members
                    </Badge>
                  </Group>

                  <Text size="xs" c="dimmed">
                    Created {new Date(org.created_at).toLocaleDateString()}
                  </Text>
                </Stack>

                <Group justify="space-between" mt="md">
                  <Button
                    variant="light"
                    size="sm"
                    leftSection={<IconEye size="1rem" />}
                    onClick={() => navigate(`/organizations/${org.id}`)}
                  >
                    View Details
                  </Button>

                  <Group gap="xs">
                    {isAdmin(org.id) && (
                      <ActionIcon
                        variant="light"
                        color="blue"
                        onClick={() => navigate(`/organizations/${org.id}?tab=settings`)}
                      >
                        <IconEdit size="1rem" />
                      </ActionIcon>
                    )}
                    {isOwner(org) && (
                      <ActionIcon
                        variant="light"
                        color="red"
                        onClick={() => handleDeleteOrganization(org)}
                      >
                        <IconTrash size="1rem" />
                      </ActionIcon>
                    )}
                  </Group>
                </Group>
              </Card>
            </Grid.Col>
          ))}
        </Grid>
      )}

      <Modal
        opened={createModalOpened}
        onClose={() => setCreateModalOpened(false)}
        title="Create New Organization"
        size="md"
      >
        <form onSubmit={form.onSubmit(handleCreateOrganization)}>
          <Stack gap="md">
            <TextInput
              label="Organization Name"
              placeholder="Enter organization name"
              {...form.getInputProps('name')}
              required
            />

            <Select
              label="Organization Type"
              placeholder="Select type"
              data={[
                { value: 'organization', label: 'Organization' },
                { value: 'domain', label: 'Domain' },
                { value: 'tenant', label: 'Tenant' },
              ]}
              {...form.getInputProps('org_type')}
              required
            />

            <Textarea
              label="Description"
              placeholder="Describe your organization"
              {...form.getInputProps('description')}
              minRows={3}
              required
            />

            <Group justify="flex-end" mt="md">
              <Button
                variant="subtle"
                onClick={() => setCreateModalOpened(false)}
                disabled={submitting}
              >
                Cancel
              </Button>
              <Button type="submit" loading={submitting}>
                Create Organization
              </Button>
            </Group>
          </Stack>
        </form>
      </Modal>
    </Container>
  );
}
EOF

# 12. Create basic page files
cat > src/pages/ProfilePage.tsx << 'EOF'
import React from 'react';
import {
  Container,
  Title,
  Paper,
  Stack,
  TextInput,
  Button,
  Group,
  Avatar,
  Text,
  Select,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { useAuth } from '../hooks/useAuth';

export function ProfilePage() {
  const { user, refreshUser } = useAuth();

  const form = useForm({
    initialValues: {
      firstName: user?.first_name || '',
      lastName: user?.last_name || '',
      email: user?.email || '',
      timeZone: user?.time_zone || 'UTC',
      uiMode: user?.ui_mode || 'system',
    },
  });

  const handleSubmit = async (values: any) => {
    console.log('Update profile:', values);
  };

  return (
    <Container size="md" py="xl">
      <Title order={1} mb="xl">Profile Settings</Title>
      
      <Paper withBorder p="xl" radius="md">
        <Group mb="xl">
          <Avatar size={80} color="blue">
            {user?.first_name?.charAt(0)}{user?.last_name?.charAt(0)}
          </Avatar>
          <div>
            <Text size="lg" fw={500}>
              {user?.first_name} {user?.last_name}
            </Text>
            <Text c="dimmed">{user?.email}</Text>
          </div>
        </Group>

        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack gap="md">
            <Group grow>
              <TextInput
                label="First Name"
                {...form.getInputProps('firstName')}
              />
              <TextInput
                label="Last Name"
                {...form.getInputProps('lastName')}
              />
            </Group>

            <TextInput
              label="Email"
              {...form.getInputProps('email')}
              disabled
            />

            <Select
              label="Time Zone"
              data={[
                { value: 'UTC', label: 'UTC' },
                { value: 'America/New_York', label: 'Eastern Time' },
                { value: 'America/Chicago', label: 'Central Time' },
                { value: 'America/Denver', label: 'Mountain Time' },
                { value: 'America/Los_Angeles', label: 'Pacific Time' },
              ]}
              {...form.getInputProps('timeZone')}
            />

            <Select
              label="UI Mode"
              data={[
                { value: 'system', label: 'System' },
                { value: 'light', label: 'Light' },
                { value: 'dark', label: 'Dark' },
              ]}
              {...form.getInputProps('uiMode')}
            />