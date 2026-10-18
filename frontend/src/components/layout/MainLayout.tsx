import React, { useState } from "react";
import { Outlet } from "@tanstack/react-router";
import { Sidebar } from "./Sidebar";
import { Header } from "./Header";
import { useAuth } from "../../context/AuthContext";

export const MainLayout: React.FC = () => {
  const { isAuthenticated } = useAuth();
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  return (
    <div className="flex h-screen bg-background overflow-hidden text-zinc-100">
      {isAuthenticated && (
        <Sidebar collapsed={sidebarCollapsed} setCollapsed={setSidebarCollapsed} />
      )}

      <div className="flex-1 flex flex-col min-w-0 h-full overflow-hidden">
        {!isAuthenticated && <Header />}
        <main className="flex-1 overflow-y-auto relative">
          <Outlet />
        </main>
      </div>
    </div>
  );
};
