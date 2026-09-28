import React from 'react';
import { MainLayout } from '../../components/layout/MainLayout';
import { useAuth } from '../../context/AuthContext';
import { Card } from '../../components/ui/Card';
import { Button } from '../../components/ui/Button';
import { User, Mail, Shield, Calendar, LogOut } from 'lucide-react';
import { toast } from 'sonner';

export const ProfilePage: React.FC = () => {
  const { user, logout } = useAuth();

  const handleLogout = async () => {
    await logout();
    toast.success('Logged out successfully');
    window.location.href = '/auth/login';
  };

  return (
    <MainLayout>
      <div className="max-w-4xl mx-auto p-8 space-y-8">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight">User Profile</h1>
          <p className="text-xs text-zinc-400 mt-1">Manage your Aegis developer account details and credentials.</p>
        </div>

        <Card className="p-6 space-y-6">
          <div className="flex items-center gap-4 pb-6 border-b border-surface-border">
            <div className="w-16 h-16 rounded-full bg-zinc-800 border-2 border-amber-500/40 flex items-center justify-center text-zinc-200 text-xl font-bold shadow-glow-yellow">
              {user?.name ? user.name.slice(0, 2).toUpperCase() : 'FE'}
            </div>
            <div>
              <h3 className="text-lg font-bold text-white">{user?.name || 'Faithful Eromosele'}</h3>
              <p className="text-xs text-zinc-400 flex items-center gap-1.5 mt-0.5">
                <Mail className="w-3.5 h-3.5 text-zinc-500" /> {user?.email || 'developer@aegis.ai'}
              </p>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="p-4 bg-zinc-900/60 rounded-xl border border-zinc-800 space-y-1">
              <span className="text-[10px] font-bold text-zinc-500 uppercase tracking-wider">Account Role</span>
              <p className="text-sm font-semibold text-amber-400 flex items-center gap-1.5">
                <Shield className="w-4 h-4 text-amber-400" /> Platform Developer / Admin
              </p>
            </div>

            <div className="p-4 bg-zinc-900/60 rounded-xl border border-zinc-800 space-y-1">
              <span className="text-[10px] font-bold text-zinc-500 uppercase tracking-wider">Member Since</span>
              <p className="text-sm font-semibold text-zinc-200 flex items-center gap-1.5">
                <Calendar className="w-4 h-4 text-zinc-400" /> September 2026
              </p>
            </div>
          </div>

          <div className="pt-4 flex justify-end">
            <Button variant="danger" onClick={handleLogout} className="flex items-center gap-2">
              <LogOut className="w-4 h-4" /> Log Out
            </Button>
          </div>
        </Card>
      </div>
    </MainLayout>
  );
};

export default ProfilePage;
