export interface User {
  id: string;
  first_name: string;
  last_name: string;
  name?: string;
  email: string;
  status: string;
  is_verified: boolean;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  user: User;
  token: AuthToken;
}

export interface AuthToken {
  token?: string;
  access_token?: string;
  refresh_token?: string;
}

export interface ProviderMeta {
  id: string;
  name: string;
  desc: string;
  placeholder: string;
  key_link: string;
  icon: string;
  icon_style: string;
}

export interface ProviderCredential {
  id?: string;
  user_id?: string;
  provider: "openai" | "openrouter" | "anthropic" | "gemini" | "mistral" | string;
  masked_api_key?: string;
  api_key_masked?: string;
  base_url?: string;
  created_at?: string;
  updated_at?: string;
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
  role: "system" | "user" | "assistant";
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

export interface DiscoveredModel {
  id: string;
  name: string;
  provider: string;
  description?: string;
  fast?: boolean;
}
