import React from 'react';
import {
  Container,
  Title,
  Text,
  Grid,
  Card,
  Group,
  ThemeIcon,
  Progress,
  Badge,
  Stack,
  Button,
  SimpleGrid,
  Paper,
  Avatar,
  ActionIcon,
} from '@mantine/core';
import {
  IconUsers,
  IconBuilding,
  IconUserCheck,
  IconTrendingUp,
  IconPlus,
  IconEye,
  IconSettings,
} from '@tabler/icons-react';
import { useAuth } from '../hooks/useAuth';
import { useNavigate } from 'react-router-dom';

interface StatsCardProps {
  title: string;
  value: string | number;
  description: string;
  icon: React.ReactNode;
  color: string;
  progress?: number;
}

function StatsCard({ title, value, description, icon, color, progress }: StatsCardProps) {
  return (
    <Card withBorder p="xl" radius="md">
      <Group justify="space-between">
        <div>
          <Text c="dimmed" fw={700} size="xs" tt="uppercase">
            {title}
          </Text>
          <Text fw={700} size="xl">
            {value}
          </Text>
          <Text c="dimmed" size="xs">
            {description}
          </Text>
        </div>
        <ThemeIcon color={color} size={38} radius="md">
          {icon}
        </ThemeIcon>
      </Group>
      {progress !== undefined && (
        <Progress value={progress} mt="md" size="sm" color={color} />
      )}
    </Card>
  );
}

