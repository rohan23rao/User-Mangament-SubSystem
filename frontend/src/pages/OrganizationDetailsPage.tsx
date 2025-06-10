import React, { useState, useEffect } from 'react';
import { useParams, useSearchParams, useNavigate } from 'react-router-dom';
import {
  Container,
  Title,
  Text,
  Paper,
  Tabs,
  Group,
  Button,
  Stack,
  Badge,
  Avatar,
  ActionIcon,
  Modal,
  TextInput,
  Select,
  Table,
  LoadingOverlay,
  Alert,
  Card,
  Divider,
  Box,
  Menu,
  ThemeIcon,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { notifications } from '@mantine/notifications';
import {
  IconSettings,
  IconUsers,
  IconPlus,
  IconDots,
  IconEdit,
  IconTrash,
  IconCrown,
  IconMail,
  IconCalendar,
  IconBuilding,
  IconArrowLeft,
  IconUserPlus,
  IconShield,
} from '@tabler/icons-react';
import { ApiService } from '../services/api';
import { Organization, Member, InviteUserRequest, UpdateMemberRoleRequest } from '../types/organization';
import { useAuth } from '../hooks/useAuth';
import { modals } from '@mantine/modals';

export default function OrganizationDetailsPage() {
  const { id } = useParams<{ id: string }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  
  const [organization, setOrganization] = useState<Organization | null>(null);
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [membersLoading, setMembersLoading] = useState(false);
  const [inviteModalOpened, setInviteModalOpened] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  
  const activeTab = searchParams.get('tab') || 'overview';

  const inviteForm = useForm<InviteUserRequest>({
    initialValues: {
      email: '',
      role: 'member',
    },
    validate: {
      email: (value) => (/^\S+@\S+$/.test(value) ? null : 'Invalid email'),
      role: (value) => (['admin', 'member'].includes(value) ? null : 'Invalid role'),
    },
  });

  useEffect(() => {
    if (id) {
      loadOrganization();
      loadMembers();
    }
  }, [id]);

  const loadOrganization = async () => {
    if (!id) return;
    
    try {
      setLoading(true);
      const data = await ApiService.getOrganization(id);
      setOrganization(data);
    } catch (error: any) {
      console.error('Failed to load organization:', error);
      notifications.show({
        title: 'Error',
        message: 'Failed to load organization details',
        color: 'red',
      });
      navigate('/organizations');
    } finally {
      setLoading(false);
    }
  };

  const loadMembers = async () => {
    if (!id) return;
    
    try {
      setMembersLoading(true);
      const data = await ApiService.getOrganizationMembers(id);
      setMembers(data || []);
    } catch (error: any) {
      console.error('Failed to load members:', error);
      notifications.show({
        title: 'Error',
        message: 'Failed to load organization members',
        color: 'red',
      });
    } finally {
      setMembersLoading(false);
    }
  };

  const handleInviteUser = async (values: InviteUserRequest) => {
    if (!id) return;
    
    try {
      setSubmitting(true);
      await ApiService.addOrganizationMember(id, values);
      notifications.show({
        title: 'Success',
        message: `User ${values.email} has been invited to the organization`,
        color: 'green',
      });
      setInviteModalOpened(false);
      inviteForm.reset();
      loadMembers();
    } catch (error: any) {
      notifications.show({
        title: 'Error',
        message: error.response?.data?.message || 'Failed to invite user',
        color: 'red',
      });
    } finally {
      setSubmitting(false);
    }
  };

  const handleRemoveMember = (member: Member) => {
    if (!id) return;
    
    modals.openConfirmModal({
      title: 'Remove Member',
      children: (
        <Text size="sm">
          Are you sure you want to remove <strong>{member.email}</strong> from this organization?
        </Text>
      ),
      labels: { confirm: 'Remove', cancel: 'Cancel' },
      confirmProps: { color: 'red' },
      onConfirm: async () => {
        try {
          await ApiService.removeOrganizationMember(id, member.user_id);
          notifications.show({
            title: 'Success',
            message: 'Member removed successfully',
            color: 'green',
          });
          loadMembers();
        } catch (error: any) {
          notifications.show({
            title: 'Error',
            message: error.response?.data?.message || 'Failed to remove member',
            color: 'red',
          });
        }
      },
    });
  };

  const handleUpdateMemberRole = async (member: Member, newRole: 'admin' | 'member') => {
    if (!id) return;
    
    try {
      await ApiService.updateMemberRole(id, member.user_id, { role: newRole });
      notifications.show({
        title: 'Success',
        message: `${member.email} role updated to ${newRole}`,
        color: 'green',
      });
      loadMembers();
    } catch (error: any) {
      notifications.show({
        title: 'Error',
        message: error.response?.data?.message || 'Failed to update member role',
        color: 'red',
      });
    }
  };

  const getUserRole = () => {
    const userOrg = user?.organizations?.find(org => org.org_id === id);
    return userOrg?.role || 'member';
  };

  const isOwner = () => {
    return organization?.owner_id === user?.id;
  };

  const isAdmin = () => {
    return getUserRole() === 'admin' || isOwner();
  };

  const canManageMembers = () => {
    return isAdmin();
  };

  if (loading) {
    return (
      <Container size="xl" py="xl">
        <Paper p="xl" radius="md" withBorder>
          <LoadingOverlay visible={true} />
        </Paper>
      </Container>
    );
  }

  if (!organization) {
    return (
      <Container size="xl" py="xl">
        <Alert color="red" title="Organization not found">
          The organization you're looking for doesn't exist or you don't have access to it.
        </Alert>
      </Container>
    );
  }

  return (
    <Container size="xl" py="xl">
      <Stack gap="lg">
        {/* Header */}
        <Group>
          <ActionIcon
            variant="subtle"
            onClick={() => navigate('/organizations')}
          >
            <IconArrowLeft size="1.2rem" />
          </ActionIcon>
          <Avatar color="blue" radius="md">
            {organization.name.charAt(0).toUpperCase()}
          </Avatar>
          <div style={{ flex: 1 }}>
            <Group gap="xs">
              <Title order={2}>{organization.name}</Title>
              {isOwner() && (
                <ThemeIcon size="sm" color="yellow" variant="light">
                  <IconCrown size="0.8rem" />
                </ThemeIcon>
              )}
            </Group>
            <Group gap="md" mt="xs">
              <Badge variant="light">{organization.org_type}</Badge>
              <Badge color="blue" variant="outline">
                {members.length} members
              </Badge>
              <Text size="sm" c="dimmed">
                Created {new Date(organization.created_at).toLocaleDateString()}
              </Text>
            </Group>
          </div>
        </Group>

        {/* Tabs */}
        <Tabs value={activeTab} onChange={(value) => navigate(`/organizations/${id}?tab=${value}`)}>
          <Tabs.List>
            <Tabs.Tab value="overview" leftSection={<IconBuilding size="0.9rem" />}>
              Overview
            </Tabs.Tab>
            <Tabs.Tab value="members" leftSection={<IconUsers size="0.9rem" />}>
              Members ({members.length})
            </Tabs.Tab>
            {isAdmin() && (
              <Tabs.Tab value="settings" leftSection={<IconSettings size="0.9rem" />}>
                Settings
              </Tabs.Tab>
            )}
          </Tabs.List>

          {/* Overview Tab */}
          <Tabs.Panel value="overview" pt="lg">
            <Stack gap="lg">
              <Card withBorder p="lg">
                <Title order={4} mb="md">Description</Title>
                <Text c="dimmed">
                  {organization.description || 'No description provided for this organization.'}
                </Text>
              </Card>

              <Card withBorder p="lg">
                <Title order={4} mb="md">Organization Details</Title>
                <Stack gap="sm">
                  <Group>
                    <Text fw={500} w={120}>Type:</Text>
                    <Badge variant="light">{organization.org_type}</Badge>
                  </Group>
                  <Group>
                    <Text fw={500} w={120}>Members:</Text>
                    <Text>{members.length} members</Text>
                  </Group>
                  <Group>
                    <Text fw={500} w={120}>Created:</Text>
                    <Text>{new Date(organization.created_at).toLocaleDateString()}</Text>
                  </Group>
                  <Group>
                    <Text fw={500} w={120}>Your Role:</Text>
                    <Badge color={getUserRole() === 'admin' ? 'green' : 'blue'}>
                      {getUserRole()}
                    </Badge>
                  </Group>
                </Stack>
              </Card>
            </Stack>
          </Tabs.Panel>

          {/* Members Tab */}
          <Tabs.Panel value="members" pt="lg">
            <Card withBorder>
              <Group justify="space-between" mb="lg">
                <Title order={4}>Organization Members</Title>
                {canManageMembers() && (
                  <Button
                    leftSection={<IconUserPlus size="1rem" />}
                    onClick={() => setInviteModalOpened(true)}
                  >
                    Invite User
                  </Button>
                )}
              </Group>

              {membersLoading ? (
                <LoadingOverlay visible={true} />
              ) : members.length === 0 ? (
                <Text c="dimmed" ta="center" py="xl">
                  No members found in this organization.
                </Text>
              ) : (
                <Table striped highlightOnHover>
                  <Table.Thead>
                    <Table.Tr>
                      <Table.Th>Member</Table.Th>
                      <Table.Th>Role</Table.Th>
                      <Table.Th>Joined</Table.Th>
                      {canManageMembers() && <Table.Th>Actions</Table.Th>}
                    </Table.Tr>
                  </Table.Thead>
                  <Table.Tbody>
                    {members.map((member) => (
                      <Table.Tr key={member.user_id}>
                        <Table.Td>
                          <Group gap="sm">
                            <Avatar size="sm" color="blue">
                              {member.email.charAt(0).toUpperCase()}
                            </Avatar>
                            <div>
                              <Text fw={500} size="sm">
                                {member.first_name && member.last_name 
                                  ? `${member.first_name} ${member.last_name}`
                                  : member.email
                                }
                              </Text>
                              <Text c="dimmed" size="xs">{member.email}</Text>
                            </div>
                          </Group>
                        </Table.Td>
                        <Table.Td>
                          <Badge 
                            color={member.role === 'admin' ? 'green' : 'blue'}
                            variant="light"
                          >
                            {member.role}
                          </Badge>
                        </Table.Td>
                        <Table.Td>
                          <Text size="sm">
                            {new Date(member.joined_at).toLocaleDateString()}
                          </Text>
                        </Table.Td>
                        {canManageMembers() && (
                          <Table.Td>
                            {member.user_id !== user?.id && (
                              <Menu shadow="md" width={200}>
                                <Menu.Target>
                                  <ActionIcon variant="subtle">
                                    <IconDots size="1rem" />
                                  </ActionIcon>
                                </Menu.Target>
                                <Menu.Dropdown>
                                  <Menu.Label>Change Role</Menu.Label>
                                  {member.role !== 'admin' && (
                                    <Menu.Item
                                      leftSection={<IconShield size="0.9rem" />}
                                      onClick={() => handleUpdateMemberRole(member, 'admin')}
                                    >
                                      Make Admin
                                    </Menu.Item>
                                  )}
                                  {member.role !== 'member' && (
                                    <Menu.Item
                                      leftSection={<IconUsers size="0.9rem" />}
                                      onClick={() => handleUpdateMemberRole(member, 'member')}
                                    >
                                      Make Member
                                    </Menu.Item>
                                  )}
                                  <Menu.Divider />
                                  <Menu.Item
                                    color="red"
                                    leftSection={<IconTrash size="0.9rem" />}
                                    onClick={() => handleRemoveMember(member)}
                                  >
                                    Remove from Organization
                                  </Menu.Item>
                                </Menu.Dropdown>
                              </Menu>
                            )}
                          </Table.Td>
                        )}
                      </Table.Tr>
                    ))}
                  </Table.Tbody>
                </Table>
              )}
            </Card>
          </Tabs.Panel>

          {/* Settings Tab */}
          {isAdmin() && (
            <Tabs.Panel value="settings" pt="lg">
              <Card withBorder p="lg">
                <Title order={4} mb="md">Organization Settings</Title>
                <Text c="dimmed" mb="lg">
                  Manage organization settings and configuration.
                </Text>
                
                <Alert color="blue" mb="lg">
                  Organization settings coming soon. For now, you can manage members in the Members tab.
                </Alert>

                {isOwner() && (
                  <Group>
                    <Button
                      color="red"
                      variant="light"
                      leftSection={<IconTrash size="1rem" />}
                    >
                      Delete Organization
                    </Button>
                  </Group>
                )}
              </Card>
            </Tabs.Panel>
          )}
        </Tabs>
      </Stack>

      {/* Invite User Modal */}
      <Modal
        opened={inviteModalOpened}
        onClose={() => setInviteModalOpened(false)}
        title="Invite User to Organization"
        size="md"
      >
        <form onSubmit={inviteForm.onSubmit(handleInviteUser)}>
          <Stack gap="md">
            <TextInput
              label="Email Address"
              placeholder="user@example.com"
              {...inviteForm.getInputProps('email')}
              required
              leftSection={<IconMail size="1rem" />}
            />

            <Select
              label="Role"
              placeholder="Select role"
              data={[
                { value: 'member', label: 'Member' },
                { value: 'admin', label: 'Admin' },
              ]}
              {...inviteForm.getInputProps('role')}
              required
            />

            <Group justify="flex-end" mt="md">
              <Button
                variant="subtle"
                onClick={() => setInviteModalOpened(false)}
                disabled={submitting}
              >
                Cancel
              </Button>
              <Button type="submit" loading={submitting}>
                Send Invitation
              </Button>
            </Group>
          </Stack>
        </form>
      </Modal>
    </Container>
  );
}
