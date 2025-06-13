// frontend/src/components/layout/OrganizationSwitcher.tsx
import React, { useState, useEffect } from 'react';
import {
  Menu,
  UnstyledButton,
  Group,
  Text,
  Avatar,
  Badge,
  Divider,
  Stack,
  ScrollArea,
  ActionIcon,
  Tooltip,
  Box,
} from '@mantine/core';
import {
  IconChevronDown,
  IconBuilding,
  IconBuildingStore,
  IconPlus,
  IconCheck,
} from '@tabler/icons-react';
import { useAuth } from '../../hooks/useAuth';
import { Organization } from '../../types/organization';
import { ApiService } from '../../services/api';

interface OrganizationSwitcherProps {
  currentOrgId?: string;
  onOrganizationChange?: (org: Organization) => void;
  onCreateClick?: () => void;
  compact?: boolean;
}

interface GroupedOrganizations {
  organizations: Organization[];
  tenants: { [orgId: string]: Organization[] };
}

export function OrganizationSwitcher({ 
  currentOrgId, 
  onOrganizationChange, 
  onCreateClick,
  compact = false
}: OrganizationSwitcherProps) {
  const { user } = useAuth();
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [loading, setLoading] = useState(false);
  const [currentOrg, setCurrentOrg] = useState<Organization | null>(null);

  useEffect(() => {
    loadOrganizations();
  }, []);

  useEffect(() => {
    if (currentOrgId && organizations.length > 0) {
      const org = organizations.find(o => o.id === currentOrgId);
      setCurrentOrg(org || null);
    }
  }, [currentOrgId, organizations]);

  const loadOrganizations = async () => {
    try {
      setLoading(true);
      const data = await ApiService.getOrganizations();
      setOrganizations(data || []);
    } catch (error) {
      console.error('Failed to load organizations:', error);
    } finally {
      setLoading(false);
    }
  };

  const groupOrganizations = (): GroupedOrganizations => {
    const orgs = organizations.filter(o => o.org_type === 'organization');
    const tenants: { [orgId: string]: Organization[] } = {};
    
    organizations
      .filter(o => o.org_type === 'tenant' && o.parent_id)
      .forEach(tenant => {
        if (!tenants[tenant.parent_id!]) {
          tenants[tenant.parent_id!] = [];
        }
        tenants[tenant.parent_id!].push(tenant);
      });

    return { organizations: orgs, tenants };
  };

  const handleSelect = (org: Organization) => {
    setCurrentOrg(org);
    onOrganizationChange?.(org);
  };

  const getOrgIcon = (orgType: string) => {
    return orgType === 'organization' ? <IconBuilding size="1rem" /> : <IconBuildingStore size="0.9rem" />;
  };

  const getOrgColor = (orgType: string) => {
    return orgType === 'organization' ? 'blue' : 'green';
  };

  const getUserRole = (orgId: string) => {
    const userOrg = user?.organizations?.find(o => o.org_id === orgId);
    return userOrg?.role || 'member';
  };

  const canCreateOrganizations = () => {
    return user?.can_create_organizations || false;
  };

  const grouped = groupOrganizations();

  return (
    <Menu shadow="md" width={320} position="bottom-start">
      <Menu.Target>
        <UnstyledButton
          style={{
            padding: compact ? '6px 8px' : '8px 12px',
            borderRadius: '8px',
            border: '1px solid var(--mantine-color-gray-3)',
            width: compact ? 'auto' : '100%',
            minWidth: compact ? 'auto' : '200px',
            backgroundColor: 'var(--mantine-color-white)',
          }}
        >
          <Group justify="space-between" gap="xs">
            <Group gap="xs">
              {currentOrg ? (
                <>
                  <Avatar 
                    size={compact ? "xs" : "sm"} 
                    color={getOrgColor(currentOrg.org_type)}
                    radius="sm"
                  >
                    {getOrgIcon(currentOrg.org_type)}
                  </Avatar>
                  {!compact && (
                    <Box style={{ flex: 1, minWidth: 0 }}>
                      <Text size="sm" fw={500} truncate>
                        {currentOrg.name}
                      </Text>
                      <Group gap="xs">
                        <Badge 
                          size="xs" 
                          color={getOrgColor(currentOrg.org_type)}
                          variant="light"
                        >
                          {currentOrg.org_type}
                        </Badge>
                        {currentOrg.parent_name && (
                          <Text size="xs" c="dimmed" truncate>
                            under {currentOrg.parent_name}
                          </Text>
                        )}
                      </Group>
                    </Box>
                  )}
                </>
              ) : (
                <Text size="sm" c="dimmed">Select workspace</Text>
              )}
            </Group>
            <IconChevronDown size="1rem" />
          </Group>
        </UnstyledButton>
      </Menu.Target>

      <Menu.Dropdown>
        <Menu.Label>Switch Workspace</Menu.Label>
        
        <ScrollArea style={{ maxHeight: 400 }}>
          <Stack gap="xs">
            {grouped.organizations.map((org) => (
              <div key={org.id}>
                <Menu.Item
                  leftSection={
                    <Avatar size="xs" color="blue" radius="sm">
                      <IconBuilding size="0.8rem" />
                    </Avatar>
                  }
                  rightSection={
                    <Group gap="xs">
                      {currentOrg?.id === org.id && (
                        <IconCheck size="1rem" color="var(--mantine-color-blue-6)" />
                      )}
                      <Badge size="xs" color="blue" variant="light">
                        {getUserRole(org.id)}
                      </Badge>
                    </Group>
                  }
                  onClick={() => handleSelect(org)}
                >
                  <div>
                    <Text size="sm" fw={500}>{org.name}</Text>
                    <Text size="xs" c="dimmed">
                      Organization • {org.member_count || 0} members
                    </Text>
                  </div>
                </Menu.Item>

                {/* Show tenants under this organization */}
                {grouped.tenants[org.id] && (
                  <Box style={{ paddingLeft: '20px', borderLeft: '2px solid var(--mantine-color-gray-3)', marginLeft: '20px' }}>
                    {grouped.tenants[org.id].map((tenant) => (
                      <Menu.Item
                        key={tenant.id}
                        leftSection={
                          <Avatar size="xs" color="green" radius="sm">
                            <IconBuildingStore size="0.7rem" />
                          </Avatar>
                        }
                        rightSection={
                          <Group gap="xs">
                            {currentOrg?.id === tenant.id && (
                              <IconCheck size="1rem" color="var(--mantine-color-green-6)" />
                            )}
                            <Badge size="xs" color="green" variant="light">
                              {getUserRole(tenant.id)}
                            </Badge>
                          </Group>
                        }
                        onClick={() => handleSelect(tenant)}
                      >
                        <div>
                          <Text size="sm" fw={500}>{tenant.name}</Text>
                          <Text size="xs" c="dimmed">
                            Tenant • {tenant.member_count || 0} members
                          </Text>
                        </div>
                      </Menu.Item>
                    ))}
                  </Box>
                )}
              </div>
            ))}
          </Stack>
        </ScrollArea>

        {canCreateOrganizations() && (
          <>
            <Divider my="xs" />
            <Menu.Item
              leftSection={<IconPlus size="1rem" />}
              onClick={onCreateClick}
            >
              Create Organization / Tenant
            </Menu.Item>
          </>
        )}
      </Menu.Dropdown>
    </Menu>
  );
}