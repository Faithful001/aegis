import React, { useState } from 'react';
import { Dialog } from '../ui/Dialog';
import { Input } from '../ui/Input';
import { Button } from '../ui/Button';
import { useAegis } from '../../context/AegisContext';
import { orgsApi } from '../../api/orgs';
import { toast } from 'sonner';

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export const CreateOrgModal: React.FC<Props> = ({ open, onOpenChange }) => {
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [loading, setLoading] = useState(false);
  const { refreshOrgs, setActiveOrg } = useAegis();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    setLoading(true);
    try {
      const org = await orgsApi.create(name.trim(), slug.trim() || undefined);
      toast.success(`Organization "${org.name}" created successfully!`);
      await refreshOrgs();
      setActiveOrg(org);
      setName('');
      setSlug('');
      onOpenChange(false);
    } catch (err: any) {
      toast.error(err?.response?.data?.message || 'Failed to create organization');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title="Create New Organization"
      description="Organizations allow team members to share BYOK provider credentials, billing, and projects."
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Organization Name"
          placeholder="e.g. Acme AI Corp"
          value={name}
          onChange={(e) => {
            setName(e.target.value);
            if (!slug) setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9]/g, '-'));
          }}
          required
        />
        <Input
          label="Slug Identifier"
          placeholder="e.g. acme-ai-corp"
          value={slug}
          onChange={(e) => setSlug(e.target.value)}
        />
        <div className="flex justify-end gap-3 pt-3">
          <Button type="button" variant="ghost" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button type="submit" disabled={loading}>
            {loading ? 'Creating...' : 'Create Organization'}
          </Button>
        </div>
      </form>
    </Dialog>
  );
};
