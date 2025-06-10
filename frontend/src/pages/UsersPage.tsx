import React, { useState, useEffect } from 'react';
import {
  Container,
  Title,
  Paper,
  Table,
  Badge,
  Text,
  LoadingOverlay,
  Group,
  Stack,
  Button,
  TextInput,
  ActionIcon,
  Tooltip,
} from '@mantine/core';
import { IconSearch, IconMail, IconMailCheck, IconShield, IconUser, IconCrown } from '@tabler/icons-react';
import { notifications } from '@mantine/notifications';
import { ApiService } from '../services/api';
import { User } from '../types/user';
import { useAuth } from '../hooks/useAuth';

export default function UsersPage() {
  const { user: currentUser } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    loadUsers();
  }, []);

  const loadUsers = async () => {
    try {
      setLoading(true);
      const data = await ApiService.getUsers();
      setUsers(data);
    } catch (error) {
      console.error('Error loading users:', error);
      notifications.show({
        title: 'Error',
        message: 'Failed to load users',
        color: 'red',
      });
    } finally {
      setLoading(false);
    }
  };

  const filteredUsers = users.filter(user =>
    user.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
    user.first_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    user.last_name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getVerificationStatus = (user: User) => {
    // Use the simplified verified field from backend
    return user.verified || false;
  };

  const getUserRoles = (user: User) => {
    if (!user.organizations || user.organizations.length === 0) {
      return ['No Organizations'];
    }
    return user.organizations.map(org => `${org.role} in ${org.org_name}`);
  };

  const getUserHighestRole = (user: User) => {
    if (!user.organizations || user.organizations.length === 0) {
      return 'no-orgs';
    }
    
    const roles = user.organizations.map(org => org.role);
    if (roles.includes('admin')) return 'admin';
    if (roles.includes('manager')) return 'manager';
    return 'member';
  };

  const getRoleIcon = (role: string) => {
    switch (role) {
      case 'admin': return <IconCrown size="1rem" />;
      case 'manager': return <IconShield size="1rem" />;
      case 'no-orgs': return <IconUser size="1rem" />;
      default: return <IconUser size="1rem" />;
    }
  };

  const getRoleColor = (role: string) => {
    switch (role) {
      case 'admin': return 'red';
      case 'manager': return 'orange';
      case 'no-orgs': return 'gray';
      default: return 'blue';
    }
  };

  // Check if current user can see admin features
  const canViewAdminFeatures = () => {
    return currentUser?.organizations?.some(org => org.role === 'admin') || false;
  };

  return (
    <Container size="xl" py="md">
      <Stack gap="md">
        <Group justify="space-between">
          <Title order={2}>Users Management</Title>
          {canViewAdminFeatures() && (
            <Badge color="red" variant="light">
              Admin View
            </Badge>
          )}
        </Group>

        <Paper p="md" withBorder>
          <Group justify="space-between" mb="md">
            <TextInput
              placeholder="Search users..."
              leftSection={<IconSearch size="1rem" />}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              style={{ flex: 1, maxWidth: 300 }}
            />
            <Button onClick={loadUsers} variant="light">
              Refresh
            </Button>
          </Group>

          <div style={{ position: 'relative' }}>
            <LoadingOverlay visible={loading} />
            
            <Table striped highlightOnHover>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>User</Table.Th>
                  <Table.Th>Email</Table.Th>
                  <Table.Th>Verification</Table.Th>
                  <Table.Th>Roles</Table.Th>
                  <Table.Th>Organizations</Table.Th>
                  <Table.Th>Last Login</Table.Th>
                  <Table.Th>Created</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                                 {filteredUsers.map((user) => {
                   const isVerified = getVerificationStatus(user);
                   const highestRole = getUserHighestRole(user);
                  
                  return (
                    <Table.Tr key={user.id}>
                      <Table.Td>
                        <Group gap="xs">
                          {getRoleIcon(highestRole)}
                          <div>
                            <Text fw={500}>
                              {user.first_name} {user.last_name}
                            </Text>
                            {user.id === currentUser?.id && (
                              <Badge size="xs" color="blue">You</Badge>
                            )}
                          </div>
                        </Group>
                      </Table.Td>
                      
                      <Table.Td>
                        <Group gap="xs">
                          <Text>{user.email}</Text>
                          <Tooltip label={isVerified ? 'Email verified' : 'Email not verified'}>
                            <ActionIcon size="sm" variant="subtle" color={isVerified ? 'green' : 'gray'}>
                              {isVerified ? <IconMailCheck size="1rem" /> : <IconMail size="1rem" />}
                            </ActionIcon>
                          </Tooltip>
                        </Group>
                      </Table.Td>
                      
                      <Table.Td>
                        <Badge 
                          color={isVerified ? 'green' : 'red'} 
                          variant={isVerified ? 'light' : 'outline'}
                        >
                          {isVerified ? 'Verified' : 'Unverified'}
                        </Badge>
                      </Table.Td>
                      
                      <Table.Td>
                        <Badge 
                          color={getRoleColor(highestRole)}
                          variant="light"
                        >
                          {highestRole === 'no-orgs' ? 'No Organizations' : highestRole}
                        </Badge>
                      </Table.Td>
                      
                      <Table.Td>
                        <Stack gap="xs">
                          {user.organizations && user.organizations.length > 0 ? (
                            user.organizations.map((org, index) => (
                              <Text key={index} size="sm" c="dimmed">
                                {org.role} in {org.org_name}
                              </Text>
                            ))
                          ) : (
                            <Text size="sm" c="dimmed">No organizations</Text>
                          )}
                        </Stack>
                      </Table.Td>
                      
                      <Table.Td>
                        <Text size="sm" c="dimmed">
                          {user.last_login 
                            ? new Date(user.last_login).toLocaleDateString()
                            : 'Never'
                          }
                        </Text>
                      </Table.Td>
                      
                      <Table.Td>
                        <Text size="sm" c="dimmed">
                          {new Date(user.created_at).toLocaleDateString()}
                        </Text>
                      </Table.Td>
                    </Table.Tr>
                  );
                })}
              </Table.Tbody>
            </Table>

            {filteredUsers.length === 0 && !loading && (
              <Text ta="center" c="dimmed" py="xl">
                {searchQuery ? 'No users found matching your search.' : 'No users found.'}
              </Text>
            )}
          </div>
        </Paper>

        <Paper p="md" withBorder>
          <Title order={4} mb="sm">Debug Information</Title>
          <Text size="sm" c="dimmed">
            Total Users: {users.length} | 
            Verified Users: {users.filter(u => getVerificationStatus(u)).length} | 
            Admin Users: {users.filter(u => getUserHighestRole(u) === 'admin').length}
          </Text>
        </Paper>
      </Stack>
    </Container>
  );
}
