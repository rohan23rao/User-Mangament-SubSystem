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
  const [hasInitialized, setHasInitialized] = useState(false);

  const form = useForm({
    initialValues: {
      code: '',
      email: '',
    },
  });

  // Only redirect if authenticated AND no verification flow is present
  useEffect(() => {
    const flowId = searchParams.get('flow');
    const code = searchParams.get('code');
    
    if (isAuthenticated && !flowId && !code) {
      // Only redirect if there's no verification flow to process
      navigate('/dashboard', { replace: true });
    }
  }, [isAuthenticated, navigate, searchParams]);

  // Initialize or fetch verification flow
  useEffect(() => {
    if (hasInitialized) return; // Prevent multiple initializations
    
    const initializeFlow = async () => {
      const flowId = searchParams.get('flow');
      const code = searchParams.get('code');
      
      setHasInitialized(true);
      
      if (flowId) {
        // Fetch existing flow
        await fetchVerificationFlow(flowId);
      } else if (code) {
        // Create flow and auto-submit with code
        await createVerificationFlowWithCode(code);
      } else {
        // Create new flow for manual email entry
        await createVerificationFlow();
      }
    };

    initializeFlow();
  }, [searchParams, hasInitialized]);

  const fetchVerificationFlow = async (flowId: string, retryCount = 0) => {
    try {
      setLoading(true);
      
      const response = await fetch(`http://localhost:4433/self-service/verification/flows?id=${flowId}`, {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Accept': 'application/json',
        },
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
          message: 'Creating a new verification flow.',
          color: 'orange',
        });
        await createVerificationFlow();
      } else if (response.status === 403 && retryCount < 2) {
        // CSRF error, clear cookies and retry
        console.log('CSRF error, clearing cookies and retrying...');
        
        // Clear all cookies for this domain
        document.cookie.split(";").forEach(function(c) { 
          document.cookie = c.replace(/^ +/, "").replace(/=.*/, "=;expires=" + new Date().toUTCString() + ";path=/"); 
        });
        
        // Wait a bit and retry
        setTimeout(() => {
          fetchVerificationFlow(flowId, retryCount + 1);
        }, 1000);
        return;
      } else {
        throw new Error(`Failed to get verification flow: ${response.status}`);
      }
    } catch (error) {
      console.error('Error fetching verification flow:', error);
      // Don't show notification here, just create new flow
      await createVerificationFlow();
    } finally {
      setLoading(false);
    }
  };

  const createVerificationFlow = async () => {
    try {
      setLoading(true);
      
      const response = await fetch('http://localhost:4433/self-service/verification/browser', {
        method: 'GET',
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
        throw new Error(`Failed to create verification flow: ${response.status}`);
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

  const createVerificationFlowWithCode = async (code: string) => {
    try {
      setLoading(true);
      
      // First create a flow
      const flowResponse = await fetch('http://localhost:4433/self-service/verification/browser', {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Accept': 'application/json',
        },
      });
      
      if (!flowResponse.ok) {
        throw new Error('Failed to create verification flow');
      }
      
      const flow = await flowResponse.json();
      setVerificationFlow(flow);
      
      // Auto-submit with the code
      await submitVerificationFlow(flow, { code });
      
    } catch (error) {
      console.error('Error with verification code:', error);
      notifications.show({
        title: 'Verification Error',
        message: 'Failed to process verification code. Please try again.',
        color: 'red',
      });
      await createVerificationFlow();
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

      // Determine the submission type
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
        console.log('Submitting verification code:', values.code);
      } else if (isEmailSubmission) {
        body.email = values.email;
        console.log('Submitting email for verification:', values.email);
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
        // Check if this was a successful verification
        if (isCodeSubmission) {
          notifications.show({
            title: 'Email Verified',
            message: 'Your email has been successfully verified!',
            color: 'green',
          });
          // Redirect to login or dashboard
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

            {hasCodeField && verificationFlow?.state === 'sent_email' && (
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