echo "🔧 Fixing session cookie and proxy issues..."

# 1. Update package.json to remove proxy and use correct API URLs
cat > package.json << 'EOF'
{
  "name": "userms-frontend",
  "version": "1.0.0",
  "private": true,
  "dependencies": {
    "@emotion/react": "^11.14.0",
    "@emotion/styled": "^11.14.0",
    "@mantine/core": "^7.17.8",
    "@mantine/dates": "^7.17.8",
    "@mantine/form": "^7.17.8",
    "@mantine/hooks": "^7.17.8",
    "@mantine/modals": "^7.17.8",
    "@mantine/notifications": "^7.17.8",
    "@tabler/icons-react": "^2.47.0",
    "@types/node": "^20.19.0",
    "@types/react": "^18.3.23",
    "@types/react-dom": "^18.3.7",
    "axios": "^1.9.0",
    "react": "^18.3.1",
    "react-dom": "^18.3.1",
    "react-router-dom": "^6.30.1",
    "react-scripts": "^5.0.1",
    "typescript": "^4.9.5"
  },
  "scripts": {
    "start": "react-scripts start",
    "build": "react-scripts build",
    "test": "react-scripts test",
    "eject": "react-scripts eject"
  },
  "eslintConfig": {
    "extends": [
      "react-app",
      "react-app/jest"
    ]
  },
  "browserslist": {
    "production": [
      ">0.2%",
      "not dead",
      "not op_mini all"
    ],
    "development": [
      "last 1 chrome version",
      "last 1 firefox version",
      "last 1 safari version"
    ]
  },
  "devDependencies": {
    "@types/react-router-dom": "^5.3.3"
  }
}
EOF

# 2. Update .env to use correct backend URLs from frontend perspective
cat > .env << 'EOF'
REACT_APP_API_URL=http://localhost:3000
REACT_APP_KRATOS_PUBLIC_URL=http://localhost:4433
PORT=3001
GENERATE_SOURCEMAP=false
EOF

# 3. Fix axios configuration in services/auth.ts to ensure cookies are sent properly
cat > src/services/auth.ts << 'EOF'
import axios from 'axios';
import { KratosSession, LoginFlow, RegistrationFlow } from '../types/auth';
import { User } from '../types/user';

const KRATOS_PUBLIC_URL = process.env.REACT_APP_KRATOS_PUBLIC_URL || 'http://localhost:4433';
const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:3000';

// Configure axios to include credentials and proper headers
axios.defaults.withCredentials = true;
axios.defaults.headers.common['Content-Type'] = 'application/json';

// Create separate axios instances for different services
const kratosApi = axios.create({
  baseURL: KRATOS_PUBLIC_URL,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

const backendApi = axios.create({
  baseURL: API_URL,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

export class AuthService {
  // Create login flow
  static async createLoginFlow(): Promise<LoginFlow> {
    const response = await kratosApi.get('/self-service/login/api');
    return response.data;
  }

  // Create registration flow
  static async createRegistrationFlow(): Promise<RegistrationFlow> {
    const response = await kratosApi.get('/self-service/registration/api');
    return response.data;
  }

  // Submit login with email/password
  static async submitLogin(flowId: string, email: string, password: string): Promise<any> {
    const response = await kratosApi.post(
      `/self-service/login?flow=${flowId}`,
      {
        method: 'password',
        password_identifier: email,
        password: password,
      }
    );
    return response.data;
  }

  // Submit registration with email/password
  static async submitRegistration(
    flowId: string, 
    email: string, 
    password: string, 
    firstName: string, 
    lastName: string
  ): Promise<any> {
    const response = await kratosApi.post(
      `/self-service/registration?flow=${flowId}`,
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
      }
    );
    return response.data;
  }

  // Get Google OAuth URL
  static async getGoogleAuthUrl(flowId: string, flowType: 'login' | 'registration'): Promise<string> {
    const endpoint = flowType === 'login' ? 'login' : 'registration';
    
    const response = await kratosApi.post(
      `/self-service/${endpoint}?flow=${flowId}`,
      {
        method: 'oidc',
        provider: 'google',
      },
      {
        maxRedirects: 0,
        validateStatus: (status) => status === 302 || status === 200,
      }
    );

    // If we get a redirect, return the Location header
    if (response.status === 302) {
      return response.headers.location;
    }

    throw new Error('Failed to get Google OAuth URL');
  }

  // Get current session
  static async getSession(): Promise<KratosSession> {
    const response = await kratosApi.get('/sessions/whoami');
    return response.data;
  }

  // Get current user from backend
  static async getCurrentUser(): Promise<User> {
    const response = await backendApi.get('/api/whoami');
    return response.data;
  }

  // Logout
  static async logout(): Promise<void> {
    await backendApi.post('/auth/logout');
  }

  // Check if user is authenticated
  static async isAuthenticated(): Promise<boolean> {
    try {
      await this.getSession();
      return true;
    } catch (error) {
      return false;
    }
  }

  // Handle OAuth callback (for Google)
  static handleOAuthCallback(): void {
    const urlParams = new URLSearchParams(window.location.search);
    const error = urlParams.get('error');
    
    if (error) {
      throw new Error(`OAuth error: ${error}`);
    }
  }

  // Create logout flow and get logout URL
  static async createLogoutFlow(): Promise<string> {
    const response = await kratosApi.get('/self-service/logout/api');
    return response.data.logout_url;
  }

  // Submit logout
  static async submitLogout(logoutToken: string): Promise<void> {
    await kratosApi.get(`/self-service/logout?token=${logoutToken}`);
  }
}

// Error handling interceptors
kratosApi.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('Kratos API Error:', error.response?.data || error.message);
    return Promise.reject(error);
  }
);

backendApi.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      console.log('Backend returned 401, redirecting to login');
      window.location.href = '/login';
    }
    console.error('Backend API Error:', error.response?.data || error.message);
    return Promise.reject(error);
  }
);
EOF

