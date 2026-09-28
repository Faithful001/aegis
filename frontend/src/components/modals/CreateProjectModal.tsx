import React, { useState } from 'react';
import { Dialog } from '../ui/Dialog';
import { Input } from '../ui/Input';
import { Button } from '../ui/Button';
import { useAegis } from '../../context/AegisContext';
import { projectsApi } from '../../api/projects';
import { toast } from 'sonner';

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export const CreateProjectModal: React.FC<Props> = ({ open, onOpenChange }) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);
  const { activeOrg, refreshProjects, setActiveProject } = useAegis();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!activeOrg) {
      toast.error('Please select an active organization first.');
      return;
    }
    if (!name.trim()) return;
    setLoading(true);
    try {
      const proj = await projectsApi.create(activeOrg.id, name.trim(), description.trim() || undefined);
      toast.success(`Project "${proj.name}" created!`);
      await refreshProjects();
      setActiveProject(proj);
      setName('');
      setDescription('');
      onOpenChange(false);
    } catch (err: any) {
      toast.error(err?.response?.data?.message || 'Failed to create project');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title="Create New Project"
      description={`Add a new inference project under ${activeOrg?.name || 'your organization'}.`}
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Project Name"
          placeholder="e.g. Chatbot Production API"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
        />
        <Input
          label="Description (Optional)"
          placeholder="e.g. High throughput LLM worker service"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
        <div className="flex justify-end gap-3 pt-3">
          <Button type="button" variant="ghost" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button type="submit" disabled={loading}>
            {loading ? 'Creating...' : 'Create Project'}
          </Button>
        </div>
      </form>
    </Dialog>
  );
};
