// frontend/src/components/oauth2/CreateClientModal.tsx
import React, { useState } from 'react';
import {
  Modal,
  TextInput,
  Textarea,
  Button,
  Stack,
  Group,
  Select,
  Alert,
  Code,
  CopyButton,
  Divider,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { notifications } from '@mantine/notifications';
import { IconCheck, IconCopy, IconKey, IconAlertCircle } from '@tabler/icons-react';
import { CreateM2MClientRequest, OAuth2Client } from '../../types/oauth2';
import { OAUTH2_SCOPES, SCOPE_DESCRIPTIONS } from '../../utils/constants';
import { ApiService } from '../../services/api';

interface CreateClientModalProps {
  opened: boolean;
  onClose: () => void;
  onSuccess: () => void;
  organizationId: string;
  organizationName: string;
}

export function CreateClientModal({
  opened,
  onClose,
  onSuccess,
  organizationId,
  organizationName,
}: CreateClientModalProps) {
  const [loading, setLoading] = useState(false);
  const [createdClient, setCreatedClient] = useState<OAuth2Client | null>(null);
  const [step, setStep] = useState<'form' | 'success'>('form');

  const form = useForm<CreateM2MClientRequest>({
    initialValues: {
      name: '',
      description: '',
      org_id: organizationId,
      scopes: OAUTH2_SCOPES.FULL_ACCESS,
    },
    validate: {
      name: (value: string) => (value.length < 3 ? 'Name must be at least 3 characters' : null),
      org_id: (value: string) => (!value ? 'Organization ID is required' : null),
    },
  });

  const handleSubmit = async (values: CreateM2MClientRequest) => {
    try {
      setLoading(true);
      const client = await ApiService.createOAuth2Client(values);
      setCreatedClient(client);
      setStep('success');
      
      notifications.show({
        title: 'Success',
        message: 'OAuth2 client created successfully',
        color: 'green',
      });
    } catch (error: any) {
      notifications.show({
        title: 'Error',
        message: error.response?.data?.message || 'Failed to create client',
        color: 'red',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (step === 'success') {
      onSuccess();
    }
    setStep('form');
    setCreatedClient(null);
    form.reset();
    onClose();
  };

  const scopeOptions = [
    {
      value: OAUTH2_SCOPES.FULL_ACCESS,
      label: 'Full Access',
      description: SCOPE_DESCRIPTIONS[OAUTH2_SCOPES.FULL_ACCESS],
    },
    {
      value: OAUTH2_SCOPES.DATA_PIPELINE,
      label: 'Data Pipeline',
      description: SCOPE_DESCRIPTIONS[OAUTH2_SCOPES.DATA_PIPELINE],
    },
    {
      value: OAUTH2_SCOPES.DATA_EXPORT,
      label: 'Data Export',
      description: SCOPE_DESCRIPTIONS[OAUTH2_SCOPES.DATA_EXPORT],
    },
    {
      value: OAUTH2_SCOPES.TELEMETRY_INGEST,
      label: 'Telemetry Ingest',
      description: SCOPE_DESCRIPTIONS[OAUTH2_SCOPES.TELEMETRY_INGEST],
    },
    {
      value: OAUTH2_SCOPES.READ_ONLY,
      label: 'Read Only',
      description: SCOPE_DESCRIPTIONS[OAUTH2_SCOPES.READ_ONLY],
    },
  ];

  return (
    <Modal
      opened={opened}
      onClose={handleClose}
      title={step === 'form' ? 'Create OAuth2 Client' : 'Client Created Successfully'}
      size="md"
      closeOnClickOutside={false}
      closeOnEscape={false}
    >
      {step === 'form' ? (
        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack gap="md">
            <Alert icon={<IconAlertCircle size="1rem" />} color="blue" variant="light">
              Creating an OAuth2 client for <strong>{organizationName}</strong>
            </Alert>

            <TextInput
              label="Client Name"
              placeholder="Enter a descriptive name for this client"
              {...form.getInputProps('name')}
              required
            />

            <Textarea
              label="Description"
              placeholder="Describe the purpose of this client (optional)"
              {...form.getInputProps('description')}
              minRows={2}
              maxRows={4}
            />

            <Select
              label="Scopes"
              description="Choose the permissions this client will have"
              data={scopeOptions.map(option => ({
                value: option.value,
                label: option.label,
              }))}
              {...form.getInputProps('scopes')}
              required
            />

            {form.values.scopes && (
              <Alert variant="light">
                <strong>Selected permissions:</strong>{' '}
                {scopeOptions.find(opt => opt.value === form.values.scopes)?.description}
              </Alert>
            )}

            <Group justify="flex-end" gap="sm">
              <Button variant="outline" onClick={handleClose} disabled={loading}>
                Cancel
              </Button>
              <Button type="submit" loading={loading} leftSection={<IconKey size="1rem" />}>
                Create Client
              </Button>
            </Group>
          </Stack>
        </form>
      ) : (
        <Stack gap="md">
          <Alert icon={<IconAlertCircle size="1rem" />} color="yellow">
            <strong>Important:</strong> Store these credentials securely. The client secret will not be shown again.
          </Alert>

          {createdClient && (
            <>
              <div>
                <strong>Client ID:</strong>
                <Group gap="xs" mt="xs">
                  <Code style={{ flex: 1 }}>{createdClient.client_id}</Code>
                  <CopyButton value={createdClient.client_id}>
                    {({ copied, copy }) => (
                      <Button
                        size="xs"
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
              </div>

              <div>
                <strong>Client Secret:</strong>
                <Group gap="xs" mt="xs">
                  <Code style={{ flex: 1 }}>{createdClient.client_secret}</Code>
                  <CopyButton value={createdClient.client_secret || ''}>
                    {({ copied, copy }) => (
                      <Button
                        size="xs"
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
              </div>

              <Divider />

              <div>
                <strong>Example Usage:</strong>
                <Code block mt="xs">
{`# Generate access token
curl -X POST http://localhost:3000/api/oauth2/token \\
  -H "Content-Type: application/json" \\
  -d '{
    "client_id": "${createdClient.client_id}",
    "client_secret": "${createdClient.client_secret}",
    "grant_type": "client_credentials"
  }'`}
                </Code>
              </div>
            </>
          )}

          <Group justify="flex-end">
            <Button onClick={handleClose}>
              Done
            </Button>
          </Group>
        </Stack>
      )}
    </Modal>
  );
}