import React from "react";

interface SettingsLayoutProps {
  children: React.ReactNode;
  title: string;
  subTitle: string;
}

const SettingsLayout: React.FC<SettingsLayoutProps> = ({ children, title, subTitle }) => {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold text-white tracking-tight">{title}</h1>
        <p className="text-xs text-zinc-400 mt-1">{subTitle}</p>
      </div>
      {children}
    </div>
  );
};

export default SettingsLayout;
