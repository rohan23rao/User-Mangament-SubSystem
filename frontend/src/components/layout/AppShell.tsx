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
