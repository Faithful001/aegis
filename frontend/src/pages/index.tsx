import React, { useState } from "react";
import { MainLayout } from "../components/layout/MainLayout";
import { ChatPromptInput } from "../components/chat/ChatPromptInput";
import { ChatMessages } from "../components/chat/ChatMessages";
import { AppExploreCards } from "../components/chat/AppExploreCards";
import { useAegis } from "../context/AegisContext";
import { useAuth } from "../context/AuthContext";
import { streamChatCompletion } from "../api/chat";
import { AddCredentialModal } from "../components/modals/AddCredentialModal";
import { CreateKeyModal } from "../components/modals/CreateKeyModal";
import { AuthModal } from "../components/modals/AuthModal";
import { toast } from "sonner";

export const HomePage: React.FC = () => {
  const { isAuthenticated } = useAuth();
  const {
    chatThreads,
    activeThreadId,
    activeApiKey,
    selectedModel,
    addMessageToThread,
    updateLastMessageInThread,
    createNewThread,
  } = useAegis();

  const [isStreaming, setIsStreaming] = useState(false);
  const [credModalOpen, setCredModalOpen] = useState(false);
  const [keyModalOpen, setKeyModalOpen] = useState(false);
  const [authModalOpen, setAuthModalOpen] = useState(false);

  const activeThread = chatThreads.find((t) => t.id === activeThreadId);
  const messages = activeThread?.messages || [];

  const handleSendMessage = async (promptText: string) => {
    if (!isAuthenticated) {
      toast.info("Please sign in to continue");
      setAuthModalOpen(true);
      return;
    }

    let currentThreadId = activeThreadId;
    if (!currentThreadId || messages.length === 0) {
      currentThreadId = createNewThread();
    }

    const userMsg = {
      id: "msg-" + Date.now(),
      role: "user" as const,
      content: promptText,
      timestamp: new Date(),
    };

    addMessageToThread(currentThreadId, userMsg);

    const assistantMsgId = "msg-" + (Date.now() + 1);
    const initialAssistantMsg = {
      id: assistantMsgId,
      role: "assistant" as const,
      content: "",
      timestamp: new Date(),
      model: selectedModel,
    };

    addMessageToThread(currentThreadId, initialAssistantMsg);
    setIsStreaming(true);

    const apiKeyToUse = activeApiKey || "aegis_default_demo_key";

    await streamChatCompletion(
      apiKeyToUse,
      {
        model: selectedModel,
        messages: [...messages, userMsg].map((m) => ({ role: m.role, content: m.content })),
      },
      (chunk) => {
        updateLastMessageInThread(currentThreadId!, chunk);
      },
      () => {
        setIsStreaming(false);
      },
      (err) => {
        setIsStreaming(false);
        toast.error(`Inference Error: ${err.message}`);
        updateLastMessageInThread(
          currentThreadId!,
          `\n\n*[Aegis Inference Error: ${err.message}. Ensure backend is running and valid API Key / BYOK is set.]*`
        );
      }
    );
  };

  return (
    <MainLayout>
      <div className="flex flex-col h-full justify-between pb-6">
        {/* Top Chat Area or Hero Section */}
        {messages.length === 0 ? (
          <div className="flex-1 flex flex-col items-center justify-center text-center px-4 pt-12">
            <h1 className="text-3xl font-bold tracking-tight text-white mb-8">
              What should we explore?
            </h1>
            <ChatPromptInput onSendMessage={handleSendMessage} disabled={isStreaming} />
          </div>
        ) : (
          <div className="flex-1 overflow-y-auto">
            <ChatMessages messages={messages} isStreaming={isStreaming} />
          </div>
        )}

        {/* Bottom Fixed Input when messages exist */}
        {messages.length > 0 && (
          <div className="pt-4 bg-gradient-to-t from-background via-background to-transparent sticky bottom-0 z-10">
            <ChatPromptInput onSendMessage={handleSendMessage} disabled={isStreaming} />
          </div>
        )}
      </div>

      <AddCredentialModal open={credModalOpen} onOpenChange={setCredModalOpen} />
      <CreateKeyModal open={keyModalOpen} onOpenChange={setKeyModalOpen} />
      <AuthModal open={authModalOpen} onOpenChange={setAuthModalOpen} initialMode="login" />
    </MainLayout>
  );
};

export default HomePage;

