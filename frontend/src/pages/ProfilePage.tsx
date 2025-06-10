import React from 'react';
import {
  Container,
  Title,
  Paper,
  Stack,
  TextInput,
  Button,
  Group,
  Avatar,
  Text,
  Select,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { useAuth } from '../hooks/useAuth';

export default function ProfilePage() {
  const { user, refreshUser } = useAuth();

  const form = useForm({
    initialValues: {
      firstName: user?.first_name || '',
      lastName: user?.last_name || '',
      email: user?.email || '',
      timeZone: user?.time_zone || 'UTC',
      uiMode: user?.ui_mode || 'system',
    },
  });

  const handleSubmit = async (values: any) => {
    console.log('Update profile:', values);
  };

  return (
    <Container size="md" py="xl">
      <Title order={1} mb="xl">Profile Settings</Title>
      
      <Paper withBorder p="xl" radius="md">
        <Group mb="xl">
          <Avatar size={80} color="blue">
            {user?.first_name?.charAt(0)}{user?.last_name?.charAt(0)}
          </Avatar>
          <div>
            <Text size="lg" fw={500}>
              {user?.first_name} {user?.last_name}
            </Text>
            <Text c="dimmed">{user?.email}</Text>
          </div>
        </Group>

        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack gap="md">
            <Group grow>
              <TextInput
                label="First Name"
                {...form.getInputProps('firstName')}
              />
              <TextInput
                label="Last Name"
                {...form.getInputProps('lastName')}
              />
            </Group>

            <TextInput
              label="Email"
              {...form.getInputProps('email')}
              disabled
            />

            <Select
              label="Time Zone"
              data={[
                { value: 'UTC', label: 'UTC' },
                { value: 'America/New_York', label: 'Eastern Time' },
                { value: 'America/Chicago', label: 'Central Time' },
                { value: 'America/Denver', label: 'Mountain Time' },
                { value: 'America/Los_Angeles', label: 'Pacific Time' },
              ]}
              {...form.getInputProps('timeZone')}
            />

            <Select
              label="UI Mode"
              data={[
                { value: 'system', label: 'System' },
                { value: 'light', label: 'Light' },
                { value: 'dark', label: 'Dark' },
              ]}
              {...form.getInputProps('uiMode')}
            />

            <Group justify="flex-end" mt="md">
              <Button type="submit">Save Changes</Button>
            </Group>
          </Stack>
        </form>
      </Paper>
    </Container>
  );
}
