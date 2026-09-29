import { ChatCompletionRequest, ChatCompletionChunk } from '../types/api';

export const streamChatCompletion = async (
  request: ChatCompletionRequest,
  onChunk: (text: string) => void,
  onComplete: () => void,
  onError: (err: Error) => void
) => {
  try {
    const token = sessionStorage.getItem('aegis_jwt_token');
    const response = await fetch('/api/v1/chat/completions', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify({ ...request, stream: true }),
    });

    if (!response.ok) {
      const errJson = await response.json().catch(() => ({}));
      const errMsg =
        errJson?.error?.message ||
        errJson?.message ||
        `HTTP ${response.status}: ${response.statusText}`;
      throw new Error(errMsg);
    }

    if (!response.body) {
      throw new Error('ReadableStream not supported by this browser.');
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder('utf-8');
    let buffer = '';

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';

      for (const line of lines) {
        const trimmed = line.trim();
        if (trimmed.startsWith('data: ')) {
          const dataStr = trimmed.substring(6).trim();
          if (dataStr === '[DONE]') {
            onComplete();
            return;
          }
          try {
            const parsed: ChatCompletionChunk = JSON.parse(dataStr);
            const content =
              parsed.choices?.[0]?.delta?.content ||
              parsed.choices?.[0]?.message?.content ||
              '';
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
