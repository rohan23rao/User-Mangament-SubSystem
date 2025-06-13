// frontend/src/pages/OrganizationsPage.tsx
import React, { useState, useEffect } from 'react';
import {
  Container,
  Title,
  Text,
  Button,
  Paper,
  Stack,
  Group,
  Grid,
  Card,
  Badge,
  Modal,
  TextInput,
  Textarea,
  Select,
  LoadingOverlay,
  Alert,
  ActionIcon,
  Avatar,
  ThemeIcon,
  Tabs,
  SimpleGrid,
  Divider,
  Box,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { notifications } from '@mantine/notifications';
import { modals } from '@mantine/modals';
import { useNavigate, useSearchParams } from 'react-router-dom';
import {
  IconPlus,
  IconEye,
  IconEdit,
  IconTrash,
  IconBuilding,
  IconBuildingStore,
  IconUsers,
  IconCalendar,
  IconHierarchy,
  IconInfoCircle,
  IconChevronRight,
} from '@tabler/icons-react';
import { useAuth } from '../hooks/useAuth';
import { ApiService } from '../services/api';
import { Organization, CreateOrgRequest } from '../types/organization';

interface GroupedOrganizations {
  organizations: Organization[];
  tenants: { [orgId: string]: Organization[] };
}

function OrganizationsPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { user } = useAuth();
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalOpened, setCreateModalOpened] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [activeTab, setActiveTab] = useState<string>('hierarchy');

  const form = useForm<CreateOrgRequest>({
    initialValues: {
      name: '',
      description: '',
      org_type: 'organization',
      parent_id: undefined,
    },
    validate: {
      name: (value: string) => (value.length < 2 ? 'Name must be at least 2 characters' : null),
      description: (value: string) => (value.length < 10 ? 'Description must be at least 10 characters' : null),
      parent_id: (value, values) => {
        if (values.org_type === 'tenant' && !value) {
          return 'Parent organization is required for tenants';
        }
        return null;
      },
    },
  });

  useEffect(() => {
    loadOrganizations();
  }, []);

  useEffect(() => {
    // Open create modal if ?create=true in URL
    if (searchParams.get('create') === 'true') {
      setCreateModalOpened(true);
      setSearchParams({});
    }
  }, [searchParams]);

  const loadOrganizations = async () => {
    try {
      setLoading(true);
      const data = await ApiService.getOrganizations();
      setOrganizations(data || []);
    } catch (error) {
      console.error('Failed to load organizations:', error);
      notifications.show({
        title: 'Error',
        message: 'Failed to load organizations',
        color: 'red',
      });
    } finally {
      setLoading(false);
    }
  };

  const groupOrganizations = (): GroupedOrganizations => {
    const orgs = organizations.filter(o => o.org_type === 'organization');
    const tenants: { [orgId: string]: Organization[] } = {};
    
    organizations
      .filter(o => o.org_type === 'tenant' && o.parent_id)
      .forEach(tenant => {
        if (!tenants[tenant.parent_id!]) {
          tenants[tenant.parent_id!] = [];
        }
        tenants[tenant.parent_id!].push(tenant);
      });

    return { organizations: orgs, tenants };
  };

  const handleCreateOrganization = async (values: CreateOrgRequest) => {
    try {
      setSubmitting(true);
      await ApiService.createOrganization(values);
      notifications.show({
        title: 'Success',
        message: `${values.org_type === 'organization' ? 'Organization' : 'Tenant'} created successfully`,
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
      title: `Delete ${org.org_type === 'organization' ? 'Organization' : 'Tenant'}`,
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
            message: `${org.org_type === 'organization' ? 'Organization' : 'Tenant'} deleted successfully`,
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
    return role === 'admin' || role === 'owner';
  };

  const canCreateOrganizations = () => {
    return user?.can_create_organizations || false;
  };

  const getAvailableParents = () => {
    return organizations
      .filter(org => org.org_type === 'organization')
      .map(org => ({
        value: org.id,
        label: org.name,
      }));
  };

  const getOrgIcon = (orgType: string, size = "1.2rem") => {
    return orgType === 'organization' ? 
      <IconBuilding size={size} /> : 
      <IconBuildingStore size={size} />;
  };

  const getOrgColor = (orgType: string) => {
    return orgType === 'organization' ? 'blue' : 'green';
  };

  const getRoleColor = (role: string) => {
    switch (role) {
      case 'owner': return 'red';
      case 'admin': return 'orange';
      default: return 'blue';
    }
  };

  const renderOrganizationCard = (org: Organization, isNested = false) => (
    <Card 
      key={org.id} 
      withBorder 
      p="lg" 
      radius="md"
      style={{
        transform: isNested ? 'scale(0.95)' : 'none',
        opacity: isNested ? 0.95 : 1,
      }}
    >
      <Group justify="space-between" mb="sm">
        <Group>
          <Avatar color={getOrgColor(org.org_type)} radius="sm" size="md">
            {getOrgIcon(org.org_type, "1.2rem")}
          </Avatar>
          <div>
            <Text fw={500} size="lg">{org.name}</Text>
            <Group gap="xs">
              <Badge color={getOrgColor(org.org_type)} variant="light" size="sm">
                {org.org_type}
              </Badge>
              <Badge 
                color={getRoleColor(getUserRole(org.id))}
                variant="light"
                size="sm"
              >
                {getUserRole(org.id)}
              </Badge>
              {org.parent_name && (
                <Badge variant="outline" size="sm" color="gray">
                  under {org.parent_name}
                </Badge>
              )}
            </Group>
          </div>
        </Group>

        <Group gap="xs">
          <ActionIcon
            variant="light"
            color="blue"
            onClick={() => navigate(`/organizations/${org.id}`)}
          >
            <IconEye size="1rem" />
          </ActionIcon>
          {isAdmin(org.id) && (
            <ActionIcon
              variant="light"
              color="orange"
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

      <Text size="sm" c="dimmed" mb="sm" lineClamp={2}>
        {org.description || 'No description provided'}
      </Text>

      <Group justify="space-between" mt="auto">
        <Group gap="lg">
          <Group gap="xs">
            <IconUsers size="0.9rem" />
            <Text size="sm">{org.member_count || 0} members</Text>
          </Group>
          <Group gap="xs">
            <IconCalendar size="0.9rem" />
            <Text size="sm">{new Date(org.created_at).toLocaleDateString()}</Text>
          </Group>
        </Group>
      </Group>
    </Card>
  );

  const renderHierarchicalView = () => {
    const grouped = groupOrganizations();
    
    return (
      <Stack gap="xl">
        {grouped.organizations.map((org) => (
          <div key={org.id}>
            {renderOrganizationCard(org)}
            
            {/* Show tenants under this organization */}
            {grouped.tenants[org.id] && grouped.tenants[org.id].length > 0 && (
              <Box style={{ marginLeft: '20px', marginTop: '15px' }}>
                <Group gap="xs" mb="sm">
                  <IconChevronRight size="1rem" color="var(--mantine-color-gray-6)" />
                  <Text size="sm" c="dimmed" fw={500}>
                    {grouped.tenants[org.id].length} tenant{grouped.tenants[org.id].length === 1 ? '' : 's'} under {org.name}
                  </Text>
                </Group>
                <SimpleGrid cols={{ base: 1, sm: 2, lg: 3 }} spacing="md">
                  {grouped.tenants[org.id].map((tenant) => (
                    renderOrganizationCard(tenant, true)
                  ))}
                </SimpleGrid>
              </Box>
            )}
          </div>
        ))}
        
        {grouped.organizations.length === 0 && (
          <Paper p="xl" radius="md" withBorder>
            <Stack align="center" gap="md">
              <ThemeIcon size="xl" color="gray" variant="light">
                <IconBuilding size="2rem" />
              </ThemeIcon>
              <div style={{ textAlign: 'center' }}>
                <Text size="lg" fw={500}>No organizations yet</Text>
                <Text size="sm" c="dimmed">
                  Create your first organization to get started
                </Text>
              </div>
              {canCreateOrganizations() && (
                <Button
                  leftSection={<IconPlus size="1rem" />}
                  onClick={() => setCreateModalOpened(true)}
                >
                  Create Organization
                </Button>
              )}
            </Stack>
          </Paper>
        )}
      </Stack>
    );
  };

  const renderGridView = () => {
    const filteredOrgs = activeTab === 'organizations' 
      ? organizations.filter(org => org.org_type === 'organization')
      : activeTab === 'tenants'
      ? organizations.filter(org => org.org_type === 'tenant')
      : organizations;

    return (
      <SimpleGrid cols={{ base: 1, sm: 2, lg: 3 }} spacing="lg">
        {filteredOrgs.map((org) => renderOrganizationCard(org))}
      </SimpleGrid>
    );
  };

  const getTabCount = (type?: string) => {
    if (type === 'organizations') {
      return organizations.filter(org => org.org_type === 'organization').length;
    }
    if (type === 'tenants') {
      return organizations.filter(org => org.org_type === 'tenant').length;
    }
    return organizations.length;
  };

  if (loading) {
    return (
      <Container size="xl" py="xl">
        <LoadingOverlay visible={true} />
      </Container>
    );
  }

  return (
    <Container size="xl" py="xl">
      <Stack gap="lg">
        {/* Header */}
        <Group justify="space-between">
          <div>
            <Title order={1}>Organizations & Tenants</Title>
            <Text c="dimmed" mt="xs">
              Manage your organizations and tenant projects
            </Text>
          </div>
          {canCreateOrganizations() && (
            <Button
              leftSection={<IconPlus size="1rem" />}
              onClick={() => setCreateModalOpened(true)}
            >
              Create New
            </Button>
          )}
        </Group>

        {/* Info Alert */}
        <Alert color="blue" icon={<IconInfoCircle size="1rem" />}>
          <Text size="sm">
            <strong>Organizations</strong> are top-level workspaces. <strong>Tenants</strong> are projects within organizations.
            You can switch between them using the workspace selector in the header.
          </Text>
        </Alert>

        {/* Tabs */}
        <Tabs value={activeTab} onChange={(value) => setActiveTab(value || 'hierarchy')}>
          <Tabs.List>
            <Tabs.Tab value="hierarchy" leftSection={<IconHierarchy size="0.9rem" />}>
              Hierarchy View
            </Tabs.Tab>
            <Tabs.Tab value="all" leftSection={<IconBuilding size="0.9rem" />}>
              All ({getTabCount()})
            </Tabs.Tab>
            <Tabs.Tab value="organizations" leftSection={<IconBuilding size="0.9rem" />}>
              Organizations ({getTabCount('organizations')})
            </Tabs.Tab>
            <Tabs.Tab value="tenants" leftSection={<IconBuildingStore size="0.9rem" />}>
              Tenants ({getTabCount('tenants')})
            </Tabs.Tab>
          </Tabs.List>

          <Tabs.Panel value="hierarchy" pt="lg">
            {renderHierarchicalView()}
          </Tabs.Panel>

          <Tabs.Panel value="all" pt="lg">
            {renderGridView()}
          </Tabs.Panel>

          <Tabs.Panel value="organizations" pt="lg">
            {renderGridView()}
          </Tabs.Panel>

          <Tabs.Panel value="tenants" pt="lg">
            {renderGridView()}
          </Tabs.Panel>
        </Tabs>

        {/* Create Organization/Tenant Modal */}
        <Modal
          opened={createModalOpened}
          onClose={() => setCreateModalOpened(false)}
          title={
            <Group gap="xs">
              <ThemeIcon size="sm" variant="light" color="blue">
                <IconPlus size="0.8rem" />
              </ThemeIcon>
              <Text fw={500}>
                Create {form.values.org_type === 'organization' ? 'Organization' : 'Tenant'}
              </Text>
            </Group>
          }
          size="md"
        >
          <form onSubmit={form.onSubmit(handleCreateOrganization)}>
            <Stack gap="md">
              <Select
                label="Type"
                placeholder="Select type"
                data={[
                  { value: 'organization', label: 'Organization - Top-level workspace' },
                  { value: 'tenant', label: 'Tenant - Project under organization' },
                ]}
                value={form.values.org_type}
                onChange={(value) => {
                  form.setFieldValue('org_type', value as 'organization' | 'tenant');
                  if (value === 'organization') {
                    form.setFieldValue('parent_id', undefined);
                  }
                }}
                required
              />

              {form.values.org_type === 'tenant' && (
                <Select
                  label="Parent Organization"
                  placeholder="Select parent organization"
                  data={getAvailableParents()}
                  {...form.getInputProps('parent_id')}
                  required
                />
              )}

              <TextInput
                label={`${form.values.org_type === 'organization' ? 'Organization' : 'Tenant'} Name`}
                placeholder={`Enter ${form.values.org_type === 'organization' ? 'organization' : 'tenant'} name`}
                {...form.getInputProps('name')}
                required
              />

              <Textarea
                label="Description"
                placeholder="Describe your organization or tenant"
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
                  Create {form.values.org_type === 'organization' ? 'Organization' : 'Tenant'}
                </Button>
              </Group>
            </Stack>
          </form>
        </Modal>
      </Stack>
    </Container>
  );
}

export default OrganizationsPage;