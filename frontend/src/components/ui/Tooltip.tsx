import React, { useState, useRef, useCallback } from "react";
import { cn } from "./Button";

export type TooltipPlacement = "top" | "bottom" | "left" | "right";

interface TooltipProps {
  content: React.ReactNode;
  placement?: TooltipPlacement;
  delay?: number;
  className?: string;
  children: React.ReactElement;
}

/**
 * Lightweight, zero-dependency tooltip built on React state + CSS transforms.
 * Wraps any single child element. The tooltip panel is rendered inside a
 * relative container so it never escapes the viewport on small screens.
 */
export const Tooltip: React.FC<TooltipProps> = ({
  content,
  placement = "top",
  delay = 120,
  className,
  children,
}) => {
  const [visible, setVisible] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const show = useCallback(() => {
    timerRef.current = setTimeout(() => setVisible(true), delay);
  }, [delay]);

  const hide = useCallback(() => {
    if (timerRef.current) clearTimeout(timerRef.current);
    setVisible(false);
  }, []);

  // Position classes for the tooltip panel relative to the trigger
  const placementClasses: Record<TooltipPlacement, string> = {
    top: "bottom-full left-1/2 -translate-x-1/2 mb-2",
    bottom: "top-full left-1/2 -translate-x-1/2 mt-2",
    left: "right-full top-1/2 -translate-y-1/2 mr-2",
    right: "left-full top-1/2 -translate-y-1/2 ml-2",
  };

  // Arrow classes
  const arrowClasses: Record<TooltipPlacement, string> = {
    top: "top-full left-1/2 -translate-x-1/2 border-t-zinc-800 border-x-transparent border-b-transparent",
    bottom: "bottom-full left-1/2 -translate-x-1/2 border-b-zinc-800 border-x-transparent border-t-transparent",
    left: "left-full top-1/2 -translate-y-1/2 border-l-zinc-800 border-y-transparent border-r-transparent",
    right: "right-full top-1/2 -translate-y-1/2 border-r-zinc-800 border-y-transparent border-l-transparent",
  };

  return (
    <div
      className="relative inline-flex"
      onMouseEnter={show}
      onMouseLeave={hide}
      onFocus={show}
      onBlur={hide}
    >
      {children}

      {/* Tooltip panel */}
      <div
        role="tooltip"
        aria-hidden={!visible}
        className={cn(
          "pointer-events-none absolute z-50 whitespace-nowrap",
          "rounded-lg border border-zinc-700/60 bg-zinc-900 px-3 py-1.5 shadow-xl",
          "text-xs font-medium text-zinc-100",
          "transition-all duration-150",
          visible ? "opacity-100 scale-100" : "opacity-0 scale-95",
          placementClasses[placement],
          className
        )}
      >
        {content}
        {/* Arrow */}
        <span
          className={cn(
            "absolute border-4",
            arrowClasses[placement]
          )}
        />
      </div>
    </div>
  );
};

export default Tooltip;