export default function Dashboard() {
  const { user } = useAuth();
  const navigate = useNavigate();

  const userOrganizations = user?.organizations || [];
  const isAdmin = userOrganizations.some(org => org.role === 'admin');

  const stats = [
    {
      title: 'Organizations',
      value: userOrganizations.length,
      description: 'Active memberships',
      icon: <IconBuilding size="1.4rem" />,
      color: 'blue',
    },
    {
      title: 'Admin Roles',
      value: userOrganizations.filter(org => org.role === 'admin').length,
      description: 'Organizations you manage',
      icon: <IconUserCheck size="1.4rem" />,
      color: 'green',
    },
    {
      title: 'Member Since',
      value: new Date(user?.created_at || '').getFullYear(),
      description: 'Account creation year',
      icon: <IconUsers size="1.4rem" />,
      color: 'violet',
    },
    {
      title: 'Last Login',
      value: user?.last_login ? 'Recent' : 'Today',
      description: 'Activity status',
      icon: <IconTrendingUp size="1.4rem" />,
      color: 'orange',
    },
  ];

  const recentOrganizations = userOrganizations.slice(0, 3);

  return (
    <Container size="xl" py="xl">
      <Group justify="space-between" mb="xl">
        <div>
          <Title order={1} mb={4}>
            Welcome back, {user?.first_name}! 👋
          </Title>
          <Text c="dimmed" size="lg">
            Here's what's happening with your organizations today
          </Text>
        </div>
        <Button
          leftSection={<IconPlus size="1rem" />}
          onClick={() => navigate('/organizations')}
        >
          Create Organization
        </Button>
      </Group>

      <SimpleGrid
        cols={{ base: 1, xs: 2, md: 4 }}
        mb="xl"
      >
        {stats.map((stat) => (
          <StatsCard key={stat.title} {...stat} />
        ))}
      </SimpleGrid>

      <Grid>
        <Grid.Col span={{ base: 12, md: 8 }}>
          <Paper withBorder p="md" radius="md" mb="md">
            <Group justify="space-between" mb="md">
              <Title order={3}>Your Organizations</Title>
              <Button
                variant="subtle"
                size="sm"
                rightSection={<IconEye size="1rem" />}
                onClick={() => navigate('/organizations')}
              >
                View All
              </Button>
            </Group>

            {recentOrganizations.length > 0 ? (
              <Stack gap="sm">
                {recentOrganizations.map((org) => (
                  <Card key={org.org_id} withBorder radius="sm" p="md">
                    <Group justify="space-between">
                      <Group>
                        <Avatar color="blue" radius="xl">
                          {org.org_name.charAt(0).toUpperCase()}
                        </Avatar>
                        <div>
                          <Text fw={500} size="sm">
                            {org.org_name}
                          </Text>
                          <Text c="dimmed" size="xs">
                            {org.org_type} • Joined {new Date(org.joined_at).toLocaleDateString()}
                          </Text>
                        </div>
                      </Group>
                      <Group gap="xs">
                        <Badge
                          color={org.role === 'admin' ? 'green' : 'blue'}
                          variant="light"
                          size="sm"
                        >
                          {org.role}
                        </Badge>
                        <ActionIcon
                          size="sm"
                          variant="subtle"
                          onClick={() => navigate(`/organizations/${org.org_id}`)}
                        >
                          <IconEye size="1rem" />
                        </ActionIcon>
                      </Group>
                    </Group>
                  </Card>
                ))}
              </Stack>
            ) : (
              <Paper p="xl" radius="md" withBorder>
                <Stack align="center" gap="sm">
                  <ThemeIcon size={60} radius="xl" color="blue" variant="light">
                    <IconBuilding size="2rem" />
                  </ThemeIcon>
                  <Text fw={500} ta="center">
                    No organizations yet
                  </Text>
                  <Text c="dimmed" size="sm" ta="center">
                    Create your first organization or ask to be invited to one
                  </Text>
                  <Button
                    leftSection={<IconPlus size="1rem" />}
                    onClick={() => navigate('/organizations')}
                  >
                    Create Organization
                  </Button>
                </Stack>
              </Paper>
            )}
          </Paper>

          <Paper withBorder p="md" radius="md">
            <Title order={3} mb="md">
              Quick Actions
            </Title>
            <SimpleGrid cols={2} spacing="sm">
              <Button
                variant="light"
                leftSection={<IconBuilding size="1rem" />}
                onClick={() => navigate('/organizations')}
                fullWidth
              >
                Manage Organizations
              </Button>
              <Button
                variant="light"
                leftSection={<IconSettings size="1rem" />}
                onClick={() => navigate('/profile')}
                fullWidth
              >
                Edit Profile
              </Button>
              {isAdmin && (
                <>
                  <Button
                    variant="light"
                    leftSection={<IconUsers size="1rem" />}
                    onClick={() => navigate('/users')}
                    fullWidth
                  >
                    Manage Users
                  </Button>
                  <Button
                    variant="light"
                    leftSection={<IconTrendingUp size="1rem" />}
                    onClick={() => navigate('/analytics')}
                    fullWidth
                  >
                    View Analytics
                  </Button>
                </>
              )}
            </SimpleGrid>
          </Paper>
        </Grid.Col>

        <Grid.Col span={{ base: 12, md: 4 }}>
          <Paper withBorder p="md" radius="md" mb="md">
            <Group mb="md">
              <Avatar size={60} color="blue">
                {user?.first_name?.charAt(0)}{user?.last_name?.charAt(0)}
              </Avatar>
              <div style={{ flex: 1 }}>
                <Text fw={500}>
                  {user?.first_name} {user?.last_name}
                </Text>
                <Text c="dimmed" size="sm">
                  {user?.email}
                </Text>
                <Badge color="green" size="sm" mt={4}>
                  Active
                </Badge>
              </div>
            </Group>
            <Stack gap="xs">
              <Group justify="space-between">
                <Text size="sm" c="dimmed">
                  Time Zone
                </Text>
                <Text size="sm">{user?.time_zone}</Text>
              </Group>
              <Group justify="space-between">
                <Text size="sm" c="dimmed">
                  UI Mode
                </Text>
                <Text size="sm">{user?.ui_mode}</Text>
              </Group>
              <Group justify="space-between">
                <Text size="sm" c="dimmed">
                  Member Since
                </Text>
                <Text size="sm">
                  {new Date(user?.created_at || '').toLocaleDateString()}
                </Text>
              </Group>
            </Stack>
            <Button
              variant="light"
              fullWidth
              mt="md"
              onClick={() => navigate('/profile')}
            >
              Edit Profile
            </Button>
          </Paper>

          <Paper withBorder p="md" radius="md">
            <Title order={4} mb="md">
              Recent Activity
            </Title>
            <Stack gap="sm">
              <Group>
                <ThemeIcon size="sm" color="blue" variant="light">
                  <IconUserCheck size="0.8rem" />
                </ThemeIcon>
                <Text size="sm">Account created</Text>
              </Group>
              {user?.last_login && (
                <Group>
                  <ThemeIcon size="sm" color="green" variant="light">
                    <IconTrendingUp size="0.8rem" />
                  </ThemeIcon>
                  <Text size="sm">
                    Last login: {new Date(user.last_login).toLocaleDateString()}
                  </Text>
                </Group>
              )}
              {userOrganizations.length > 0 && (
                <Group>
                  <ThemeIcon size="sm" color="violet" variant="light">
                    <IconBuilding size="0.8rem" />
                  </ThemeIcon>
                  <Text size="sm">
                    Member of {userOrganizations.length} organization{userOrganizations.length !== 1 ? 's' : ''}
                  </Text>
                </Group>
              )}
            </Stack>
          </Paper>
        </Grid.Col>
      </Grid>
    </Container>
  );
}
