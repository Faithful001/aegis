import React, { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { MainLayout } from "../components/layout/MainLayout";
import { ChatPromptInput } from "../components/chat/ChatPromptInput";
import { useAegis } from "../context/AegisContext";
import { useAuth } from "../context/AuthContext";
import { AddCredentialModal } from "../components/modals/AddCredentialModal";
import { AuthModal } from "../components/modals/AuthModal";
import { toast } from "sonner";
import { router } from "@/router";

const ROTATING_WORDS = ["explore", "search", "discover", "build"];

export const HomePage: React.FC = () => {
  const { isAuthenticated, user } = useAuth();
  const { selectedModel, models, createNewThread, sendMessage, isStreaming, setActiveThreadId } =
    useAegis();

  const [credModalOpen, setCredModalOpen] = useState(false);
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [wordIndex, setWordIndex] = useState(0);

  useEffect(() => {
    setActiveThreadId(null);
  }, []);

  useEffect(() => {
    const timer = setInterval(() => {
      setWordIndex((prev) => (prev + 1) % ROTATING_WORDS.length);
    }, 2600);
    return () => clearInterval(timer);
  }, []);

  const modelList = Array.isArray(models) ? models : [];
  const selectedModelMeta =
    modelList.find((m) => m.id === selectedModel) || (modelList.length > 0 ? modelList[0] : null);

  const handleSendMessage = async (promptText: string) => {
    if (!isAuthenticated) {
      toast.info("Please sign in to continue");
      setAuthModalOpen(true);
      return;
    }

    const newThreadId = await createNewThread(promptText, selectedModel);
    router.navigate({ to: `/c/${newThreadId}` });
    sendMessage(promptText, newThreadId);
  };

  return (
    <>
      <div className="flex flex-col h-full justify-center pb-6">
        <div className="flex-1 flex flex-col items-center justify-center text-center px-4 pt-12">
          <div className="relative inline-block mb-8 select-none">
            <div className="ambient-mesh-glow absolute -inset-6 bg-gradient-to-r from-emerald-500/10 via-cyan-500/10 to-indigo-500/10 rounded-full blur-2xl pointer-events-none" />
            <h1 className="relative flex flex-wrap items-center justify-center gap-x-2.5 gap-y-1.5 text-3xl tracking-tight font-medium text-white">
              <span>
                {isAuthenticated && user?.full_name
                  ? `Hey ${user.full_name.split(" ")[0] || "there"}, what should we`
                  : "What should we"}
              </span>
              <span className="live-mesh-badge text-zinc-950 font-bold px-2.5 py-0.5 rounded-lg inline-flex items-center justify-center shadow-lg overflow-hidden align-middle">
                <AnimatePresence mode="wait">
                  <motion.span
                    key={ROTATING_WORDS[wordIndex]}
                    initial={{ y: 14, opacity: 0, filter: "blur(4px)" }}
                    animate={{ y: 0, opacity: 1, filter: "blur(0px)" }}
                    exit={{ y: -14, opacity: 0, filter: "blur(4px)" }}
                    transition={{ duration: 0.28, ease: "easeInOut" }}
                    className="inline-block"
                  >
                    {ROTATING_WORDS[wordIndex]}
                  </motion.span>
                </AnimatePresence>
              </span>
              <span>?</span>
            </h1>
          </div>
          <ChatPromptInput onSendMessage={handleSendMessage} disabled={isStreaming} />
        </div>
      </div>

      <AddCredentialModal
        open={credModalOpen}
        onOpenChange={setCredModalOpen}
        defaultProvider={selectedModelMeta?.provider}
      />
      <AuthModal open={authModalOpen} onOpenChange={setAuthModalOpen} initialMode="login" />
    </>
  );
};

export default HomePage;
