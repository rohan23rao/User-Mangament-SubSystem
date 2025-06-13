// frontend/src/components/oauth2/OAuth2ClientsList.tsx
import React, { useState, useEffect } from 'react';
import {
  Stack,
  Group,
  Text,
  Button,
  SimpleGrid,
  Alert,
  LoadingOverlay,
  Badge,
  TextInput,
  Select,
} from '@mantine/core';
import {
  IconPlus,
  IconKey,
  IconSearch,
  IconFilter,
  IconAlertCircle,
} from '@tabler/icons-react';
import { notifications } from '@mantine/notifications';
import { OAuth2Client } from '../../types/oauth2';
import { ApiService } from '../../services/api';
import { ClientCard } from './ClientCard';
import { CreateClientModal } from './CreateClientModal';
import { TokenGenerator } from './TokenGenerator';

interface OAuth2ClientsListProps {
  organizationId: string;
  organizationName: string;
  canManageClients: boolean;
}

export function OAuth2ClientsList({
  organizationId,
  organizationName,
  canManageClients,
}: OAuth2ClientsListProps) {
  const [clients, setClients] = useState<OAuth2Client[]>([]);
  const [loading, setLoading] = useState(true);
  const [createModalOpened, setCreateModalOpened] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');

  useEffect(() => {
    loadClients();
  }, [organizationId]);

  const loadClients = async () => {
    try {
      setLoading(true);
      const response = await ApiService.getOAuth2Clients();
      
      // Handle null or undefined response
      if (!response || !response.clients) {
        console.log('No clients data received, setting empty array');
        setClients([]);
        return;
      }
      
      // Filter clients by organization - handle null clients array
      const allClients = Array.isArray(response.clients) ? response.clients : [];
      const orgClients = allClients.filter((client: any) => client.org_id === organizationId);
      setClients(orgClients);
      
      console.log(`Loaded ${orgClients.length} OAuth2 clients for organization ${organizationId}`);
    } catch (error) {
      console.error('Failed to load OAuth2 clients:', error);
      
      // Set empty array on error to prevent crashes
      setClients([]);
      
      notifications.show({
        title: 'Error',
        message: 'Failed to load OAuth2 clients',
        color: 'red',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleClientUpdate = () => {
    loadClients();
  };

  const handleCreateSuccess = () => {
    loadClients();
    setCreateModalOpened(false);
  };

  // Filter clients based on search and status
  const filteredClients = clients.filter(client => {
    const matchesSearch = client.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         client.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         client.client_id.toLowerCase().includes(searchQuery.toLowerCase());
    
    const matchesStatus = statusFilter === 'all' || 
                         (statusFilter === 'active' && client.is_active) ||
                         (statusFilter === 'inactive' && !client.is_active);
    
    return matchesSearch && matchesStatus;
  });

  const activeClientsCount = clients.filter(c => c.is_active).length;
  const inactiveClientsCount = clients.filter(c => !c.is_active).length;

  return (
    <Stack gap="lg">
      {/* Header */}
      <Group justify="space-between">
        <div>
          <Group gap="xs">
            <IconKey size="1.5rem" />
            <div>
              <Text fw={600} size="lg">OAuth2 Clients</Text>
              <Text size="sm" c="dimmed">
                Manage API access for {organizationName}
              </Text>
            </div>
          </Group>
        </div>
        
        {canManageClients && (
          <Button
            leftSection={<IconPlus size="1rem" />}
            onClick={() => setCreateModalOpened(true)}
          >
            Create Client
          </Button>
        )}
      </Group>

      {/* Stats */}
      <Group gap="md">
        <Badge size="lg" color="blue" variant="light">
          {clients.length} Total
        </Badge>
        <Badge size="lg" color="green" variant="light">
          {activeClientsCount} Active
        </Badge>
        {inactiveClientsCount > 0 && (
          <Badge size="lg" color="red" variant="light">
            {inactiveClientsCount} Inactive
          </Badge>
        )}
      </Group>

      {/* Filters */}
      <Group gap="md">
        <TextInput
          placeholder="Search clients..."
          leftSection={<IconSearch size="1rem" />}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          style={{ flex: 1, maxWidth: 300 }}
        />
        <Select
          placeholder="Filter by status"
          leftSection={<IconFilter size="1rem" />}
          data={[
            { value: 'all', label: 'All Clients' },
            { value: 'active', label: 'Active Only' },
            { value: 'inactive', label: 'Inactive Only' },
          ]}
          value={statusFilter}
          onChange={(value) => setStatusFilter(value || 'all')}
          style={{ width: 150 }}
        />
      </Group>

      {/* Token Generator */}
      {clients.length > 0 && (
        <TokenGenerator clients={clients.filter(c => c.is_active)} />
      )}

      {/* Clients List */}
      <div style={{ position: 'relative' }}>
        <LoadingOverlay visible={loading} />
        
        {!loading && clients.length === 0 ? (
          <Alert
            icon={<IconAlertCircle size="1rem" />}
            title="No OAuth2 clients yet"
            color="blue"
            variant="light"
          >
            <Text size="sm">
              Create your first OAuth2 client to enable API access for this organization.
              OAuth2 clients allow secure machine-to-machine authentication for data pipelines,
              integrations, and automated systems.
            </Text>
            {canManageClients && (
              <Button
                mt="md"
                leftSection={<IconPlus size="1rem" />}
                onClick={() => setCreateModalOpened(true)}
              >
                Create First Client
              </Button>
            )}
          </Alert>
        ) : filteredClients.length === 0 && !loading ? (
          <Alert
            icon={<IconSearch size="1rem" />}
            title="No clients match your filters"
            color="gray"
            variant="light"
          >
            <Text size="sm">
              Try adjusting your search terms or filter criteria.
            </Text>
          </Alert>
        ) : (
          <SimpleGrid cols={{ base: 1, md: 2 }} spacing="md">
            {filteredClients.map((client) => (
              <ClientCard
                key={client.id}
                client={client}
                onUpdate={handleClientUpdate}
                canManage={canManageClients}
              />
            ))}
          </SimpleGrid>
        )}
      </div>

      {/* Information */}
      {clients.length > 0 && (
        <Alert icon={<IconAlertCircle size="1rem" />} color="blue" variant="light">
          <Text size="sm">
            <strong>Security Note:</strong> Client secrets are only shown once after creation.
            Store them securely and never expose them in client-side code or public repositories.
            Use environment variables or secure configuration management for production deployments.
          </Text>
        </Alert>
      )}

      {/* Create Client Modal */}
      <CreateClientModal
        opened={createModalOpened}
        onClose={() => setCreateModalOpened(false)}
        onSuccess={handleCreateSuccess}
        organizationId={organizationId}
        organizationName={organizationName}
      />
    </Stack>
  );
}