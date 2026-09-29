import React, { createContext, useContext, useState, useEffect } from "react";
import { ProviderCredential, ChatThread, ChatMessage } from "../types/api";
import { credentialsApi } from "../api/credentials";
import { useAuth } from "./AuthContext";

interface AegisContextType {
  credentials: ProviderCredential[];
  refreshCredentials: () => Promise<void>;
  hasCredential: (provider: string) => boolean;
  selectedModel: string;
  setSelectedModel: (model: string) => void;
  chatThreads: ChatThread[];
  activeThreadId: string | null;
  setActiveThreadId: (id: string | null) => void;
  createNewThread: () => string;
  addMessageToThread: (threadId: string, msg: ChatMessage) => void;
  updateLastMessageInThread: (threadId: string, content: string) => void;
}

const AegisContext = createContext<AegisContextType | undefined>(undefined);

export const AVAILABLE_MODELS = [
  {
    id: "mistral-small-latest",
    name: "Mistral Small",
    provider: "mistral",
    description: "Fast, balanced general reasoning model",
    fast: true,
  },
  {
    id: "mistral-large-latest",
    name: "Mistral Large",
    provider: "mistral",
    description: "Top-tier flagship reasoning & coding model",
    fast: false,
  },
  {
    id: "gpt-4o-mini",
    name: "OpenAI GPT-4o Mini",
    provider: "openai",
    description: "High speed direct OpenAI BYOK route",
    fast: true,
  },
  {
    id: "claude-3-5-sonnet",
    name: "Claude 3.5 Sonnet",
    provider: "anthropic",
    description: "Frontier coding & multi-turn reasoning",
    fast: false,
  },
  {
    id: "openrouter/auto",
    name: "OpenRouter Auto",
    provider: "openrouter",
    description: "Universal BYOK router with fallback routing",
    fast: true,
  },
];

export const AegisProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuth();
  const [credentials, setCredentials] = useState<ProviderCredential[]>([]);
  const [selectedModel, setSelectedModel] = useState<string>("mistral-small-latest");

  const [chatThreads, setChatThreads] = useState<ChatThread[]>([]);
  const [activeThreadId, setActiveThreadId] = useState<string | null>("default-thread-1");

  const refreshCredentials = async () => {
    if (!isAuthenticated) {
      setCredentials([]);
      return;
    }
    try {
      const list = await credentialsApi.list();
      setCredentials(list);
    } catch (e) {
      console.error("Error fetching user credentials:", e);
    }
  };

  const hasCredential = (provider: string): boolean => {
    return credentials.some((c) => c.provider.toLowerCase() === provider.toLowerCase());
  };

  useEffect(() => {
    if (isAuthenticated) {
      refreshCredentials();
    } else {
      setCredentials([]);
    }
  }, [isAuthenticated]);

  const createNewThread = (): string => {
    const newId = "thread-" + Date.now();
    const newThread: ChatThread = {
      id: newId,
      title: "New Exploration",
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      messages: [],
    };
    setChatThreads((prev) => [newThread, ...prev]);
    setActiveThreadId(newId);
    return newId;
  };

  const addMessageToThread = (threadId: string, msg: ChatMessage) => {
    setChatThreads((prev) =>
      prev.map((t) => {
        if (t.id === threadId) {
          const updatedMessages = [...t.messages, msg];
          let title = t.title;
          if (t.title === "New Exploration" && msg.role === "user") {
            title = msg.content.slice(0, 30) + (msg.content.length > 30 ? "..." : "");
          }
          return {
            ...t,
            title,
            updated_at: new Date().toISOString(),
            messages: updatedMessages,
          };
        }
        return t;
      })
    );
  };

  const updateLastMessageInThread = (threadId: string, content: string) => {
    setChatThreads((prev) =>
      prev.map((t) => {
        if (t.id === threadId && t.messages.length > 0) {
          const msgs = [...t.messages];
          const lastIndex = msgs.length - 1;
          msgs[lastIndex] = {
            ...msgs[lastIndex],
            content: msgs[lastIndex].content + content,
          };
          return {
            ...t,
            updated_at: new Date().toISOString(),
            messages: msgs,
          };
        }
        return t;
      })
    );
  };

  return (
    <AegisContext.Provider
      value={{
        credentials,
        refreshCredentials,
        hasCredential,
        selectedModel,
        setSelectedModel,
        chatThreads,
        activeThreadId,
        setActiveThreadId,
        createNewThread,
        addMessageToThread,
        updateLastMessageInThread,
      }}
    >
      {children}
    </AegisContext.Provider>
  );
};

export const useAegis = () => {
  const ctx = useContext(AegisContext);
  if (!ctx) throw new Error("useAegis must be used within AegisProvider");
  return ctx;
};
