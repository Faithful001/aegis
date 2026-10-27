import { ChatCompletionRequest, ChatCompletionChunk } from "../types/api";
import { API_BASE_URL, getAccessToken, getRefreshToken, setTokens, clearTokens } from "./client";
import { authApi } from "./auth";

export const streamChatCompletion = async (
  request: ChatCompletionRequest,
  onChunk: (text: string) => void,
  onComplete: () => void,
  onError: (err: Error) => void
) => {
  try {
    const url = `${API_BASE_URL.replace(/\/+$/, "")}/chat/completions`;
    let token = getAccessToken();
    let response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify({ ...request, stream: true }),
    });

    if (response.status === 401) {
      const refreshToken = getRefreshToken();
      if (refreshToken) {
        try {
          const refreshRes = await authApi.refreshToken(refreshToken);
          const newToken = refreshRes.data?.token?.access_token;
          const newRefresh = refreshRes.data?.token?.refresh_token;
          if (newToken) {
            setTokens(newToken, newRefresh);
            token = newToken;
            response = await fetch(url, {
              method: "POST",
              headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${newToken}`,
              },
              body: JSON.stringify({ ...request, stream: true }),
            });
          }
        } catch {
          clearTokens();
          window.dispatchEvent(new Event("unauthorized"));
        }
      }
    }

    if (!response.ok) {
      const errJson = await response.json().catch(() => ({}));
      const errMsg =
        errJson?.error?.message ||
        errJson?.message ||
        `HTTP ${response.status}: ${response.statusText}`;
      throw new Error(errMsg);
    }

    if (!response.body) {
      throw new Error("ReadableStream not supported by this browser.");
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder("utf-8");
    let buffer = "";

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split("\n");
      buffer = lines.pop() || "";

      for (const line of lines) {
        const trimmed = line.trim();
        if (trimmed.startsWith("data: ")) {
          const dataStr = trimmed.substring(6).trim();
          if (dataStr === "[DONE]") {
            onComplete();
            return;
          }
          try {
            const parsed: ChatCompletionChunk = JSON.parse(dataStr);
            const content =
              parsed.choices?.[0]?.delta?.content || parsed.choices?.[0]?.message?.content || "";
            if (content) {
              onChunk(content);
            }
          } catch (e) {
            if (dataStr) onChunk(dataStr);
          }
        }
      }
    }
    onComplete();
  } catch (err: any) {
    onError(err instanceof Error ? err : new Error(String(err)));
  }
};
