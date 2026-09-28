import React, { useState } from 'react';
import { Plus, ChevronDown, Mic, Zap, Send } from 'lucide-react';

interface ChatPromptInputProps {
  onSendMessage: (prompt: string) => void;
  disabled?: boolean;
}

export const ChatPromptInput: React.FC<ChatPromptInputProps> = ({ onSendMessage, disabled }) => {
  const [input, setInput] = useState('');

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      submit();
    }
  };

  const submit = () => {
    if (!input.trim() || disabled) return;
    onSendMessage(input.trim());
    setInput('');
  };

  return (
    <div className="w-full max-w-3xl mx-auto px-4">
      <div className="relative bg-surface-card/90 border border-surface-border rounded-3xl p-3.5 shadow-glow-card transition-all focus-within:border-zinc-600 focus-within:ring-1 focus-within:ring-zinc-600">
        {/* Main Textarea */}
        <textarea
          rows={2}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Type @ to search your apps"
          disabled={disabled}
          className="w-full bg-transparent text-sm text-zinc-100 placeholder:text-zinc-500 focus:outline-none resize-none px-2 py-1 min-h-[48px]"
        />

        {/* Bottom Bar Controls inside prompt box */}
        <div className="flex items-center justify-between pt-2 border-t border-surface-border/40 px-1">
          {/* Left Attachment Button */}
          <button
            type="button"
            className="p-1.5 rounded-full text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors"
          >
            <Plus className="w-4 h-4" />
          </button>

          {/* Right Action Icons: Fast Speed Dropdown, Mic Icon, Send/Voice Wave */}
          <div className="flex items-center gap-2">
            <button
              type="button"
              className="flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-medium text-zinc-300 hover:bg-zinc-800 transition-colors"
            >
              <span>Fast</span>
              <ChevronDown className="w-3.5 h-3.5 text-zinc-400" />
            </button>

            <button
              type="button"
              className="p-1.5 rounded-full text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors"
            >
              <Mic className="w-4 h-4" />
            </button>

            <button
              type="button"
              onClick={submit}
              disabled={disabled || !input.trim()}
              className="w-8 h-8 rounded-full bg-sky-500 hover:bg-sky-400 text-white flex items-center justify-center transition-all disabled:opacity-40 disabled:hover:bg-sky-500"
            >
              <Send className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
