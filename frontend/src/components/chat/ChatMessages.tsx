import React, { useRef, useEffect } from 'react';
import { ChatMessage } from '../../types/api';
import { Bot, User, Copy, Check } from 'lucide-react';
import { toast } from 'sonner';

interface ChatMessagesProps {
  messages: ChatMessage[];
  isStreaming?: boolean;
}

export const ChatMessages: React.FC<ChatMessagesProps> = ({ messages, isStreaming }) => {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, isStreaming]);

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast.success('Copied to clipboard');
  };

  return (
    <div className="w-full max-w-3xl mx-auto px-4 py-6 space-y-6">
      {messages.map((msg) => (
        <div key={msg.id} className="flex gap-4 items-start group">
          {msg.role === 'user' ? (
            <div className="w-8 h-8 rounded-full bg-zinc-800 border border-zinc-700 flex items-center justify-center text-zinc-300 text-xs font-semibold flex-shrink-0">
              <User className="w-4 h-4" />
            </div>
          ) : (
            <div className="w-8 h-8 rounded-full bg-zinc-900 border border-zinc-700 flex items-center justify-center text-amber-400 text-xs font-semibold flex-shrink-0 shadow-sm">
              <Bot className="w-4 h-4" />
            </div>
          )}

          <div className="flex-1 space-y-1 overflow-hidden">
            <div className="flex items-center justify-between text-xs text-zinc-500">
              <span className="font-semibold text-zinc-300">
                {msg.role === 'user' ? 'You' : 'Aegis Engine'}
              </span>
              <button
                onClick={() => copyToClipboard(msg.content)}
                className="opacity-0 group-hover:opacity-100 p-1 text-zinc-400 hover:text-white transition-opacity"
              >
                <Copy className="w-3.5 h-3.5" />
              </button>
            </div>

            <div className="text-sm text-zinc-200 leading-relaxed whitespace-pre-wrap font-sans">
              {msg.content}
            </div>
          </div>
        </div>
      ))}

      {isStreaming && (
        <div className="flex gap-4 items-center">
          <div className="w-8 h-8 rounded-full bg-zinc-900 border border-zinc-700 flex items-center justify-center text-amber-400">
            <Bot className="w-4 h-4 animate-spin" />
          </div>
          <div className="flex items-center gap-1.5 text-xs text-zinc-400">
            <span className="w-2 h-2 rounded-full bg-amber-400 animate-ping"></span>
            <span>Streaming tokens from Go worker pool...</span>
          </div>
        </div>
      )}

      <div ref={bottomRef} />
    </div>
  );
};
