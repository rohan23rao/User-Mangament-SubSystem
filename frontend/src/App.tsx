import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { MantineProvider, createTheme } from '@mantine/core';
import { useLocalStorage } from '@mantine/hooks';
import { Notifications } from '@mantine/notifications';
import { ModalsProvider } from '@mantine/modals';

import { AuthProvider } from './hooks/useAuth';
import { AppShell } from './components/layout/AppShell';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import VerificationPage from './pages/VerificationPage';
import Dashboard from './pages/Dashboard';
import UsersPage from './pages/UsersPage';
import OrganizationsPage from './pages/OrganizationsPage';
import OrganizationDetailsPage from './pages/OrganizationDetailsPage';
import ProfilePage from './pages/ProfilePage';
import DebugPage from './pages/DebugPage';
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
              <Route path="/verification" element={<VerificationPage />} />
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
                        <Route path="/users" element={
                          <ProtectedRoute adminOnly>
                            <UsersPage />
                          </ProtectedRoute>
                        } />
                        <Route path="/organizations" element={
                          <ProtectedRoute adminOnly>
                            <OrganizationsPage />
                          </ProtectedRoute>
                        } />
                        <Route path="/organizations/:id" element={
                          <ProtectedRoute adminOnly>
                            <OrganizationDetailsPage />
                          </ProtectedRoute>
                        } />
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
