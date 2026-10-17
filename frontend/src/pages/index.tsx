import React, { useState } from "react";
import { MainLayout } from "../components/layout/MainLayout";
import { ChatPromptInput } from "../components/chat/ChatPromptInput";
import { ChatMessages } from "../components/chat/ChatMessages";
import { useAegis } from "../context/AegisContext";
import { useAuth } from "../context/AuthContext";
import { streamChatCompletion } from "../api/chat";
import { AddCredentialModal } from "../components/modals/AddCredentialModal";
import { AuthModal } from "../components/modals/AuthModal";
import { toast } from "sonner";
import { Key } from "lucide-react";

export const HomePage: React.FC = () => {
  const { isAuthenticated } = useAuth();
  const {
    chatThreads,
    activeThreadId,
    selectedModel,
    models,
    hasCredential,
    addMessageToThread,
    updateLastMessageInThread,
    createNewThread,
    persistMessage,
  } = useAegis();

  const [isStreaming, setIsStreaming] = useState(false);
  const [credModalOpen, setCredModalOpen] = useState(false);
  const [authModalOpen, setAuthModalOpen] = useState(false);

  const activeThread = chatThreads.find((t) => t.id === activeThreadId);
  const messages = activeThread?.messages || [];

  const modelList = Array.isArray(models) ? models : [];
  const selectedModelMeta =
    modelList.find((m) => m.id === selectedModel) || (modelList.length > 0 ? modelList[0] : null);

  const handleSendMessage = async (promptText: string) => {
    if (!isAuthenticated) {
      toast.info("Please sign in to continue");
      setAuthModalOpen(true);
      return;
    }

    let currentThreadId = activeThreadId;
    if (!currentThreadId) {
      currentThreadId = await createNewThread(promptText, selectedModel);
    }

    const userMsg = {
      id: "msg-" + Date.now(),
      role: "user" as const,
      content: promptText,
      timestamp: new Date(),
    };

    addMessageToThread(currentThreadId, userMsg);
    console.log("currentThreadId", currentThreadId);
    // Persist user prompt to backend thread
    persistMessage(currentThreadId, "user", promptText);

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

    let fullAssistantResponse = "";

    await streamChatCompletion(
      {
        model: selectedModel,
        messages: [...messages, userMsg].map((m) => ({ role: m.role, content: m.content })),
      },
      (chunk) => {
        fullAssistantResponse += chunk;
        updateLastMessageInThread(currentThreadId!, chunk);
      },
      async () => {
        setIsStreaming(false);
        if (fullAssistantResponse.trim()) {
          await persistMessage(currentThreadId!, "assistant", fullAssistantResponse, selectedModel);
        }
      },
      (err) => {
        setIsStreaming(false);
        const errMsg = err.message || "Inference failed";
        toast.error(`Inference Error: ${errMsg}`);
        const errorContent = `\n\n*[Aegis Error: ${errMsg}. If you are using a BYOK model, ensure your provider key is saved in your Vault.]*`;
        updateLastMessageInThread(currentThreadId!, errorContent);
        fullAssistantResponse += errorContent;
        if (fullAssistantResponse.trim()) {
          persistMessage(currentThreadId!, "assistant", fullAssistantResponse, selectedModel);
        }
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

        {messages.length > 0 && (
          <div className="pt-4 bg-gradient-to-t from-background via-background to-transparent sticky bottom-0 z-10">
            <ChatPromptInput onSendMessage={handleSendMessage} disabled={isStreaming} />
          </div>
        )}
      </div>

      <AddCredentialModal
        open={credModalOpen}
        onOpenChange={setCredModalOpen}
        defaultProvider={selectedModelMeta?.provider}
      />
      <AuthModal open={authModalOpen} onOpenChange={setAuthModalOpen} initialMode="login" />
    </MainLayout>
  );
};

export default HomePage;
