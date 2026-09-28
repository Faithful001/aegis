import React, { useState, useRef, useEffect } from 'react';
import { Plus, ChevronDown, Mic, Send, Check } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { useAegis, AVAILABLE_MODELS } from '../../context/AegisContext';

interface ChatPromptInputProps {
  onSendMessage: (prompt: string) => void;
  disabled?: boolean;
}

export const ChatPromptInput: React.FC<ChatPromptInputProps> = ({ onSendMessage, disabled }) => {
  const [input, setInput] = useState('');
  const [modelDropdownOpen, setModelDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const { selectedModel, setSelectedModel } = useAegis();

  const currentModelObj =
    AVAILABLE_MODELS.find((m) => m.id === selectedModel) || AVAILABLE_MODELS[0];

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setModelDropdownOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

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

          {/* Right Action Icons: Model Selector Dropdown, Mic Icon, Send */}
          <div className="flex items-center gap-2">
            {/* Model Selector Dropdown inside prompt bar */}
            <div className="relative" ref={dropdownRef}>
              <button
                type="button"
                onClick={() => setModelDropdownOpen(!modelDropdownOpen)}
                className="flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium bg-zinc-800/80 hover:bg-zinc-800 text-zinc-200 border border-zinc-700/60 transition-all"
              >
                <span>{currentModelObj.name}</span>
                <ChevronDown
                  className={`w-3.5 h-3.5 text-zinc-400 transition-transform duration-200 ${
                    modelDropdownOpen ? 'rotate-180' : ''
                  }`}
                />
              </button>

              <AnimatePresence>
                {modelDropdownOpen && (
                  <motion.div
                    initial={{ opacity: 0, scale: 0.95, y: 8 }}
                    animate={{ opacity: 1, scale: 1, y: 0 }}
                    exit={{ opacity: 0, scale: 0.95, y: 8 }}
                    transition={{ duration: 0.15, ease: 'easeOut' }}
                    className="absolute right-0 bottom-full mb-2 w-64 bg-surface border border-surface-border rounded-2xl p-1.5 shadow-2xl z-50 space-y-1"
                  >
                    <div className="px-3 py-1.5 text-[10px] uppercase font-bold text-zinc-400 tracking-wider">
                      Select Model
                    </div>
                    {AVAILABLE_MODELS.map((model) => (
                      <button
                        key={model.id}
                        type="button"
                        onClick={() => {
                          setSelectedModel(model.id);
                          setModelDropdownOpen(false);
                        }}
                        className={`w-full text-left px-3 py-2 rounded-xl transition-all ${
                          selectedModel === model.id
                            ? 'bg-zinc-800 text-white font-medium'
                            : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200'
                        }`}
                      >
                        <div className="flex items-center justify-between text-xs font-semibold">
                          <span className="flex items-center gap-1.5">
                            {model.name}
                            {selectedModel === model.id && (
                              <Check className="w-3.5 h-3.5 text-emerald-400" />
                            )}
                          </span>
                          {model.fast && <span className="text-[10px] text-emerald-400">Fast</span>}
                        </div>
                        <p className="text-[11px] text-zinc-400 mt-0.5">{model.description}</p>
                      </button>
                    ))}
                  </motion.div>
                )}
              </AnimatePresence>
            </div>

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


