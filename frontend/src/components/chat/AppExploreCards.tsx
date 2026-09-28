import React, { useState } from 'react';
import { Landmark, Key, Cpu, Sparkles, X } from 'lucide-react';
import { Button } from '../ui/Button';

interface AppCardProps {
  id: string;
  icon: React.ReactNode;
  title: string;
  description: string;
  actionText: string;
  onAction: () => void;
}

export const AppExploreCards: React.FC<{ onAddCredential?: () => void; onGenerateKey?: () => void }> = ({
  onAddCredential,
  onGenerateKey,
}) => {
  const [dismissed, setDismissed] = useState<string[]>([]);

  const cards: AppCardProps[] = [
    {
      id: 'finance',
      icon: <Landmark className="w-5 h-5 text-emerald-400" />,
      title: 'Finance',
      description: 'Connect accounts to manage your finances in chat',
      actionText: 'Add',
      onAction: () => alert('Finance module connected'),
    },
    {
      id: 'byok',
      icon: <Key className="w-5 h-5 text-amber-400" />,
      title: 'BYOK Provider Credentials',
      description: 'Store OpenAI/OpenRouter API keys in organization secret vault',
      actionText: 'Configure',
      onAction: () => onAddCredential?.(),
    },
    {
      id: 'inference',
      icon: <Cpu className="w-5 h-5 text-sky-400" />,
      title: 'Distributed Inference',
      description: 'Issue project API keys to stream LLM responses via Aegis Go worker plane',
      actionText: 'Manage Keys',
      onAction: () => onGenerateKey?.(),
    },
  ];

  const visibleCards = cards.filter((c) => !dismissed.includes(c.id));

  if (visibleCards.length === 0) return null;

  return (
    <div className="w-full max-w-3xl mx-auto px-4 mt-6 space-y-3">
      {visibleCards.map((card) => (
        <div
          key={card.id}
          className="bg-surface-card border border-surface-border rounded-2xl p-4 flex items-center justify-between shadow-sm hover:border-zinc-700/80 transition-all"
        >
          <div className="flex items-center gap-3.5">
            <div className="w-10 h-10 rounded-xl bg-zinc-900 border border-zinc-800 flex items-center justify-center flex-shrink-0">
              {card.icon}
            </div>
            <div>
              <h4 className="text-sm font-semibold text-zinc-100">{card.title}</h4>
              <p className="text-xs text-zinc-400 mt-0.5">{card.description}</p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={() => setDismissed((prev) => [...prev, card.id])}
              className="text-xs text-zinc-400 hover:text-zinc-200 transition-colors"
            >
              Dismiss
            </button>
            <Button
              size="sm"
              className="bg-white text-zinc-950 font-semibold px-4 py-1.5 rounded-full text-xs hover:bg-zinc-200"
              onClick={card.onAction}
            >
              {card.actionText}
            </Button>
          </div>
        </div>
      ))}
    </div>
  );
};
