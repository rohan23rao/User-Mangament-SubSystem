import React, { useState, useEffect } from 'react';
import {
  Container,
  Title,
  Paper,
  Button,
  Text,
  Stack,
  Code,
  Group,
} from '@mantine/core';
import { AuthService } from '../services/auth';
import { ApiService } from '../services/api';

export default function DebugPage() {
  const [sessionData, setSessionData] = useState<any>(null);
  const [userdata, setUserData] = useState<any>(null);
  const [cookies, setCookies] = useState<string>('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setCookies(document.cookie);
  }, []);

  const testSession = async () => {
    setLoading(true);
    try {
      const session = await AuthService.getSession();
      setSessionData(session);
      console.log('Session:', session);
    } catch (error) {
      console.error('Session error:', error);
      // Type-safe error handling
      const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
      setSessionData({ error: errorMessage });
    }
    setLoading(false);
  };

  const testUser = async () => {
    setLoading(true);
    try {
      const user = await ApiService.getCurrentUser();
      setUserData(user);
      console.log('User:', user);
    } catch (error) {
      console.error('User error:', error);
      // Type-safe error handling
      const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
      setUserData({ error: errorMessage });
    }
    setLoading(false);
  };

  const goToKratosLogin = () => {
    window.location.href = 'http://localhost:4433/self-service/login/browser';
  };

  return (
    <Container size="lg" py="xl">
      <Title order={1} mb="xl">Debug Authentication</Title>
      
      <Stack gap="md">
        <Paper withBorder p="md">
          <Title order={3} mb="sm">Current Cookies</Title>
          <Code block>{cookies || 'No cookies'}</Code>
        </Paper>

        <Paper withBorder p="md">
          <Title order={3} mb="sm">Actions</Title>
          <Group>
            <Button onClick={testSession} loading={loading}>
              Test Kratos Session
            </Button>
            <Button onClick={testUser} loading={loading}>
              Test Backend User
            </Button>
            <Button onClick={goToKratosLogin} color="blue">
              Go to Kratos Login
            </Button>
          </Group>
        </Paper>

        {sessionData && (
          <Paper withBorder p="md">
            <Title order={3} mb="sm">Session Data</Title>
            <Code block>{JSON.stringify(sessionData, null, 2)}</Code>
          </Paper>
        )}

        {userdata && (
          <Paper withBorder p="md">
            <Title order={3} mb="sm">User Data</Title>
            <Code block>{JSON.stringify(userdata, null, 2)}</Code>
          </Paper>
        )}
      </Stack>
    </Container>
  );
}