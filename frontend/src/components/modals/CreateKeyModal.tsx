import React, { useState } from 'react';
import { Dialog } from '../ui/Dialog';
import { Input } from '../ui/Input';
import { Button } from '../ui/Button';
import { useAegis } from '../../context/AegisContext';
import { projectsApi } from '../../api/projects';
import { toast } from 'sonner';
import { Copy, Check } from 'lucide-react';

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export const CreateKeyModal: React.FC<Props> = ({ open, onOpenChange }) => {
  const [name, setName] = useState('');
  const [generatedKey, setGeneratedKey] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [loading, setLoading] = useState(false);
  const { activeProject, refreshKeys, setActiveApiKey } = useAegis();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!activeProject) {
      toast.error('No project selected.');
      return;
    }
    if (!name.trim()) return;
    setLoading(true);
    try {
      const keyObj = await projectsApi.createApiKey(activeProject.id, name.trim());
      const rawSecret = keyObj.secret || `aegis_${keyObj.key_prefix}_${Date.now()}`;
      setGeneratedKey(rawSecret);
      setActiveApiKey(rawSecret);
      toast.success('API key generated successfully!');
      await refreshKeys();
      setName('');
    } catch (err: any) {
      toast.error(err?.response?.data?.message || 'Failed to generate API key');
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = () => {
    if (generatedKey) {
      navigator.clipboard.writeText(generatedKey);
      setCopied(true);
      toast.success('API key copied to clipboard!');
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleClose = () => {
    setGeneratedKey(null);
    onOpenChange(false);
  };

  return (
    <Dialog
      open={open}
      onOpenChange={handleClose}
      title={generatedKey ? 'API Key Generated' : 'Generate Project API Key'}
      description={
        generatedKey
          ? 'Make sure to copy your secret API key now. You won’t be able to see it again!'
          : `Create an authentication key for project ${activeProject?.name || ''}.`
      }
    >
      {generatedKey ? (
        <div className="space-y-4">
          <div className="p-3.5 bg-zinc-900 border border-zinc-700/80 rounded-xl flex items-center justify-between font-mono text-xs text-amber-400 break-all">
            <span>{generatedKey}</span>
            <Button size="icon" variant="ghost" onClick={handleCopy} className="ml-2 flex-shrink-0 text-zinc-300">
              {copied ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4" />}
            </Button>
          </div>
          <p className="text-xs text-zinc-400">
            This API Key has been set as your active inference key for Aegis model requests.
          </p>
          <div className="flex justify-end pt-2">
            <Button onClick={handleClose}>Done</Button>
          </div>
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label="Key Label / Description"
            placeholder="e.g. Production Worker Key"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
          <div className="flex justify-end gap-3 pt-3">
            <Button type="button" variant="ghost" onClick={handleClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? 'Generating...' : 'Generate Key'}
            </Button>
          </div>
        </form>
      )}
    </Dialog>
  );
};
