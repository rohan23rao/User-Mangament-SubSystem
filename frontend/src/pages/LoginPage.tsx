import React, { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import {
  Paper,
  TextInput,
  PasswordInput,
  Button,
  Title,
  Text,
  Anchor,
  Container,
  Divider,
  Stack,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { IconBrandGoogle } from '@tabler/icons-react';
import { useAuth } from '../hooks/useAuth';

export default function LoginPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { login, loginWithGoogle, loading, isAuthenticated, refreshUser } = useAuth();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const form = useForm({
    initialValues: {
      email: '',
      password: '',
    },
    validate: {
      email: (value: string) => (/^\S+@\S+$/.test(value) ? null : 'Invalid email'),
      password: (value: string) => (value.length > 0 ? null : 'Password is required'),
    },
  });

  // Redirect if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      console.log('User is authenticated, redirecting...');
      setIsSubmitting(false); // Reset submitting state
      const returnTo = searchParams.get('return_to') || '/dashboard';
      navigate(returnTo, { replace: true });
    }
  }, [isAuthenticated, navigate, searchParams]);

  // Handle OAuth callback
  useEffect(() => {
    const flow = searchParams.get('flow');
    const returnTo = searchParams.get('return_to');
    
    if (flow) {
      // This is an OAuth callback, force refresh the auth state
      console.log('OAuth callback detected, refreshing auth state...');
      const refreshAuthWithRetry = async () => {
        for (let i = 0; i < 5; i++) {
          try {
            await new Promise(resolve => setTimeout(resolve, 500 * (i + 1))); // Wait longer each retry
            await refreshUser(); // Trigger the useAuth hook to refresh
            // Small delay to let state update
            await new Promise(resolve => setTimeout(resolve, 100));
            if (isAuthenticated) {
              console.log('OAuth authentication successful, redirecting...');
              navigate(returnTo || '/dashboard', { replace: true });
              return;
            }
          } catch (error) {
            console.log(`Auth check attempt ${i + 1} failed:`, error);
          }
        }
        console.log('OAuth authentication failed after retries');
      };
      refreshAuthWithRetry();
    }
  }, [searchParams, navigate, refreshUser, isAuthenticated]);

  const handleSubmit = async (values: { email: string; password: string }) => {
    try {
      setIsSubmitting(true);
      await login(values.email, values.password);
      
      // Don't navigate immediately - let the useEffect handle it when isAuthenticated becomes true
      console.log('Login submission completed, waiting for auth state update...');
    } catch (error) {
      // Error handling is done in the useAuth hook
      console.error('Login error:', error);
      setIsSubmitting(false); // Only reset if there's an error
    }
    // Don't reset isSubmitting here - let the auth state update handle the navigation
  };

  const handleGoogleLogin = async () => {
    try {
      await loginWithGoogle();
    } catch (error) {
      console.error('Google login error:', error);
    }
  };

  const goToRegistration = () => {
    const returnTo = searchParams.get('return_to');
    const registrationUrl = returnTo 
      ? `/register?return_to=${encodeURIComponent(returnTo)}`
      : '/register';
    navigate(registrationUrl);
  };

  return (
    <Container size={420} my={40}>
      <Title ta="center" order={1} mb="md">
        Welcome back!
      </Title>
      <Text c="dimmed" size="sm" ta="center" mt={5} mb={30}>
        Sign in to your account
      </Text>

      <Paper withBorder shadow="md" p={30} mt={30} radius="md">
        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack gap="md">
            <TextInput
              label="Email"
              placeholder="your@email.com"
              required
              {...form.getInputProps('email')}
            />
            
            <PasswordInput
              label="Password"
              placeholder="Your password"
              required
              {...form.getInputProps('password')}
            />

            <Button 
              type="submit" 
              fullWidth 
              mt="xl"
              loading={isSubmitting || loading}
            >
              Sign in
            </Button>
          </Stack>
        </form>

        <Divider label="Or continue with" labelPosition="center" my="lg" />

        <Button
          variant="outline"
          fullWidth
          leftSection={<IconBrandGoogle size={16} />}
          onClick={handleGoogleLogin}
          loading={loading}
        >
          Google
        </Button>

        <Text c="dimmed" size="sm" ta="center" mt={20}>
          Don't have an account?{' '}
          <Anchor component="button" type="button" onClick={goToRegistration}>
            Create account
          </Anchor>
        </Text>

        {/* Debug info in development */}
        {process.env.NODE_ENV === 'development' && (
          <Text size="xs" c="dimmed" ta="center" mt="xl">
            Debug: Check browser console for auth flow details
          </Text>
        )}
      </Paper>
    </Container>
  );
}