# 4. Update services/api.ts to use the backend API instance correctly
cat > src/services/api.ts << 'EOF'
import axios from 'axios';
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
    if (error.response?.status === 401) {
      console.log('API returned 401, redirecting to login');
      window.location.href = '/login';
    }
    console.error('API Error:', error.response?.data || error.message);
    return Promise.reject(error);
  }
);
EOF

# 5. Create a test page to debug the session issue
cat > src/pages/DebugPage.tsx << 'EOF'
import React, { useState, useEffect } from 'react';
import {
  Container,
  Title,
  Paper,
  Button,
  Text,
  Stack,
  Code,
  Group,
} from '@mantine/core';
import { AuthService } from '../services/auth';
import { ApiService } from '../services/api';

export function DebugPage() {
  const [sessionData, setSessionData] = useState<any>(null);
  const [userdata, setUserData] = useState<any>(null);
  const [cookies, setCookies] = useState<string>('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setCookies(document.cookie);
  }, []);

  const testSession = async () => {
    setLoading(true);
    try {
      const session = await AuthService.getSession();
      setSessionData(session);
      console.log('Session:', session);
    } catch (error) {
      console.error('Session error:', error);
      setSessionData({ error: error.message });
    }
    setLoading(false);
  };

  const testUser = async () => {
    setLoading(true);
    try {
      const user = await ApiService.getCurrentUser();
      setUserData(user);
      console.log('User:', user);
    } catch (error) {
      console.error('User error:', error);
      setUserData({ error: error.message });
    }
    setLoading(false);
  };

  const goToKratosLogin = () => {
    window.location.href = 'http://localhost:4433/self-service/login/browser';
  };

  return (
    <Container size="lg" py="xl">
      <Title order={1} mb="xl">Debug Authentication</Title>
      
      <Stack gap="md">
        <Paper withBorder p="md">
          <Title order={3} mb="sm">Current Cookies</Title>
          <Code block>{cookies || 'No cookies'}</Code>
        </Paper>

        <Paper withBorder p="md">
          <Title order={3} mb="sm">Actions</Title>
          <Group>
            <Button onClick={testSession} loading={loading}>
              Test Kratos Session
            </Button>
            <Button onClick={testUser} loading={loading}>
              Test Backend User
            </Button>
            <Button onClick={goToKratosLogin} color="blue">
              Go to Kratos Login
            </Button>
          </Group>
        </Paper>

        {sessionData && (
          <Paper withBorder p="md">
            <Title order={3} mb="sm">Session Data</Title>
            <Code block>{JSON.stringify(sessionData, null, 2)}</Code>
          </Paper>
        )}

        {userdata && (
          <Paper withBorder p="md">
            <Title order={3} mb="sm">User Data</Title>
            <Code block>{JSON.stringify(userdata, null, 2)}</Code>
          </Paper>
        )}
      </Stack>
    </Container>
  );
}
EOF

# 6. Add debug route to App.tsx
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
import { DebugPage } from './pages/DebugPage';
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
      h1: { fontSize: '2rem', fontWeight: '700' },
      h2: { fontSize: '1.5rem', fontWeight: '700' },
      h3: { fontSize: '1.25rem', fontWeight: '600' },
      h4: { fontSize: '1.125rem', fontWeight: '600' },
      h5: { fontSize: '1rem', fontWeight: '600' },
      h6: { fontSize: '0.875rem', fontWeight: '600' },
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
              <Route path="/debug" element={<DebugPage />} />
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

echo "✅ Session cookie and proxy issues fixed!"
echo ""
echo "Key changes made:"
echo "- ✅ Removed problematic proxy configuration"
echo "- ✅ Fixed axios instances with proper baseURL and credentials"
echo "- ✅ Added separate Kratos and Backend API instances"
echo "- ✅ Improved error handling and logging"
echo "- ✅ Added debug page for troubleshooting"
echo ""
echo "🧪 Testing steps:"
echo "1. Rebuild: docker-compose up frontend --build -d"
echo "2. Test at: http://localhost:3001/debug"
echo "3. Try login: http://localhost:3001/login"
echo ""
echo "If still not working, we can test the session flow manually!"
EOF