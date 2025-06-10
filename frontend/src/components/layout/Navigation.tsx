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
