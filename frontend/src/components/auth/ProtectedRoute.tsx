import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { LoadingOverlay, Center } from '@mantine/core';
import { useAuth } from '../../hooks/useAuth';

interface ProtectedRouteProps {
  children: React.ReactNode;
  adminOnly?: boolean;
}

export function ProtectedRoute({ children, adminOnly = false }: ProtectedRouteProps) {
  const { isAuthenticated, loading, canCreateOrganizations } = useAuth();
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
    // Use canCreateOrganizations which includes bootstrap logic
    if (!canCreateOrganizations()) {
      return <Navigate to="/dashboard" replace />;
    }
  }

  return <>{children}</>;
}
