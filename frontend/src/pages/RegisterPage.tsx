import React, { useState, useEffect } from 'react';
import { useNavigate, Link, useSearchParams } from 'react-router-dom';
import {
  Container,
  Paper,
  TextInput,
  PasswordInput,
  Button,
  Title,
  Text,
  Anchor,
  Stack,
  Divider,
  Group,
  LoadingOverlay,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { IconMail, IconLock, IconUser, IconBrandGoogle } from '@tabler/icons-react';
import { useAuth } from '../hooks/useAuth';

interface RegisterFormValues {
  firstName: string;
  lastName: string;
  email: string;
  password: string;
  confirmPassword: string;
}

export default function RegisterPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { register, registerWithGoogle, loading, isAuthenticated } = useAuth();
  const [googleLoading, setGoogleLoading] = useState(false);

  useEffect(() => {
    if (isAuthenticated) {
      const returnTo = searchParams.get('return_to') || '/dashboard';
      navigate(returnTo, { replace: true });
    }
  }, [isAuthenticated, navigate, searchParams]);

  const form = useForm<RegisterFormValues>({
    initialValues: {
      firstName: '',
      lastName: '',
      email: '',
      password: '',
      confirmPassword: '',
    },
    validate: {
      firstName: (value: string) => (value.length < 2 ? 'First name must be at least 2 characters' : null),
      lastName: (value: string) => (value.length < 2 ? 'Last name must be at least 2 characters' : null),
      email: (value: string) => (/^\S+@\S+$/.test(value) ? null : 'Invalid email'),
      password: (value: string) => (value.length < 6 ? 'Password must be at least 6 characters' : null),
      confirmPassword: (value: string, values: RegisterFormValues) =>
        value !== values.password ? 'Passwords do not match' : null,
    },
  });

  const handleSubmit = async (values: RegisterFormValues) => {
    try {
      await register(values.email, values.password, values.firstName, values.lastName);
      // Don't navigate immediately - let the useEffect handle it when isAuthenticated becomes true
      console.log('Registration submission completed, waiting for auth state update...');
    } catch (error) {
      // Error is handled in the register function
    }
  };

  const handleGoogleRegister = async () => {
    try {
      setGoogleLoading(true);
      await registerWithGoogle();
    } catch (error) {
      setGoogleLoading(false);
    }
  };

  return (
    <Container size={420} my={40}>
      <Title ta="center" fw={900}>
        Create your account
      </Title>
      <Text c="dimmed" size="sm" ta="center" mt={5}>
        Already have an account?{' '}
        <Anchor 
          size="sm" 
          component={Link} 
          to={searchParams.get('return_to') ? `/login?return_to=${encodeURIComponent(searchParams.get('return_to')!)}` : '/login'}
        >
          Sign in
        </Anchor>
      </Text>

      <Paper withBorder shadow="md" p={30} mt={30} radius="md" style={{ position: 'relative' }}>
        <LoadingOverlay visible={loading} />

        <Button
          variant="light"
          fullWidth
          leftSection={<IconBrandGoogle size="1rem" />}
          onClick={handleGoogleRegister}
          loading={googleLoading}
          disabled={loading}
          color="red"
          mb="md"
        >
          Continue with Google
        </Button>

        <Divider label="Or register with email" labelPosition="center" my="lg" />

        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack>
            <Group grow>
              <TextInput
                label="First name"
                placeholder="John"
                leftSection={<IconUser size="1rem" />}
                {...form.getInputProps('firstName')}
                disabled={loading}
              />
              <TextInput
                label="Last name"
                placeholder="Doe"
                leftSection={<IconUser size="1rem" />}
                {...form.getInputProps('lastName')}
                disabled={loading}
              />
            </Group>

            <TextInput
              label="Email"
              placeholder="your@email.com"
              leftSection={<IconMail size="1rem" />}
              {...form.getInputProps('email')}
              disabled={loading}
            />

            <PasswordInput
              label="Password"
              placeholder="Your password"
              leftSection={<IconLock size="1rem" />}
              {...form.getInputProps('password')}
              disabled={loading}
            />

            <PasswordInput
              label="Confirm password"
              placeholder="Confirm your password"
              leftSection={<IconLock size="1rem" />}
              {...form.getInputProps('confirmPassword')}
              disabled={loading}
            />

            <Button type="submit" fullWidth loading={loading}>
              Create account
            </Button>
          </Stack>
        </form>

        <Text size="xs" c="dimmed" ta="center" mt="md">
          By creating an account, you agree to our{' '}
          <Anchor size="xs" href="#" onClick={(e) => e.preventDefault()}>
            Terms of Service
          </Anchor>{' '}
          and{' '}
          <Anchor size="xs" href="#" onClick={(e) => e.preventDefault()}>
            Privacy Policy
          </Anchor>
        </Text>
      </Paper>
    </Container>
  );
}
