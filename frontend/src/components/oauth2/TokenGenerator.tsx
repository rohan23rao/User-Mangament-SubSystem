// frontend/src/components/oauth2/TokenGenerator.tsx
import React, { useState } from 'react';
import {
  Card,
  Stack,
  Group,
  Text,
  Button,
  Select,
  TextInput,
  Code,
  Alert,
  Badge,
  CopyButton,
  Divider,
  ActionIcon,
  Tooltip,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { notifications } from '@mantine/notifications';
import {
  IconKey,
  IconCopy,
  IconCheck,
  IconClock,
  IconShield,
  IconRefresh,
  IconAlertCircle,
} from '@tabler/icons-react';
import { OAuth2Client, TokenRequest, TokenResponse } from '../../types/oauth2';
import { ApiService } from '../../services/api';

interface TokenGeneratorProps {
  clients: OAuth2Client[];
}

export function TokenGenerator({ clients }: TokenGeneratorProps) {
  const [loading, setLoading] = useState(false);
  const [token, setToken] = useState<TokenResponse | null>(null);
  const [tokenExpiry, setTokenExpiry] = useState<Date | null>(null);

  const form = useForm<TokenRequest>({
    initialValues: {
      client_id: '',
      client_secret: '',
      grant_type: 'client_credentials',
      scope: '',
    },
    validate: {
      client_id: (value: string) => (!value ? 'Client ID is required' : null),
      client_secret: (value: string) => (!value ? 'Client Secret is required' : null),
    },
  });

  const handleGenerateToken = async (values: TokenRequest) => {
    try {
      setLoading(true);
      const tokenResponse = await ApiService.generateToken(values);
      setToken(tokenResponse);
      
      // Calculate expiry time
      const expiryDate = new Date();
      expiryDate.setSeconds(expiryDate.getSeconds() + tokenResponse.expires_in);
      setTokenExpiry(expiryDate);

      notifications.show({
        title: 'Success',
        message: 'Access token generated successfully',
        color: 'green',
      });
    } catch (error: any) {
      notifications.show({
        title: 'Error',
        message: error.response?.data?.message || 'Failed to generate token',
        color: 'red',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleClientSelect = (clientId: string | null) => {
    if (clientId) {
      const selectedClient = clients.find(c => c.client_id === clientId);
      if (selectedClient) {
        form.setFieldValue('client_id', clientId);
        form.setFieldValue('scope', selectedClient.scopes);
      }
    }
  };

  const getTimeRemaining = () => {
    if (!tokenExpiry) return null;
    
    const now = new Date();
    const diff = tokenExpiry.getTime() - now.getTime();
    
    if (diff <= 0) return 'Expired';
    
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    
    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    }
    return `${minutes}m`;
  };

  const isTokenExpired = () => {
    if (!tokenExpiry) return false;
    return new Date() >= tokenExpiry;
  };

  const clientOptions = clients.map(client => ({
    value: client.client_id,
    label: `${client.name} (${client.client_id.substring(0, 12)}...)`,
  }));

  return (
    <Card withBorder p="lg">
      <Stack gap="md">
        <Group gap="xs">
          <IconKey size="1.2rem" />
          <Text fw={600} size="lg">Token Generator</Text>
        </Group>

        <Text size="sm" c="dimmed">
          Generate access tokens for API authentication using your OAuth2 clients.
        </Text>

        <form onSubmit={form.onSubmit(handleGenerateToken)}>
          <Stack gap="md">
            <Select
              label="Select Client"
              placeholder="Choose an OAuth2 client"
              data={clientOptions}
              onChange={handleClientSelect}
              searchable
              required
            />

            <TextInput
              label="Client ID"
              placeholder="Client ID will be filled automatically"
              {...form.getInputProps('client_id')}
              readOnly
              required
            />

            <TextInput
              label="Client Secret"
              placeholder="Enter the client secret"
              type="password"
              {...form.getInputProps('client_secret')}
              required
            />

            <TextInput
              label="Scope"
              placeholder="Scopes will be filled automatically"
              {...form.getInputProps('scope')}
              readOnly
            />

            <Button
              type="submit"
              loading={loading}
              leftSection={<IconKey size="1rem" />}
              disabled={!form.values.client_id || !form.values.client_secret}
            >
              Generate Token
            </Button>
          </Stack>
        </form>

        {token && (
          <>
            <Divider />
            
            <Stack gap="sm">
              <Group justify="space-between">
                <Text fw={500}>Generated Access Token</Text>
                <Group gap="xs">
                  <Badge
                    color={isTokenExpired() ? 'red' : 'green'}
                    variant="light"
                    leftSection={<IconClock size="0.8rem" />}
                  >
                    {isTokenExpired() ? 'Expired' : `Expires in ${getTimeRemaining()}`}
                  </Badge>
                  <Tooltip label="Generate new token">
                    <ActionIcon
                      variant="subtle"
                      onClick={() => form.onSubmit(handleGenerateToken)()}
                      loading={loading}
                    >
                      <IconRefresh size="1rem" />
                    </ActionIcon>
                  </Tooltip>
                </Group>
              </Group>

              {isTokenExpired() && (
                <Alert icon={<IconAlertCircle size="1rem" />} color="red">
                  This token has expired. Generate a new one to continue using the API.
                </Alert>
              )}

              <Group gap="xs">
                <Code style={{ flex: 1, fontSize: '0.75rem' }}>
                  {token.access_token}
                </Code>
                <CopyButton value={token.access_token}>
                  {({ copied, copy }) => (
                    <Button
                      size="sm"
                      color={copied ? 'teal' : 'blue'}
                      variant="outline"
                      onClick={copy}
                      leftSection={copied ? <IconCheck size="0.8rem" /> : <IconCopy size="0.8rem" />}
                    >
                      {copied ? 'Copied' : 'Copy'}
                    </Button>
                  )}
                </CopyButton>
              </Group>

              <Stack gap="xs">
                <Text size="sm" fw={500}>Token Details:</Text>
                <Group gap="md">
                  <Group gap="xs">
                    <IconShield size="0.9rem" style={{ color: 'var(--mantine-color-dimmed)' }} />
                    <Text size="sm" c="dimmed">Type: {token.token_type}</Text>
                  </Group>
                  <Group gap="xs">
                    <IconClock size="0.9rem" style={{ color: 'var(--mantine-color-dimmed)' }} />
                    <Text size="sm" c="dimmed">Expires: {tokenExpiry?.toLocaleString()}</Text>
                  </Group>
                </Group>
                <Text size="sm" c="dimmed">Scopes: {token.scope}</Text>
              </Stack>

              <Divider />

              <Stack gap="xs">
                <Text size="sm" fw={500}>Usage Example:</Text>
                <Code block>
{`curl -H "Authorization: Bearer ${token.access_token}" \\
     -H "Content-Type: application/json" \\
     https://api.yourservice.com/data`}
                </Code>
              </Stack>
            </Stack>
          </>
        )}
      </Stack>
    </Card>
  );
}