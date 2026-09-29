import React, { useState } from 'react';
import { Sidebar } from './Sidebar';
import { Header } from './Header';
import { useAuth } from '../../context/AuthContext';

interface MainLayoutProps {
  children: React.ReactNode;
}

export const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  const { isAuthenticated } = useAuth();
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  return (
    <div className="flex h-screen bg-background overflow-hidden text-zinc-100">
      {isAuthenticated && (
        <Sidebar collapsed={sidebarCollapsed} setCollapsed={setSidebarCollapsed} />
      )}

      <div className="flex-1 flex flex-col min-w-0 h-full overflow-hidden">
        {!isAuthenticated && <Header />}
        <main className="flex-1 overflow-y-auto relative">{children}</main>
      </div>
    </div>
  );
};
