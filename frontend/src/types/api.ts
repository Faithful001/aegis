export interface User {
  id: string;
  full_name: string;
  name?: string;
  first_name?: string;
  last_name?: string;
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
  is_configured: boolean;
  masked_api_key: string | null;
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
  organization_id: string;
  input_tokens: number;
  output_tokens: number;
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
  thread_id?: string;
  role: "system" | "user" | "assistant";
  content: string;
  timestamp?: Date | string;
  created_at?: string;
  model?: string;
}

export interface ChatThread {
  id: string;
  user_id?: string;
  title: string;
  model?: string;
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
  is_active?: boolean;
}

export interface PaginatedResult<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface TokenActivityBucket {
  date: string; // "YYYY-MM-DD" for daily/cumulative, "IYYY-WIW" for weekly
  tokens: number;
  requests: number;
}

export interface TokenActivityResponse {
  granularity: "daily" | "weekly" | "cumulative";
  buckets: TokenActivityBucket[];
  max_tokens: number;
}

export interface UserSession {
  id: string;
  device_name: string;
  ip_address: string;
  last_active_at: string;
  created_at: string;
  is_current: boolean;
}

export interface UserAlertSettings {
  id?: string;
  user_id?: string;
  is_budget_alert_enabled: boolean;
  monthly_budget_usd: number;
  threshold_percentage: number;
  alert_email: string;
  is_rate_limit_alert_enabled: boolean;
  is_failed_request_alert_enabled: boolean;
  is_modified_credential_alert_enabled: boolean;
  is_weekly_usage_alert_enabled: boolean;
}

export interface ConnectedAccountItem {
  provider: "google" | "github" | string;
  connected: boolean;
  email?: string;
  avatar_url?: string;
  connected_at?: string;
}

export interface ConnectedAccountsResponse {
  accounts: ConnectedAccountItem[];
  has_password: boolean;
}
