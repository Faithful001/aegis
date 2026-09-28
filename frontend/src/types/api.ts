export interface User {
  id: string;
  email: string;
  name?: string;
  created_at?: string;
}

export interface AuthResponse {
  success: boolean;
  token?: string;
  refresh_token?: string;
  user?: User;
  data?: {
    user_id?: string;
    email?: string;
    token?: string;
    user?: User;
  };
  message?: string;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  owner_id: string;
  created_at: string;
  updated_at?: string;
}

export interface OrgMember {
  id: string;
  organization_id: string;
  user_id: string;
  role: string;
  email?: string;
  created_at: string;
}

export interface ProviderCredential {
  id?: string;
  organization_id: string;
  provider: 'openai' | 'openrouter' | 'anthropic' | string;
  api_key_masked?: string;
  created_at?: string;
}

export interface Project {
  id: string;
  organization_id: string;
  name: string;
  slug: string;
  description?: string;
  created_at: string;
}

export interface APIKey {
  id: string;
  project_id: string;
  name: string;
  key_prefix: string;
  secret?: string; // Only returned upon creation
  created_at: string;
  expires_at?: string;
  revoked?: boolean;
}

export interface UsageMetric {
  total_requests: number;
  total_tokens: number;
  prompt_tokens: number;
  completion_tokens: number;
  total_cost_usd: number;
  time_series?: {
    timestamp: string;
    tokens: number;
    requests: number;
  }[];
}

export interface ChatMessage {
  id: string;
  role: 'system' | 'user' | 'assistant';
  content: string;
  timestamp: Date;
  model?: string;
}

export interface ChatThread {
  id: string;
  title: string;
  created_at: string;
  updated_at: string;
  messages: ChatMessage[];
}

export interface ChatCompletionRequest {
  model: string;
  messages: { role: string; content: string }[];
  temperature?: number;
  max_tokens?: number;
  stream?: boolean;
}

export interface ChatCompletionChoice {
  index: number;
  delta?: { content?: string; role?: string };
  message?: { content: string; role: string };
  finish_reason?: string | null;
}

export interface ChatCompletionChunk {
  id: string;
  object: string;
  created: number;
  model: string;
  choices: ChatCompletionChoice[];
}
