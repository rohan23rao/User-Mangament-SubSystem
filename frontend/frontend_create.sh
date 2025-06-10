#!/bin/bash

# You're already in the frontend directory, so let's create all the files
cd ~/umange/late-claude/frontend

echo "🚀 Creating frontend file structure..."

# Create directory structure
mkdir -p src/{components/{auth,layout,dashboard,organizations},pages,services,hooks,types,utils}
mkdir -p public

echo "✅ Directory structure created"

# Create public/index.html
cat > public/index.html << 'EOF'
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta name="theme-color" content="#000000" />
    <meta name="description" content="User Management System with Ory Kratos" />
    <title>UserMS - User Management System</title>
  </head>
  <body>
    <noscript>You need to enable JavaScript to run this app.</noscript>
    <div id="root"></div>
  </body>
</html>
EOF

# Create tsconfig.json
cat > tsconfig.json << 'EOF'
{
  "compilerOptions": {
    "target": "es5",
    "lib": [
      "dom",
      "dom.iterable",
      "es6"
    ],
    "allowJs": true,
    "skipLibCheck": true,
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "strict": true,
    "forceConsistentCasingInFileNames": true,
    "noFallthroughCasesInSwitch": true,
    "module": "esnext",
    "moduleResolution": "node",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx"
  },
  "include": [
    "src"
  ]
}
EOF

# Create .env
cat > .env << 'EOF'
REACT_APP_API_URL=http://localhost:3000
REACT_APP_KRATOS_PUBLIC_URL=http://localhost:4433
PORT=3001
EOF

# Create Dockerfile
cat > Dockerfile << 'EOF'
FROM node:18-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy source code
COPY . .

# Expose port
EXPOSE 3001

# Development command
CMD ["npm", "start"]
EOF

echo "✅ Configuration files created"

# Create src/index.tsx
cat > src/index.tsx << 'EOF'
import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);

root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
EOF

# Create src/theme.ts
cat > src/theme.ts << 'EOF'
import { MantineTheme } from '@mantine/core';

export const theme: Partial<MantineTheme> = {
  colorScheme: 'light',
  primaryColor: 'blue',
  fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif',
  headings: {
    fontFamily: 'Greycliff CF, Inter, -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif',
  },
  components: {
    Button: {
      styles: (theme) => ({
        root: {
          fontWeight: 500,
        },
      }),
    },
    Card: {
      styles: (theme) => ({
        root: {
          backgroundColor: theme.colorScheme === 'dark' ? theme.colors.dark[7] : theme.white,
        },
      }),
    },
    Paper: {
      styles: (theme) => ({
        root: {
          backgroundColor: theme.colorScheme === 'dark' ? theme.colors.dark[7] : theme.white,
        },
      }),
    },
  },
};
EOF

# Create src/utils/constants.ts
cat > src/utils/constants.ts << 'EOF'
export const API_ENDPOINTS = {
  KRATOS_PUBLIC: process.env.REACT_APP_KRATOS_PUBLIC_URL || 'http://localhost:4433',
  API_BASE: process.env.REACT_APP_API_URL || 'http://localhost:3000',
};

export const ROUTES = {
  LOGIN: '/login',
  REGISTER: '/register',
  DASHBOARD: '/dashboard',
  PROFILE: '/profile',
  USERS: '/users',
  ORGANIZATIONS: '/organizations',
  SETTINGS: '/settings',
};

export const USER_ROLES = {
  MEMBER: 'member',
  ADMIN: 'admin',
  OWNER: 'owner',
} as const;

export const ORG_TYPES = {
  DOMAIN: 'domain',
  ORGANIZATION: 'organization',
  TENANT: 'tenant',
} as const;
EOF

echo "✅ Basic files created"

# Create TypeScript type files
cat > src/types/auth.ts << 'EOF'
export interface KratosSession {
  id: string;
  active: boolean;
  expires_at: string;
  authenticated_at: string;
  authenticator_assurance_level: string;
  authentication_methods: Array<{
    method: string;
    aal: string;
    completed_at: string;
    provider?: string;
  }>;
  issued_at: string;
  identity: KratosIdentity;
}

export interface KratosIdentity {
  id: string;
  schema_id: string;
  schema_url: string;
  state: string;
  state_changed_at: string;
  traits: {
    email: string;
    name?: {
      first: string;
      last: string;
    };
  };
  verifiable_addresses: Array<{
    id: string;
    value: string;
    verified: boolean;
    via: string;
    status: string;
    created_at: string;
    updated_at: string;
  }>;
  recovery_addresses: Array<{
    id: string;
    value: string;
    via: string;
    created_at: string;
    updated_at: string;
  }>;
  metadata_public?: any;
  created_at: string;
  updated_at: string;
}

export interface LoginFlow {
  id: string;
  type: string;
  expires_at: string;
  issued_at: string;
  request_url: string;
  ui: {
    action: string;
    method: string;
    nodes: Array<{
      type: string;
      group: string;
      attributes: any;
      messages: any[];
      meta: any;
    }>;
  };
  created_at: string;
  updated_at: string;
  refresh?: boolean;
  requested_aal?: string;
}

export interface RegistrationFlow {
  id: string;
  type: string;
  expires_at: string;
  issued_at: string;
  request_url: string;
  ui: {
    action: string;
    method: string;
    nodes: Array<{
      type: string;
      group: string;
      attributes: any;
      messages: any[];
      meta: any;
    }>;
  };
  created_at: string;
  updated_at: string;
}
EOF

cat > src/types/user.ts << 'EOF'
export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  time_zone: string;
  ui_mode: string;
  traits: {
    email: string;
    name?: {
      first: string;
      last: string;
    };
  };
  organizations?: OrgMember[];
  created_at: string;
  updated_at: string;
  last_login?: string;
}

export interface OrgMember {
  org_id: string;
  org_name: string;
  org_type: string;
  role: string;
  joined_at: string;
}
EOF

cat > src/types/organization.ts << 'EOF'
export interface Organization {
  id: string;
  domain_id?: string;
  org_id?: string;
  org_type: string;
  name: string;
  description: string;
  owner_id?: string;
  data: Record<string, any>;
  members?: Member[];
  created_at: string;
  updated_at: string;
}

export interface Member {
  user_id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: string;
  joined_at: string;
}

export interface CreateOrgRequest {
  name: string;
  description: string;
  org_type: 'domain' | 'organization' | 'tenant';
  domain_id?: string;
  org_id?: string;
  data?: Record<string, any>;
}

export interface InviteUserRequest {
  email: string;
  role: 'member' | 'admin';
}

export interface UpdateMemberRoleRequest {
  role: 'member' | 'admin';
}

export type UserRole = 'member' | 'admin' | 'owner';
EOF

echo "✅ TypeScript types created"

echo ""
echo "🎉 Frontend structure created successfully!"
echo ""
echo "Next steps:"
echo "1. Copy the component files from the artifacts above"
echo "2. Add the frontend service to docker-compose.yml"
echo "3. Run: docker-compose up --build -d"
echo ""
echo "Files still needed:"
echo "- src/App.tsx"
echo "- src/services/auth.ts"
echo "- src/services/api.ts"
echo "- src/hooks/useAuth.tsx"
echo "- src/components/auth/ProtectedRoute.tsx"
echo "- src/components/layout/AppShell.tsx"
echo "- src/components/layout/Navigation.tsx"
echo "- src/pages/*.tsx (all page components)"
echo ""
echo "You can copy these from the artifacts I created above!"