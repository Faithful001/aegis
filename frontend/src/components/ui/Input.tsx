import React, { useState } from "react";
import { Eye, EyeOff } from "lucide-react";
import { cn } from "./Button";

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  showToggle?: boolean;
}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, error, type = "text", showToggle, ...props }, ref) => {
    const [visible, setVisible] = useState(false);

    const isPassword = type === "password";
    const inputType = isPassword && showToggle ? (visible ? "text" : "password") : type;

    return (
      <div className="w-full space-y-1.5">
        {label && <label className="block text-xs font-medium text-zinc-400">{label}</label>}
        <div className="relative">
          <input
            type={inputType}
            ref={ref}
            className={cn(
              "w-full px-3.5 py-2 text-sm bg-surface-card border border-surface-border text-zinc-100 rounded-xl placeholder:text-zinc-500 focus:outline-none focus:border-zinc-500 focus:ring-1 focus:ring-zinc-500 transition-all duration-150",
              isPassword && showToggle && "pr-10",
              error && "border-red-500/80 focus:border-red-500 focus:ring-red-500",
              className
            )}
            {...props}
          />
          {isPassword && showToggle && (
            <button
              type="button"
              onClick={() => setVisible((v) => !v)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300 transition-colors"
              tabIndex={-1}
              aria-label={visible ? "Hide password" : "Show password"}
            >
              {visible ? <EyeOff size={15} /> : <Eye size={15} />}
            </button>
          )}
        </div>
        {error && <p className="text-xs text-red-400 font-medium">{error}</p>}
      </div>
    );
  }
);

Input.displayName = "Input";

