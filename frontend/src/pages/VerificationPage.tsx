import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { 
  Container, 
  Paper, 
  Title, 
  Text, 
  TextInput, 
  Button, 
  LoadingOverlay, 
  Alert,
  Stack
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { notifications } from '@mantine/notifications';
import { IconAlertCircle, IconCheck } from '@tabler/icons-react';
import { useAuth } from '../hooks/useAuth';

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
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();
  const [loading, setLoading] = useState(false);
  const [verificationFlow, setVerificationFlow] = useState<VerificationFlow | null>(null);

  const form = useForm({
    initialValues: {
      code: '',
      email: '',
    },
  });

  // Only redirect if authenticated AND no verification flow is present
  useEffect(() => {
    const flowId = searchParams.get('flow');
    if (isAuthenticated && !flowId) {
      // Only redirect if there's no verification flow to process
      navigate('/dashboard', { replace: true });
    }
  }, [isAuthenticated, navigate, searchParams]);

  // Initialize or fetch verification flow
  useEffect(() => {
    const initializeFlow = async () => {
      const flowId = searchParams.get('flow');
      
      if (flowId) {
        // Fetch existing flow
        try {
          setLoading(true);
          const response = await fetch(`http://localhost:4433/self-service/verification/flows?id=${flowId}`, {
            credentials: 'include',
          });
          
          if (response.ok) {
            const flow = await response.json();
            setVerificationFlow(flow);
            
            // Pre-fill email if available from flow
            const emailNode = flow.ui?.nodes?.find((node: any) => node.attributes?.name === 'email');
            if (emailNode?.attributes?.value) {
              form.setValues({ email: emailNode.attributes.value });
            }
          } else if (response.status === 410) {
            // Flow expired, create a new one
            notifications.show({
              title: 'Verification Link Expired',
              message: 'The verification link has expired. Creating a new verification flow.',
              color: 'orange',
            });
            createVerificationFlow();
          } else {
            throw new Error('Failed to get verification flow');
          }
        } catch (error) {
          console.error('Error fetching verification flow:', error);
          notifications.show({
            title: 'Verification Error',
            message: 'Failed to load verification. Please try again.',
            color: 'red',
          });
          createVerificationFlow();
        } finally {
          setLoading(false);
        }
      } else {
        // No flow ID, create new verification flow
        createVerificationFlow();
      }
    };

    initializeFlow();
  }, [searchParams, form]);

  const createVerificationFlow = async () => {
    try {
      setLoading(true);
      const response = await fetch('http://localhost:4433/self-service/verification/browser', {
        credentials: 'include',
        headers: {
          'Accept': 'application/json',
        },
      });
      
      if (response.ok) {
        const flow = await response.json();
        setVerificationFlow(flow);
        // Update URL with flow ID
        navigate(`/verification?flow=${flow.id}`, { replace: true });
      } else {
        throw new Error('Failed to create verification flow');
      }
    } catch (error) {
      console.error('Error creating verification flow:', error);
      notifications.show({
        title: 'Verification Error',
        message: 'Failed to create verification flow. Please try again.',
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

    try {
      setLoading(true);

      // Find CSRF token from flow
      const csrfTokenNode = verificationFlow.ui?.nodes?.find(
        (node: any) => node.attributes?.name === 'csrf_token'
      );

      // Determine the submission type based on flow state and available values
      const isCodeSubmission = values.code && values.code.trim() !== '';
      const isEmailSubmission = values.email && values.email.trim() !== '';
      
      const body: any = {
        method: 'code',
      };

      if (isCodeSubmission) {
        body.code = values.code;
      } else if (isEmailSubmission) {
        body.email = values.email;
      } else {
        // If no valid input, show error
        notifications.show({
          title: 'Input Required',
          message: 'Please enter a valid email address or verification code.',
          color: 'red',
        });
        return;
      }

      if (csrfTokenNode?.attributes?.value) {
        body.csrf_token = csrfTokenNode.attributes.value;
      }

      const response = await fetch(verificationFlow.ui.action, {
        method: verificationFlow.ui.method,
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify(body),
      });

      if (response.ok) {
        const result = await response.json();
        
        // Check if this was a successful verification
        if (isCodeSubmission) {
          notifications.show({
            title: 'Email Verified',
            message: 'Your email has been successfully verified!',
            color: 'green',
          });
          // Redirect to dashboard if already authenticated, otherwise to login
          navigate(isAuthenticated ? '/dashboard' : '/login', { replace: true });
        } else {
          // This was a request for new code
          notifications.show({
            title: 'Verification Code Sent',
            message: 'A new verification code has been sent to your email.',
            color: 'blue',
          });
          // Update the flow
          setVerificationFlow(result);
        }
      } else {
        const errorData = await response.json();
        
        // Update flow with new data (including errors)
        setVerificationFlow(errorData);
        
        const errorMessage = errorData.ui?.messages?.[0]?.text || 'Verification failed';
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
            {hasEmailField && (
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
              />
            )}

            <Button type="submit" fullWidth mt="xl" disabled={loading}>
              {isChooseMethodState ? 'Send Verification Code' : 'Verify Email'}
            </Button>

            {isSentEmailState && (
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