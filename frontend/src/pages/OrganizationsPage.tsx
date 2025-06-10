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
  Box,
  Divider,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { notifications } from '@mantine/notifications';
import {
  IconPlus,
  IconBuilding,
  IconEye,
  IconEdit,
  IconTrash,
  IconCrown,
  IconShield,
  IconShieldCheck,
  IconCalendar,
  IconUsers,
} from '@tabler/icons-react';
import { useNavigate } from 'react-router-dom';
import { ApiService } from '../services/api';
import { Organization, CreateOrgRequest } from '../types/organization';
import { useAuth } from '../hooks/useAuth';
import { modals } from '@mantine/modals';

export default function OrganizationsPage() {
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
    },
    validate: {
      name: (value: string) => (value.length < 2 ? 'Name must be at least 2 characters' : null),
      description: (value: string | undefined) => (!value || value.length < 10 ? 'Description must be at least 10 characters' : null),
    },
  });

  useEffect(() => {
    loadOrganizations();
  }, []);

  const loadOrganizations = async () => {
    try {
      setLoading(true);
      const data = await ApiService.getOrganizations();
      setOrganizations(data || []); // Ensure we always set an array
    } catch (error) {
      console.error('Failed to load organizations:', error);
      setOrganizations([]); // Reset to empty array on error
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
    return role === 'admin' || isOwner((organizations || []).find(o => o.id === orgId)!);
  };

  const canCreateOrganizations = () => {
    // For now, allow any authenticated user to create organizations
    // In production, you might want to restrict this to specific roles
    return !!user;
  };

  const isUserAdmin = () => {
    return canCreateOrganizations();
  };

  const getVerificationStatus = () => {
    const emailAddress = user?.verifiable_addresses?.find(addr => addr.value === user?.email);
    return emailAddress?.verified || false;
  };

  // Member View - Simple display of their organizations and account info
  const renderMemberView = () => (
    <Container size="lg" py="xl">
      <Stack gap="xl">
        {/* Account Summary */}
        <Paper withBorder radius="md" p="xl">
          <Group align="flex-start" gap="xl">
            <Avatar size={80} color="blue" radius="md">
              {user?.first_name?.charAt(0)?.toUpperCase() || user?.email?.charAt(0)?.toUpperCase()}
            </Avatar>
            <Stack gap="xs" style={{ flex: 1 }}>
              <Title order={2}>{user?.first_name} {user?.last_name}</Title>
              <Text c="dimmed" size="lg">{user?.email}</Text>
              
              <Group gap="md" mt="sm">
                <Badge
                  leftSection={getVerificationStatus() ? <IconShieldCheck size="0.8rem" /> : <IconShield size="0.8rem" />}
                  color={getVerificationStatus() ? 'green' : 'orange'}
                  variant="light"
                >
                  {getVerificationStatus() ? 'Verified' : 'Unverified'}
                </Badge>
                <Badge
                  leftSection={<IconCalendar size="0.8rem" />}
                  color="blue"
                  variant="light"
                >
                  Joined {new Date(user?.created_at || '').toLocaleDateString()}
                </Badge>
              </Group>
            </Stack>
          </Group>
        </Paper>

        {/* My Organizations */}
        <div>
          <Group justify="space-between" mb="md">
            <Title order={3}>My Organizations</Title>
            <Badge color="blue" variant="outline">
              {user?.organizations?.length || 0} organizations
            </Badge>
          </Group>

          {!user?.organizations || user.organizations.length === 0 ? (
            <Paper withBorder radius="md" p="xl" ta="center">
              <Stack align="center" gap="md">
                <ThemeIcon size={60} radius="xl" color="gray" variant="light">
                  <IconBuilding size="2rem" />
                </ThemeIcon>
                <Text c="dimmed" size="lg">You are not part of any organizations yet</Text>
                <Text c="dimmed" size="sm">Contact an administrator to be added to an organization</Text>
              </Stack>
            </Paper>
          ) : (
            <Grid>
              {user.organizations.map((orgMember) => (
                <Grid.Col key={orgMember.org_id} span={{ base: 12, md: 6 }}>
                  <Card withBorder radius="md" p="md">
                    <Group justify="space-between" mb="xs">
                      <Group gap="xs">
                        <Avatar color="blue" radius="sm" size="sm">
                          {orgMember.org_name.charAt(0).toUpperCase()}
                        </Avatar>
                        <div>
                          <Text fw={500} size="sm">{orgMember.org_name}</Text>
                          <Text c="dimmed" size="xs">{orgMember.org_type}</Text>
                        </div>
                      </Group>
                      {orgMember.role === 'admin' && (
                        <ThemeIcon size="sm" color="green" variant="light">
                          <IconCrown size="0.8rem" />
                        </ThemeIcon>
                      )}
                    </Group>

                    <Group justify="space-between" mt="md">
                      <Badge
                        color={orgMember.role === 'admin' ? 'green' : 'blue'}
                        variant="light"
                        size="sm"
                      >
                        {orgMember.role}
                      </Badge>
                      <Text size="xs" c="dimmed">
                        Joined {new Date(orgMember.joined_at).toLocaleDateString()}
                      </Text>
                    </Group>

                    <Button
                      variant="light"
                      size="sm"
                      fullWidth
                      mt="md"
                      leftSection={<IconEye size="1rem" />}
                      onClick={() => navigate(`/organizations/${orgMember.org_id}`)}
                    >
                      View Organization
                    </Button>
                  </Card>
                </Grid.Col>
              ))}
            </Grid>
          )}
        </div>
      </Stack>
    </Container>
  );

  // Admin View - Full organization management
  const renderAdminView = () => (
    <Container size="xl" py="xl">
      <Group justify="space-between" mb="xl">
        <div>
          <Title order={1}>Organization Management</Title>
          <Text c="dimmed" size="lg">
            Create and manage organizations
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
      ) : (!organizations || organizations.length === 0) ? (
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
          {(organizations || []).map((org) => (
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

      {/* Create Organization Modal */}
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

  // Show loading state
  if (loading && !user) {
    return (
      <Container size="lg" py="xl">
        <Paper p="xl" radius="md" withBorder>
          <LoadingOverlay visible={true} />
        </Paper>
      </Container>
    );
  }

  // Render appropriate view based on user role
  return isUserAdmin() ? renderAdminView() : renderMemberView();
}
