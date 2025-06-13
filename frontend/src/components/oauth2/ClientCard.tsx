// frontend/src/components/oauth2/ClientCard.tsx
import React, { useState } from 'react';
import {
  Card,
  Group,
  Text,
  Badge,
  ActionIcon,
  Menu,
  Button,
  Stack,
  Code,
  Tooltip,
  CopyButton,
  Alert,
} from '@mantine/core';
import {
  IconDots,
  IconKey,
  IconTrash,
  IconRefresh,
  IconCopy,
  IconCheck,
  IconCalendar,
  IconShield,
  IconActivity,
} from '@tabler/icons-react';
import { notifications } from '@mantine/notifications';
import { modals } from '@mantine/modals';
import { OAuth2Client } from '../../types/oauth2';
import { ApiService } from '../../services/api';

interface ClientCardProps {
  client: OAuth2Client;
  onUpdate: () => void;
  canManage: boolean;
}

export function ClientCard({ client, onUpdate, canManage }: ClientCardProps) {
  const [loading, setLoading] = useState(false);
  const [showSecret, setShowSecret] = useState(false);

  const handleRevoke = () => {
    modals.openConfirmModal({
      title: 'Revoke OAuth2 Client',
      children: (
        <Stack>
          <Text size="sm">
            Are you sure you want to revoke <strong>{client.name}</strong>?
          </Text>
          <Alert color="red">
            This action cannot be undone. Any applications using this client will lose access.
          </Alert>
        </Stack>
      ),
      labels: { confirm: 'Revoke', cancel: 'Cancel' },
      confirmProps: { color: 'red' },
      onConfirm: async () => {
        try {
          setLoading(true);
          await ApiService.revokeOAuth2Client(client.client_id);
          notifications.show({
            title: 'Success',
            message: 'OAuth2 client revoked successfully',
            color: 'green',
          });
          onUpdate();
        } catch (error: any) {
          notifications.show({
            title: 'Error',
            message: error.response?.data?.message || 'Failed to revoke client',
            color: 'red',
          });
        } finally {
          setLoading(false);
        }
      },
    });
  };

  const handleRegenerateSecret = () => {
    modals.openConfirmModal({
      title: 'Regenerate Client Secret',
      children: (
        <Stack>
          <Text size="sm">
            Are you sure you want to regenerate the secret for <strong>{client.name}</strong>?
          </Text>
          <Alert color="orange">
            The old secret will stop working immediately. You'll need to update all applications using this client.
          </Alert>
        </Stack>
      ),
      labels: { confirm: 'Regenerate', cancel: 'Cancel' },
      confirmProps: { color: 'orange' },
      onConfirm: async () => {
        try {
          setLoading(true);
          const updatedClient = await ApiService.regenerateClientSecret(client.client_id);
          
          // Show the new secret in a modal
          modals.open({
            title: 'New Client Secret Generated',
            children: (
              <Stack>
                <Alert color="yellow">
                  Store this secret securely - it will not be shown again!
                </Alert>
                <Group>
                  <Code style={{ flex: 1 }}>{updatedClient.client_secret}</Code>
                  <CopyButton value={updatedClient.client_secret || ''}>
                    {({ copied, copy }) => (
                      <Button
                        color={copied ? 'teal' : 'blue'}
                        variant="outline"
                        onClick={copy}
                        leftSection={copied ? <IconCheck size="1rem" /> : <IconCopy size="1rem" />}
                      >
                        {copied ? 'Copied' : 'Copy'}
                      </Button>
                    )}
                  </CopyButton>
                </Group>
              </Stack>
            ),
          });
          
          notifications.show({
            title: 'Success',
            message: 'Client secret regenerated successfully',
            color: 'green',
          });
          onUpdate();
        } catch (error: any) {
          notifications.show({
            title: 'Error',
            message: error.response?.data?.message || 'Failed to regenerate secret',
            color: 'red',
          });
        } finally {
          setLoading(false);
        }
      },
    });
  };

  const getScopesBadges = (scopes: string) => {
    return scopes.split(' ').map((scope) => (
      <Badge key={scope} size="xs" variant="light">
        {scope}
      </Badge>
    ));
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <Card withBorder p="md">
      <Stack gap="sm">
        {/* Header */}
        <Group justify="space-between">
          <div style={{ flex: 1 }}>
            <Group gap="xs">
              <Text fw={500} size="sm">
                {client.name}
              </Text>
              <Badge
                size="xs"
                color={client.is_active ? 'green' : 'red'}
                variant="light"
              >
                {client.is_active ? 'Active' : 'Inactive'}
              </Badge>
            </Group>
            {client.description && (
              <Text size="xs" c="dimmed" lineClamp={1}>
                {client.description}
              </Text>
            )}
          </div>

          {canManage && (
            <Menu shadow="md" width={200}>
              <Menu.Target>
                <ActionIcon variant="subtle" loading={loading}>
                  <IconDots size="1rem" />
                </ActionIcon>
              </Menu.Target>
              <Menu.Dropdown>
                <Menu.Item
                  leftSection={<IconKey size="0.9rem" />}
                  onClick={() => setShowSecret(!showSecret)}
                >
                  {showSecret ? 'Hide' : 'Show'} Client ID
                </Menu.Item>
                <Menu.Item
                  leftSection={<IconRefresh size="0.9rem" />}
                  onClick={handleRegenerateSecret}
                >
                  Regenerate Secret
                </Menu.Item>
                <Menu.Divider />
                <Menu.Item
                  leftSection={<IconTrash size="0.9rem" />}
                  color="red"
                  onClick={handleRevoke}
                >
                  Revoke Client
                </Menu.Item>
              </Menu.Dropdown>
            </Menu>
          )}
        </Group>

        {/* Client ID */}
        {showSecret && (
          <Group gap="xs">
            <Code style={{ flex: 1 }}>
              {client.client_id}
            </Code>
            <CopyButton value={client.client_id}>
              {({ copied, copy }) => (
                <Tooltip label={copied ? 'Copied' : 'Copy client ID'}>
                  <ActionIcon
                    size="sm"
                    color={copied ? 'teal' : 'gray'}
                    variant="subtle"
                    onClick={copy}
                  >
                    {copied ? <IconCheck size="0.8rem" /> : <IconCopy size="0.8rem" />}
                  </ActionIcon>
                </Tooltip>
              )}
            </CopyButton>
          </Group>
        )}

        {/* Scopes */}
        <Group gap="xs">
          <IconShield size="0.8rem" style={{ color: 'var(--mantine-color-dimmed)' }} />
          <Group gap="xs">
            {getScopesBadges(client.scopes)}
          </Group>
        </Group>

        {/* Metadata */}
        <Group gap="md" justify="space-between">
          <Group gap="xs">
            <IconCalendar size="0.8rem" style={{ color: 'var(--mantine-color-dimmed)' }} />
            <Text size="xs" c="dimmed">
              Created {formatDate(client.created_at)}
            </Text>
          </Group>

          {client.last_used_at && (
            <Group gap="xs">
              <IconActivity size="0.8rem" style={{ color: 'var(--mantine-color-dimmed)' }} />
              <Text size="xs" c="dimmed">
                Last used {formatDate(client.last_used_at)}
              </Text>
            </Group>
          )}
        </Group>
      </Stack>
    </Card>
  );
}