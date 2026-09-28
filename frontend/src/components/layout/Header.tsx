import React, { useState } from 'react';
import { Sparkles, Lock, ChevronDown, Bot, Key } from 'lucide-react';
import { useAegis, AVAILABLE_MODELS } from '../../context/AegisContext';
import { CreateKeyModal } from '../modals/CreateKeyModal';
import { AddCredentialModal } from '../modals/AddCredentialModal';

export const Header: React.FC = () => {
  const { selectedModel, setSelectedModel, activeApiKey, activeOrg } = useAegis();
  const [modelDropdownOpen, setModelDropdownOpen] = useState(false);
  const [keyModalOpen, setKeyModalOpen] = useState(false);
  const [credModalOpen, setCredModalOpen] = useState(false);

  const currentModelObj = AVAILABLE_MODELS.find((m) => m.id === selectedModel) || AVAILABLE_MODELS[0];

  return (
    <header className="h-14 px-6 border-b border-surface-border/40 flex items-center justify-between select-none bg-background/50 backdrop-blur-md sticky top-0 z-20">
      {/* Left side empty or active thread title */}
      <div></div>

      {/* Right side Controls: Model Selector, Private Badge, Sparkle */}
      <div className="flex items-center gap-3">
        {/* Model Selector Dropdown */}
        <div className="relative">
          <button
            onClick={() => setModelDropdownOpen(!modelDropdownOpen)}
            className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-surface-card border border-surface-border text-xs font-medium text-zinc-200 hover:bg-zinc-800 transition-all shadow-sm"
          >
            <Bot className="w-3.5 h-3.5 text-zinc-400" />
            <span>{currentModelObj.name}</span>
            <ChevronDown className="w-3.5 h-3.5 text-zinc-400" />
          </button>

          {modelDropdownOpen && (
            <div className="absolute right-0 top-full mt-1.5 w-64 bg-surface border border-surface-border rounded-2xl p-1.5 shadow-2xl z-50 space-y-1">
              <div className="px-3 py-1.5 text-[10px] uppercase font-bold text-zinc-400 tracking-wider">
                Select Inference Model
              </div>
              {AVAILABLE_MODELS.map((model) => (
                <button
                  key={model.id}
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
                    <span>{model.name}</span>
                    {model.fast && <span className="text-[10px] text-emerald-400">Fast</span>}
                  </div>
                  <p className="text-[11px] text-zinc-400 mt-0.5">{model.description}</p>
                </button>
              ))}
            </div>
          )}
        </div>

        {/* API Key / Credential Status */}
        <button
          onClick={() => setKeyModalOpen(true)}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-zinc-900 border border-zinc-800 text-xs font-medium text-zinc-300 hover:border-zinc-700 transition-all"
        >
          <Key className="w-3.5 h-3.5 text-amber-400" />
          <span className="hidden sm:inline">{activeApiKey ? 'API Key Active' : 'Set Key'}</span>
        </button>

        {/* Private Privacy Badge */}
        <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-surface-card border border-surface-border text-xs font-medium text-zinc-300">
          <Lock className="w-3.5 h-3.5 text-zinc-400" />
          <span>Private</span>
        </div>

        {/* Yellow Sparkle Highlight Icon */}
        <div className="w-8 h-8 rounded-full bg-amber-500/10 border border-amber-500/30 flex items-center justify-center text-amber-400 shadow-glow-yellow">
          <Sparkles className="w-4 h-4" />
        </div>
      </div>

      <CreateKeyModal open={keyModalOpen} onOpenChange={setKeyModalOpen} />
      <AddCredentialModal open={credModalOpen} onOpenChange={setCredModalOpen} />
    </header>
  );
};
