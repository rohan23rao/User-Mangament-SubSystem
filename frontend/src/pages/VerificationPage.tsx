import React, { useState, useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import {
  Container,
  Title,
  Text,
  Paper,
  Button,
  TextInput,
  Stack,
  Alert,
  LoadingOverlay,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { notifications } from '@mantine/notifications';
import { IconAlertCircle, IconCheck } from '@tabler/icons-react';
import { useAuth } from '../hooks/useAuth';

// Use the correct Kratos URL from environment
const KRATOS_PUBLIC_URL = process.env.REACT_APP_KRATOS_PUBLIC_URL || 'http://172.16.1.65:4433';

interface VerificationFlow {
  id: string;
  state?: string;
  ui: {
    action: string;
    method: string;
    nodes: Array<{
      type: string;
      group: string;
      attributes: {
        name: string;
        type: string;
        value?: string;
        required?: boolean;
        disabled?: boolean;
      };
      messages?: Array<{
        id: number;
        text: string;
        type: string;
      }>;
      meta?: {
        label?: {
          text: string;
        };
      };
    }>;
    messages?: Array<{
      id: number;
      text: string;
      type: string;
    }>;
  };
}

export default function VerificationPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();
  const [loading, setLoading] = useState(false);
  const [verificationFlow, setVerificationFlow] = useState<VerificationFlow | null>(null);
  const [hasInitialized, setHasInitialized] = useState(false);

  const form = useForm({
    initialValues: {
      code: '',
      email: '',
    },
  });

  // Initialize verification flow
  useEffect(() => {
    if (hasInitialized) return;
    
    const initializeFlow = async () => {
      const flowId = searchParams.get('flow');
      const code = searchParams.get('code');
      
      setHasInitialized(true);
      
      if (code && flowId) {
        // Auto-submit verification with code from email link
        await handleVerificationWithCode(code, flowId);
      } else if (flowId) {
        // Fetch existing flow
        await fetchVerificationFlow(flowId);
      } else {
        // Create new flow for manual verification
        await createVerificationFlow();
      }
    };

    initializeFlow();
  }, [searchParams, hasInitialized]);

  const handleVerificationWithCode = async (code: string, flowId: string) => {
    try {
      setLoading(true);
      
      // First get the flow to get CSRF token
      const flowResponse = await fetch(`${KRATOS_PUBLIC_URL}/self-service/verification/flows?id=${flowId}`, {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Accept': 'application/json',
        },
      });

      if (!flowResponse.ok) {
        throw new Error('Failed to get verification flow');
      }

      const flow = await flowResponse.json();
      
      // Submit verification with code
      await submitVerificationFlow(flow, { code });
      
    } catch (error) {
      console.error('Auto-verification error:', error);
      notifications.show({
        title: 'Verification Error',
        message: 'Failed to verify automatically. Please try entering the code manually.',
        color: 'red',
      });
      // Create a new flow for manual entry
      await createVerificationFlow();
    } finally {
      setLoading(false);
    }
  };

  const fetchVerificationFlow = async (flowId: string) => {
    try {
      setLoading(true);
      
      const response = await fetch(`${KRATOS_PUBLIC_URL}/self-service/verification/flows?id=${flowId}`, {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Accept': 'application/json',
        },
      });
      
      if (response.ok) {
        const flow = await response.json();
        setVerificationFlow(flow);
        
        // Pre-fill email if available
        const emailNode = flow.ui?.nodes?.find((node: any) => node.attributes?.name === 'email');
        if (emailNode?.attributes?.value) {
          form.setValues({ email: emailNode.attributes.value });
        }
      } else if (response.status === 410 || response.status === 403) {
        // Flow expired or CSRF error, create new flow
        notifications.show({
          title: 'Verification Link Expired',
          message: 'Creating a new verification flow.',
          color: 'yellow',
        });
        await createVerificationFlow();
      } else {
        throw new Error(`Failed to get verification flow: ${response.status}`);
      }
    } catch (error) {
      console.error('Error fetching verification flow:', error);
      await createVerificationFlow();
    } finally {
      setLoading(false);
    }
  };

  const createVerificationFlow = async () => {
    try {
      setLoading(true);
      
      const response = await fetch(`${KRATOS_PUBLIC_URL}/self-service/verification/browser`, {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Accept': 'application/json',
        },
      });
      
      if (response.ok) {
        const flow = await response.json();
        setVerificationFlow(flow);
        
        // Update URL with new flow ID
        setSearchParams({ flow: flow.id });
      } else {
        throw new Error('Failed to create verification flow');
      }
    } catch (error) {
      console.error('Error creating verification flow:', error);
      notifications.show({
        title: 'Error',
        message: 'Failed to create verification flow. Please refresh the page.',
        color: 'red',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (values: { code?: string; email?: string }) => {
    if (!verificationFlow) {
      notifications.show({
        title: 'Error',
        message: 'No verification flow available. Please refresh the page.',
        color: 'red',
      });
      return;
    }

    await submitVerificationFlow(verificationFlow, values);
  };

  const submitVerificationFlow = async (flow: VerificationFlow, values: { code?: string; email?: string }) => {
    try {
      setLoading(true);

      // Find CSRF token from flow
      const csrfTokenNode = flow.ui?.nodes?.find(
        (node: any) => node.attributes?.name === 'csrf_token'
      );

      const isCodeSubmission = values.code && values.code.trim() !== '';
      const isEmailSubmission = values.email && values.email.trim() !== '';
      
      if (!isCodeSubmission && !isEmailSubmission) {
        notifications.show({
          title: 'Input Required',
          message: 'Please enter a valid email address or verification code.',
          color: 'red',
        });
        return;
      }

      const body: any = {
        method: 'code',
      };

      // Only include the field that has data
      if (isCodeSubmission) {
        body.code = values.code;
      } else if (isEmailSubmission) {
        body.email = values.email;
      }

      // Always include CSRF token if available
      if (csrfTokenNode?.attributes?.value) {
        body.csrf_token = csrfTokenNode.attributes.value;
      }

      console.log('Submitting verification with body:', body);

      const response = await fetch(flow.ui.action, {
        method: flow.ui.method,
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify(body),
      });

      const result = await response.json();

      if (response.ok) {
        if (isCodeSubmission) {
          notifications.show({
            title: 'Email Verified',
            message: 'Your email has been successfully verified!',
            color: 'green',
          });
          // Redirect based on authentication status
          navigate(isAuthenticated ? '/dashboard' : '/login', { replace: true });
        } else {
          notifications.show({
            title: 'Verification Code Sent',
            message: 'A new verification code has been sent to your email.',
            color: 'blue',
          });
          setVerificationFlow(result);
        }
      } else {
        // Handle errors
        setVerificationFlow(result);
        
        const errorMessage = result.ui?.messages?.[0]?.text || 'Verification failed';
        notifications.show({
          title: 'Verification Failed',
          message: errorMessage,
          color: 'red',
        });
      }
    } catch (error) {
      console.error('Verification error:', error);
      notifications.show({
        title: 'Verification Error',
        message: 'An error occurred during verification. Please try again.',
        color: 'red',
      });
    } finally {
      setLoading(false);
    }
  };

  const requestNewCode = () => {
    handleSubmit({ email: form.values.email });
  };

  // Get field values and errors from flow
  const getFieldError = (fieldName: string) => {
    const node = verificationFlow?.ui?.nodes?.find(
      (node: any) => node.attributes?.name === fieldName
    );
    return node?.messages?.[0]?.text;
  };

  const hasCodeField = verificationFlow?.ui?.nodes?.some(
    (node: any) => node.attributes?.name === 'code'
  );

  const hasEmailField = verificationFlow?.ui?.nodes?.some(
    (node: any) => node.attributes?.name === 'email'
  );

  const isChooseMethodState = verificationFlow?.state === 'choose_method';
  const isSentEmailState = verificationFlow?.state === 'sent_email';

  if (!hasInitialized) {
    return (
      <Container size={420} my={40}>
        <LoadingOverlay visible={true} />
      </Container>
    );
  }

  return (
    <Container size={420} my={40}>
      <Title ta="center" order={1} mb="md">
        Verify Your Email
      </Title>
      <Text c="dimmed" size="sm" ta="center" mt={5} mb={30}>
        {isChooseMethodState && "Enter your email address to receive a verification code"}
        {isSentEmailState && "Enter the verification code sent to your email address"}
        {!isChooseMethodState && !isSentEmailState && hasCodeField && "Enter the verification code sent to your email address"}
        {!isChooseMethodState && !isSentEmailState && !hasCodeField && "Enter your email address to receive a verification code"}
      </Text>

      <Paper withBorder shadow="md" p={30} mt={30} radius="md" style={{ position: 'relative' }}>
        <LoadingOverlay visible={loading} />

        {/* Show flow messages */}
        {verificationFlow?.ui?.messages?.map((message: any, index: number) => (
          <Alert
            key={index}
            icon={message.type === 'error' ? <IconAlertCircle size="1rem" /> : <IconCheck size="1rem" />}
            color={message.type === 'error' ? 'red' : 'blue'}
            mb="md"
          >
            {message.text}
          </Alert>
        ))}

        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack>
            {hasEmailField && !hasCodeField && (
              <TextInput
                label="Email Address"
                placeholder="your@email.com"
                required
                {...form.getInputProps('email')}
                error={getFieldError('email')}
              />
            )}

            {hasCodeField && (
              <TextInput
                label="Verification Code"
                placeholder="Enter your 6-digit code"
                {...form.getInputProps('code')}
                error={getFieldError('code')}
                autoComplete="one-time-code"
                inputMode="numeric"
                pattern="[0-9]*"
              />
            )}

            <Button type="submit" fullWidth mt="xl" disabled={loading}>
              {hasCodeField ? 'Verify Email' : 'Send Verification Code'}
            </Button>

            {hasCodeField && (
              <Button 
                variant="subtle" 
                fullWidth 
                onClick={requestNewCode}
                disabled={loading}
              >
                Resend verification code
              </Button>
            )}
          </Stack>
        </form>
      </Paper>
    </Container>
  );
}