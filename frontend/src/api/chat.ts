import axios from 'axios';
import { ChatCompletionRequest, ChatCompletionChunk } from '../types/api';

export const streamChatCompletion = async (
  apiKey: string,
  request: ChatCompletionRequest,
  onChunk: (text: string) => void,
  onComplete: () => void,
  onError: (err: Error) => void
) => {
  try {
    const response = await axios.post(
      '/v1/chat/completions',
      { ...request, stream: true },
      {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${apiKey}`,
        },
        responseType: 'stream',
      }
    );

    const stream = response.data;
    if (stream && typeof stream.getReader === 'function') {
      const reader = stream.getReader();
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
            const dataStr = trimmed.substring(6);
            if (dataStr === '[DONE]') {
              onComplete();
              return;
            }
            try {
              const parsed: ChatCompletionChunk = JSON.parse(dataStr);
              const content = parsed.choices?.[0]?.delta?.content || parsed.choices?.[0]?.message?.content || '';
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
    } else {
      const textData = typeof response.data === 'string' ? response.data : JSON.stringify(response.data);
      onChunk(textData);
      onComplete();
    }
  } catch (err: any) {
    const errMsg = err?.response?.data?.message || err?.message || String(err);
    onError(new Error(errMsg));
  }
};

