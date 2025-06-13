// frontend/src/components/layout/Navigation.tsx
import React from 'react';
import { Group, Text, ThemeIcon, UnstyledButton, Badge, Stack } from '@mantine/core';
import {
  IconDashboard,
  IconUsers,
  IconBuilding,
  IconUser,
  IconSettings,
  IconChartBar,
  IconBuildingStore,
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
  color?: string;
}

export function Navigation({ onNavigate }: NavigationProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { user } = useAuth();

  const handleNavigate = (path: string) => {
    navigate(path);
    onNavigate?.();
  };

  const canCreateOrganizations = () => {
    return user?.can_create_organizations || false;
  };

  const navItems: NavItem[] = [
    {
      icon: <IconDashboard size="1.2rem" />,
      label: 'Dashboard',
      path: '/dashboard',
      color: 'blue',
    },
    {
      icon: <IconBuilding size="1.2rem" />,
      label: 'Organizations',
      path: '/organizations',
      color: 'blue',
    },
    {
      icon: <IconUsers size="1.2rem" />,
      label: 'Users',
      path: '/users',
      adminOnly: true,
      color: 'green',
    },
    {
      icon: <IconChartBar size="1.2rem" />,
      label: 'Analytics',
      path: '/analytics',
      adminOnly: true,
      color: 'orange',
    },
    {
      icon: <IconUser size="1.2rem" />,
      label: 'Profile',
      path: '/profile',
      color: 'grape',
    },
    {
      icon: <IconSettings size="1.2rem" />,
      label: 'Settings',
      path: '/settings',
      color: 'gray',
    },
  ];

  const filteredNavItems = navItems.filter(item => {
    if (item.adminOnly) {
      return canCreateOrganizations();
    }
    return true;
  });

  const isActive = (path: string) => {
    return location.pathname === path || location.pathname.startsWith(path + '/');
  };

  return (
    <Stack gap="xs">
      {/* Main Navigation */}
      <Stack gap="xs">
        <Text size="xs" fw={500} c="dimmed" tt="uppercase" px="xs">
          Navigation
        </Text>
        {filteredNavItems.map((item) => (
          <UnstyledButton
            key={item.path}
            onClick={() => handleNavigate(item.path)}
            style={{
              display: 'block',
              width: '100%',
              padding: '8px 12px',
              borderRadius: '8px',
              backgroundColor: isActive(item.path) 
                ? 'var(--mantine-color-blue-light)' 
                : 'transparent',
              color: isActive(item.path) 
                ? 'var(--mantine-color-blue-6)' 
                : 'var(--mantine-color-text)',
              border: isActive(item.path) 
                ? '1px solid var(--mantine-color-blue-3)' 
                : '1px solid transparent',
            }}
          >
            <Group gap="sm">
              <ThemeIcon
                size="sm"
                variant={isActive(item.path) ? 'filled' : 'light'}
                color={isActive(item.path) ? 'blue' : item.color}
              >
                {item.icon}
              </ThemeIcon>
              <Text size="sm" fw={isActive(item.path) ? 600 : 400}>
                {item.label}
              </Text>
              {item.badge && (
                <Badge size="xs" color="red" variant="filled">
                  {item.badge}
                </Badge>
              )}
            </Group>
          </UnstyledButton>
        ))}
      </Stack>

      {/* User Organizations Quick Access */}
      {user?.organizations && user.organizations.length > 0 && (
        <Stack gap="xs" mt="md">
          <Text size="xs" fw={500} c="dimmed" tt="uppercase" px="xs">
            Your Workspaces
          </Text>
          {user.organizations.slice(0, 5).map((org) => (
            <UnstyledButton
              key={org.org_id}
              onClick={() => handleNavigate(`/organizations/${org.org_id}`)}
              style={{
                display: 'block',
                width: '100%',
                padding: '6px 12px',
                borderRadius: '6px',
                backgroundColor: isActive(`/organizations/${org.org_id}`) 
                  ? 'var(--mantine-color-gray-1)' 
                  : 'transparent',
              }}
            >
              <Group gap="sm">
                <ThemeIcon
                  size="xs"
                  variant="light"
                  color={org.org_type === 'organization' ? 'blue' : 'green'}
                >
                  {org.org_type === 'organization' ? 
                    <IconBuilding size="0.8rem" /> : 
                    <IconBuildingStore size="0.8rem" />
                  }
                </ThemeIcon>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <Text size="xs" fw={500} truncate>
                    {org.org_name}
                  </Text>
                  {org.parent_name && (
                    <Text size="xs" c="dimmed" truncate>
                      under {org.parent_name}
                    </Text>
                  )}
                </div>
                <Badge
                  size="xs"
                  color={org.role === 'owner' ? 'red' : org.role === 'admin' ? 'orange' : 'blue'}
                  variant="dot"
                >
                  {org.role}
                </Badge>
              </Group>
            </UnstyledButton>
          ))}
          {user.organizations.length > 5 && (
            <UnstyledButton
              onClick={() => handleNavigate('/organizations')}
              style={{
                display: 'block',
                width: '100%',
                padding: '6px 12px',
                borderRadius: '6px',
                textAlign: 'center',
              }}
            >
              <Text size="xs" c="dimmed">
                +{user.organizations.length - 5} more
              </Text>
            </UnstyledButton>
          )}
        </Stack>
      )}
    </Stack>
  );
}