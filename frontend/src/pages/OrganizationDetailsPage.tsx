// frontend/src/pages/OrganizationDetailsPage.tsx
import React, { useState, useEffect } from 'react';
import {
  Container,
  Title,
  Text,
  Button,
  Paper,
  Stack,
  Group,
  Badge,
  Avatar,
  ActionIcon,
  Table,
  Menu,
  Modal,
  TextInput,
  Select,
  Tabs,
  Alert,
  Card,
  ThemeIcon,
  LoadingOverlay,
  SimpleGrid,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { notifications } from '@mantine/notifications';
import { modals } from '@mantine/modals';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import {
  IconArrowLeft,
  IconUsers,
  IconSettings,
  IconPlus,
  IconDots,
  IconEdit,
  IconTrash,
  IconUserPlus,
  IconCrown,
  IconShield,
  IconUser,
  IconBuilding,
  IconBuildingStore,
  IconInfoCircle,
  IconChevronRight,
} from '@tabler/icons-react';
import { useAuth } from '../hooks/useAuth';
import { ApiService } from '../services/api';
import { Organization, Member, InviteUserRequest, UpdateMemberRoleRequest } from '../types/organization';

export default function OrganizationDetailsPage() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [searchParams] = useSearchParams();
  const { user } = useAuth();
  const [organization, setOrganization] = useState<Organization | null>(null);
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [inviteModalOpened, setInviteModalOpened] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [tenants, setTenants] = useState<Organization[]>([]);

  const activeTab = searchParams.get('tab') || 'overview';

  const inviteForm = useForm<InviteUserRequest>({
    initialValues: {
      email: '',
      role: 'member',
    },
    validate: {
      email: (value) => (/^\S+@\S+$/.test(value) ? null : 'Invalid email'),
    },
  });

  useEffect(() => {
    if (id) {
      loadOrganization();
      loadMembers();
      loadTenants();
    }
  }, [id]);

  const loadOrganization = async () => {
    if (!id) return;
    try {
      const data = await ApiService.getOrganization(id);
      setOrganization(data);
    } catch (error) {
      console.error('Failed to load organization:', error);
      notifications.show({
        title: 'Error',
        message: 'Failed to load organization details',
        color: 'red',
      });
    }
  };

  const loadMembers = async () => {
    if (!id) return;
    try {
      const data = await ApiService.getOrganizationMembers(id);
      setMembers(data);
    } catch (error) {
      console.error('Failed to load members:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadTenants = async () => {
    if (!id || !organization || organization.org_type !== 'organization') return;
    try {
      const data = await ApiService.getOrganizationWithTenants(id);
      setTenants(data.children || []);
    } catch (error) {
      console.error('Failed to load tenants:', error);
    }
  };

  const handleInviteMember = async (values: InviteUserRequest) => {
    if (!id) return;
    try {
      setSubmitting(true);
      await ApiService.addOrganizationMember(id, values);
      notifications.show({
        title: 'Success',
        message: 'Member invited successfully',
        color: 'green',
      });
      setInviteModalOpened(false);
      inviteForm.reset();
      loadMembers();
    } catch (error: any) {
      notifications.show({
        title: 'Error',
        message: error.response?.data?.message || 'Failed to invite member',
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

  const getRoleIcon = (role: string) => {
    switch (role) {
      case 'owner': return <IconCrown size="0.9rem" />;
      case 'admin': return <IconShield size="0.9rem" />;
      default: return <IconUser size="0.9rem" />;
    }
  };

  const getRoleColor = (role: string) => {
    switch (role) {
      case 'owner': return 'red';
      case 'admin': return 'orange';
      default: return 'blue';
    }
  };

  const getOrgIcon = (orgType: string) => {
    return orgType === 'organization' ? <IconBuilding size="1rem" /> : <IconBuildingStore size="1rem" />;
  };

  const getOrgColor = (orgType: string) => {
    return orgType === 'organization' ? 'blue' : 'green';
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
          <Avatar color={getOrgColor(organization.org_type)} radius="md">
            {getOrgIcon(organization.org_type)}
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
              <Badge color={getOrgColor(organization.org_type)} variant="light">
                {organization.org_type}
              </Badge>
              <Badge color={getRoleColor(getUserRole())} variant="outline">
                {getUserRole()}
              </Badge>
              <Badge color="blue" variant="outline">
                {members.length} members
              </Badge>
              {organization.parent_name && (
                <Badge variant="outline" color="gray">
                  under {organization.parent_name}
                </Badge>
              )}
              <Text size="sm" c="dimmed">
                Created {new Date(organization.created_at).toLocaleDateString()}
              </Text>
            </Group>
          </div>
        </Group>

        {/* Tabs */}
        <Tabs value={activeTab} onChange={(value) => navigate(`/organizations/${id}?tab=${value}`)}>
          <Tabs.List>
            <Tabs.Tab value="overview" leftSection={<IconInfoCircle size="0.9rem" />}>
              Overview
            </Tabs.Tab>
            <Tabs.Tab value="members" leftSection={<IconUsers size="0.9rem" />}>
              Members ({members.length})
            </Tabs.Tab>
            {organization.org_type === 'organization' && (
              <Tabs.Tab value="tenants" leftSection={<IconBuildingStore size="0.9rem" />}>
                Tenants ({tenants.length})
              </Tabs.Tab>
            )}
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

              <SimpleGrid cols={{ base: 1, sm: 2, lg: 3 }} spacing="lg">
                <Card withBorder p="lg">
                  <Group gap="md">
                    <ThemeIcon size="xl" color="blue" variant="light">
                      <IconUsers size="1.5rem" />
                    </ThemeIcon>
                    <div>
                      <Text size="xl" fw={700}>{members.length}</Text>
                      <Text size="sm" c="dimmed">Members</Text>
                    </div>
                  </Group>
                </Card>

                {organization.org_type === 'organization' && (
                  <Card withBorder p="lg">
                    <Group gap="md">
                      <ThemeIcon size="xl" color="green" variant="light">
                        <IconBuildingStore size="1.5rem" />
                      </ThemeIcon>
                      <div>
                        <Text size="xl" fw={700}>{tenants.length}</Text>
                        <Text size="sm" c="dimmed">Tenants</Text>
                      </div>
                    </Group>
                  </Card>
                )}

                <Card withBorder p="lg">
                  <Group gap="md">
                    <ThemeIcon size="xl" color="orange" variant="light">
                      <IconSettings size="1.5rem" />
                    </ThemeIcon>
                    <div>
                      <Text size="xl" fw={700}>{getUserRole()}</Text>
                      <Text size="sm" c="dimmed">Your Role</Text>
                    </div>
                  </Group>
                </Card>
              </SimpleGrid>
            </Stack>
          </Tabs.Panel>

          {/* Members Tab */}
          <Tabs.Panel value="members" pt="lg">
            <Card withBorder>
              <Group justify="space-between" mb="md">
                <Title order={4}>Members</Title>
                {canManageMembers() && (
                  <Button
                    leftSection={<IconUserPlus size="1rem" />}
                    onClick={() => setInviteModalOpened(true)}
                  >
                    Invite Member
                  </Button>
                )}
              </Group>

              {members.length === 0 ? (
                <Text c="dimmed" ta="center" py="xl">No members found</Text>
              ) : (
                <Table>
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
                            <Avatar size="sm" color="blue" radius="xl">
                              {member.first_name?.charAt(0) || member.email.charAt(0)}
                            </Avatar>
                            <div>
                              <Text size="sm" fw={500}>
                                {member.first_name} {member.last_name}
                              </Text>
                              <Text size="xs" c="dimmed">
                                {member.email}
                              </Text>
                            </div>
                          </Group>
                        </Table.Td>
                        <Table.Td>
                          <Badge
                            leftSection={getRoleIcon(member.role)}
                            color={getRoleColor(member.role)}
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
                            {member.role !== 'owner' && member.user_id !== user?.id && (
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
                                      leftSection={<IconUser size="0.9rem" />}
                                      onClick={() => handleUpdateMemberRole(member, 'member')}
                                    >
                                      Make Member
                                    </Menu.Item>
                                  )}
                                  <Menu.Divider />
                                  <Menu.Item
                                    leftSection={<IconTrash size="0.9rem" />}
                                    color="red"
                                    onClick={() => handleRemoveMember(member)}
                                  >
                                    Remove Member
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

          {/* Tenants Tab (only for organizations) */}
          {organization.org_type === 'organization' && (
            <Tabs.Panel value="tenants" pt="lg">
              <Card withBorder>
                <Group justify="space-between" mb="md">
                  <Title order={4}>Tenants under {organization.name}</Title>
                  {canManageMembers() && (
                    <Button
                      leftSection={<IconPlus size="1rem" />}
                      onClick={() => navigate('/organizations?create=true')}
                    >
                      Create Tenant
                    </Button>
                  )}
                </Group>

                {tenants.length === 0 ? (
                  <Stack align="center" gap="md" py="xl">
                    <ThemeIcon size="xl" color="gray" variant="light">
                      <IconBuildingStore size="2rem" />
                    </ThemeIcon>
                    <div style={{ textAlign: 'center' }}>
                      <Text size="lg" fw={500}>No tenants yet</Text>
                      <Text size="sm" c="dimmed">
                        Create tenant projects under this organization
                      </Text>
                    </div>
                  </Stack>
                ) : (
                  <SimpleGrid cols={{ base: 1, sm: 2 }} spacing="md">
                    {tenants.map((tenant) => (
                      <Card key={tenant.id} withBorder p="md" style={{ cursor: 'pointer' }}
                            onClick={() => navigate(`/organizations/${tenant.id}`)}>
                        <Group justify="space-between">
                          <Group>
                            <Avatar color="green" size="sm">
                              <IconBuildingStore size="1rem" />
                            </Avatar>
                            <div>
                              <Text fw={500}>{tenant.name}</Text>
                              <Text size="xs" c="dimmed" lineClamp={1}>
                                {tenant.description}
                              </Text>
                            </div>
                          </Group>
                          <IconChevronRight size="1rem" />
                        </Group>
                      </Card>
                    ))}
                  </SimpleGrid>
                )}
              </Card>
            </Tabs.Panel>
          )}

          {/* Settings Tab (admin only) */}
          {isAdmin() && (
            <Tabs.Panel value="settings" pt="lg">
              <Card withBorder>
                <Title order={4} mb="md">Organization Settings</Title>
                <Text c="dimmed">
                  Organization settings and configuration options will be available here.
                </Text>
              </Card>
            </Tabs.Panel>
          )}
        </Tabs>
      </Stack>

      {/* Invite Member Modal */}
      <Modal
        opened={inviteModalOpened}
        onClose={() => setInviteModalOpened(false)}
        title="Invite Member"
      >
        <form onSubmit={inviteForm.onSubmit(handleInviteMember)}>
          <Stack gap="md">
            <TextInput
              label="Email"
              placeholder="Enter email address"
              {...inviteForm.getInputProps('email')}
              required
            />
            <Select
              label="Role"
              data={[
                { value: 'member', label: 'Member' },
                { value: 'admin', label: 'Admin' },
              ]}
              {...inviteForm.getInputProps('role')}
              required
            />
            <Group justify="flex-end" gap="sm">
              <Button
                variant="outline"
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