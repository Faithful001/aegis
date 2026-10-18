import React, { useRef, useEffect, useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { ChatMessage } from "../../types/api";
import { Check, Copy } from "@phosphor-icons/react";
import { toast } from "sonner";

interface ChatMessagesProps {
  messages: ChatMessage[];
  isStreaming?: boolean;
}

export const ChatMessages: React.FC<ChatMessagesProps> = ({ messages, isStreaming }) => {
  const bottomRef = useRef<HTMLDivElement>(null);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, isStreaming]);

  const copyToClipboard = (id: string, text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  return (
    <div className="w-full max-w-3xl mx-auto px-4 py-8 space-y-8">
      {messages.map((msg) => {
        if (msg.role === "user") {
          return (
            <div key={msg.id} className="flex justify-end w-full">
              <div className="bg-[#1c1c1f] text-zinc-100 text-[15px] px-4 py-2.5 rounded-[18px] max-w-[80%] leading-relaxed border border-zinc-800/80 shadow-sm select-text font-sans">
                {msg.content}
              </div>
            </div>
          );
        }

        // Assistant response
        return (
          <div key={msg.id} className="w-full space-y-2 group">
            <div className="text-[#e4e4e7] leading-[1.75] font-normal tracking-normal select-text">
              <ReactMarkdown
                remarkPlugins={[remarkGfm]}
                components={{
                  code({ node, className, children, ...props }: any) {
                    const match = /language-(\w+)/.exec(className || "");
                    const isInline = !match && !String(children).includes("\n");

                    if (isInline) {
                      return (
                        <code
                          className="bg-[#1c1c20] text-[#f87171] border border-zinc-700/60 px-1.5 py-0.5 rounded-[6px] font-mono text-[13.5px] mx-0.5 align-baseline"
                          {...props}
                        >
                          {children}
                        </code>
                      );
                    }

                    return (
                      <pre className="bg-[#161618] border border-zinc-800/80 rounded-xl p-4 overflow-x-auto my-3 text-zinc-200 font-mono text-[13.5px] leading-relaxed">
                        <code className={className} {...props}>
                          {children}
                        </code>
                      </pre>
                    );
                  },
                  p({ children }: any) {
                    return <p className="mb-3 last:mb-0 leading-[1.75]">{children}</p>;
                  },
                  ul({ children }: any) {
                    return <ul className="list-disc pl-5 my-2 space-y-1.5">{children}</ul>;
                  },
                  ol({ children }: any) {
                    return <ol className="list-decimal pl-5 my-2 space-y-1.5">{children}</ol>;
                  },
                  li({ children }: any) {
                    return <li className="leading-[1.75]">{children}</li>;
                  },
                  strong({ children }: any) {
                    return <strong className="font-semibold text-zinc-100">{children}</strong>;
                  },
                  a({ href, children }: any) {
                    return (
                      <a
                        href={href}
                        target="_blank"
                        rel="noreferrer"
                        className="text-amber-400 hover:underline underline-offset-2"
                      >
                        {children}
                      </a>
                    );
                  },
                }}
              >
                {msg.content}
              </ReactMarkdown>
            </div>

            {/* Action buttons below assistant message */}
            {msg.content && (
              <div className="flex items-center gap-1.5 pt-1 opacity-0 group-hover:opacity-100 transition-opacity">
                <button
                  onClick={() => copyToClipboard(msg.id, msg.content)}
                  className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs text-zinc-400 hover:text-zinc-200 border border-zinc-800/80 hover:border-zinc-700 bg-[#161618]/60 hover:bg-zinc-800/50 transition-colors"
                  title="Copy response"
                >
                  {copiedId === msg.id ? (
                    <>
                      <Check className="w-3.5 h-3.5" />
                      <span className="">Copied</span>
                    </>
                  ) : (
                    <>
                      <Copy className="w-3.5 h-3.5" />
                      <span>Copy</span>
                    </>
                  )}
                </button>
              </div>
            )}
          </div>
        );
      })}

      {isStreaming && (
        <div className="flex items-center gap-2 text-xs text-zinc-400 py-1">
          <span className="w-2 h-2 rounded-full bg-white animate-pulse"></span>
          <span>Thinking...</span>
        </div>
      )}

      <div ref={bottomRef} />
    </div>
  );
};
