export function validateApiKey(
  providerId: string,
  apiKey: string,
  baseUrl?: string
): string | null {
  const trimmedKey = apiKey.trim();
  if (!trimmedKey) {
    return "API key is required";
  }

  if (/\s/.test(trimmedKey)) {
    return "API key cannot contain spaces or whitespace";
  }

  if (trimmedKey.length < 10) {
    return "API key is too short (minimum 10 characters)";
  }

  if (trimmedKey.length > 512) {
    return "API key is too long (maximum 512 characters)";
  }

  const hasCustomBaseUrl = !!baseUrl && baseUrl.trim().length > 0;
  if (!hasCustomBaseUrl) {
    const p = providerId.toLowerCase().trim();
    switch (p) {
      case "openai":
        if (!trimmedKey.startsWith("sk-")) {
          return "OpenAI API keys start with 'sk-' (e.g. sk-proj-...)";
        }
        break;
      case "anthropic":
        if (!trimmedKey.startsWith("sk-ant-")) {
          return "Anthropic API keys start with 'sk-ant-'";
        }
        break;
      case "gemini":
        if (!trimmedKey.startsWith("AIzaSy")) {
          return "Google Gemini API keys start with 'AIzaSy'";
        }
        break;
      case "openrouter":
        if (!trimmedKey.startsWith("sk-or-")) {
          return "OpenRouter API keys start with 'sk-or-'";
        }
        break;
      case "mistral":
        if (trimmedKey.length < 20) {
          return "Mistral API keys must be at least 20 characters";
        }
        break;
    }
  }

  return null;
}

/**
 * Validates a custom base URL.
 * Base URL is optional, but if provided, must be a valid HTTP or HTTPS URL.
 * Returns null if valid (or empty), or a descriptive error string if invalid.
 */
export function validateBaseUrl(baseUrl?: string): string | null {
  if (!baseUrl) return null;
  const trimmed = baseUrl.trim();
  if (!trimmed) return null;

  if (/\s/.test(trimmed)) {
    return "Base URL cannot contain spaces or whitespace";
  }

  if (trimmed.length > 2048) {
    return "Base URL is too long (maximum 2048 characters)";
  }

  try {
    const parsed = new URL(trimmed);
    if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
      return "Base URL must start with http:// or https://";
    }
    if (!parsed.hostname) {
      return "Base URL must contain a valid hostname";
    }
  } catch {
    return "Please enter a valid URL (e.g. https://api.openai.com/v1)";
  }

  return null;
}
