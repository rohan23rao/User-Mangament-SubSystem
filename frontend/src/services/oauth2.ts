// frontend/src/services/oauth2.ts
import axios from 'axios';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:3000';

// Define types locally to avoid circular imports
interface OAuth2Client {
  id: string;
  client_id: string;
  client_secret?: string;
  user_id: string;
  org_id: string;
  name: string;
  description: string;
  scopes: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  last_used_at?: string;
}

interface CreateM2MClientRequest {
  name: string;
  description: string;
  org_id: string;
  scopes?: string;
}

interface TokenRequest {
  client_id: string;
  client_secret: string;
  grant_type?: string;
  scope?: string;
}

interface TokenResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  scope: string;
  refresh_token?: string;
}

interface OAuth2ClientsResponse {
  clients: OAuth2Client[];
  count: number;
}

interface TokenInfo {
  active: boolean;
  client_id: string;
  scope: string;
  expires_at: string;
  issued_at?: string;
  subject?: string;
}

// Create axios instance with proper configuration
const oauth2Api = axios.create({
  baseURL: API_URL,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

export class OAuth2Service {
  // Client Management
  static async createClient(data: CreateM2MClientRequest): Promise<OAuth2Client> {
    const response = await oauth2Api.post('/api/oauth2/clients', data);
    return response.data;
  }

  static async getClients(): Promise<OAuth2ClientsResponse> {
    const response = await oauth2Api.get('/api/oauth2/clients');
    return response.data;
  }

  static async getClient(clientId: string): Promise<OAuth2Client> {
    const response = await oauth2Api.get(`/api/oauth2/clients/${clientId}`);
    return response.data;
  }

  static async revokeClient(clientId: string): Promise<void> {
    await oauth2Api.delete(`/api/oauth2/clients/${clientId}`);
  }

  static async regenerateClientSecret(clientId: string): Promise<OAuth2Client> {
    const response = await oauth2Api.post(`/api/oauth2/clients/${clientId}/regenerate`);
    return response.data;
  }

  // Token Management
  static async generateToken(data: TokenRequest): Promise<TokenResponse> {
    const response = await oauth2Api.post('/api/oauth2/token', data);
    return response.data;
  }

  static async validateToken(token: string): Promise<TokenInfo> {
    const response = await oauth2Api.post('/api/oauth2/validate', {}, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    return response.data;
  }

  // Utility Methods
  static async testClientCredentials(clientId: string, clientSecret: string): Promise<boolean> {
    try {
      await this.generateToken({
        client_id: clientId,
        client_secret: clientSecret,
        grant_type: 'client_credentials',
      });
      return true;
    } catch (error) {
      return false;
    }
  }

  static formatScopes(scopes: string): string[] {
    return scopes.split(' ').filter(scope => scope.trim() !== '');
  }

  static isTokenExpired(expiresAt: Date): boolean {
    return new Date() >= expiresAt;
  }

  static getTokenExpiryTime(expiresIn: number): Date {
    const expiryDate = new Date();
    expiryDate.setSeconds(expiryDate.getSeconds() + expiresIn);
    return expiryDate;
  }

  static formatTokenExpiry(expiryDate: Date): string {
    const now = new Date();
    const diff = expiryDate.getTime() - now.getTime();
    
    if (diff <= 0) return 'Expired';
    
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    
    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    }
    return `${minutes}m`;
  }

  // Generate example code snippets
  static generateCurlExample(token: string, endpoint: string = 'https://api.yourservice.com/data'): string {
    return `curl -H "Authorization: Bearer ${token}" \\
     -H "Content-Type: application/json" \\
     ${endpoint}`;
  }

  static generateJavaScriptExample(token: string): string {
    return `const response = await fetch('https://api.yourservice.com/data', {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer ${token}',
    'Content-Type': 'application/json'
  }
});

const data = await response.json();
console.log(data);`;
  }

  static generatePythonExample(token: string): string {
    return `import requests

headers = {
    'Authorization': 'Bearer ${token}',
    'Content-Type': 'application/json'
}

response = requests.get('https://api.yourservice.com/data', headers=headers)
data = response.json()
print(data)`;
  }

  static generateGoExample(clientId: string, clientSecret: string): string {
    return `package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strings"
)

func getAccessToken() (string, error) {
    data := url.Values{}
    data.Set("grant_type", "client_credentials")
    data.Set("client_id", "${clientId}")
    data.Set("client_secret", "${clientSecret}")

    req, _ := http.NewRequest("POST", "http://localhost:3000/api/oauth2/token", 
        strings.NewReader(data.Encode()))
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var tokenResp map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&tokenResp)
    
    return tokenResp["access_token"].(string), nil
}`;
  }
}

// Response interceptor for error handling
oauth2Api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Handle unauthorized access
      console.error('OAuth2 API: Unauthorized access');
    }
    return Promise.reject(error);
  }
);

// Request interceptor for debugging
oauth2Api.interceptors.request.use(
  (config) => {
    console.log(`OAuth2 API Request: ${config.method?.toUpperCase()} ${config.url}`);
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);