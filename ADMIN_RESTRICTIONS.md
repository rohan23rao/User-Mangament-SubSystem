# 🔒 Admin Restrictions Implementation

## Overview
As of this update, normal users can no longer create organizations. Only existing administrators can create new organizations.

## How It Works

### Backend Restrictions
- **New Function**: `isAdminOfAnyOrg(userID)` - checks if user has admin role in any organization
- **Bootstrap Protection**: `hasAnyAdmins()` - allows creation when no admins exist (fresh deployment)
- **Organization Creation**: Now requires admin permissions unless it's a bootstrap scenario

### Frontend Restrictions  
- **Create Button**: Only visible to users with admin role in at least one organization
- **Bootstrap Mode**: Button appears if no organizations exist at all

### Permission Logic
```go
// Backend: organization creation check
if !s.isAdminOfAnyOrg(session.Identity.Id) && s.hasAnyAdmins() {
    return "Forbidden - Only existing organization administrators can create new organizations"
}
```

```typescript  
// Frontend: button visibility check
const canCreateOrganizations = () => {
    const hasAdminRole = user?.organizations?.some(org => org.role === 'admin') || false;
    const noOrganizationsExist = !organizations || organizations.length === 0;
    return hasAdminRole || noOrganizationsExist;
};
```

## Updated Ways to Get Admin Permissions

### ✅ Still Available:
1. **Get Invited as Admin** - Existing admins can invite new users directly with admin role
2. **Get Promoted** - Existing admins can promote members to admin 
3. **Bootstrap First Org** - When no admins exist, anyone can create the first organization

### ❌ No Longer Available:
1. **~~Create Organization as Regular User~~** - This path is now blocked

## Migration Impact

### Existing Users:
- **Current admins**: No impact, can still create organizations
- **Current members**: Can no longer create organizations, must be promoted by admins

### New Deployments:
- **First user**: Can create the bootstrap organization and becomes admin
- **Subsequent users**: Must be invited by the first admin

## Bootstrap Process (Fresh Deployment)

1. First user registers and logs in
2. Since no organizations exist, they can create the first organization
3. They become owner/admin of that organization  
4. They can now invite other users and grant admin permissions
5. Only admins can create additional organizations

## Security Benefits

- **Controlled Growth**: Prevents organization sprawl by unauthorized users
- **Admin Governance**: Ensures administrative oversight of all organizations
- **Permission Hierarchy**: Maintains clear admin/member distinction
- **Bootstrap Safety**: Doesn't lock out fresh deployments

## API Changes

### Organization Creation Endpoint
- **Endpoint**: `POST /api/organizations`
- **New Behavior**: Returns 403 Forbidden for non-admin users (when admins exist)
- **Error Message**: "Forbidden - Only existing organization administrators can create new organizations"

### Frontend Changes
- **Organizations Page**: Create button only visible to admins
- **Navigation**: Admin-only features properly gated
- **User Experience**: Clear messaging for permission requirements 