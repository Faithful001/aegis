import React, { useState, useRef, useEffect } from "react";
import {
  CaretDownIcon,
  CheckIcon,
  CircleNotchIcon,
  FileTextIcon,
  MicrophoneIcon,
  MicrophoneSlashIcon,
  PaperPlaneRightIcon,
  PlusIcon,
  XIcon,
} from "@phosphor-icons/react";
import { motion, AnimatePresence } from "framer-motion";
import { useAegis } from "../../context/AegisContext";
import { useLocation } from "@tanstack/react-router";
import { useSpeechRecognition } from "../../hooks/useSpeechRecognition";
import { uploadApi } from "../../api/upload";
import { AttachmentItem } from "../../types/api";
import { toast } from "sonner";

const ALLOWED_PHOTO_EXTS = [".png", ".jpg", ".jpeg", ".webp", ".gif"];
const ALLOWED_PHOTO_MIMES = ["image/png", "image/jpeg", "image/webp", "image/gif"];

const ALLOWED_DOC_EXTS = [
  ".pdf", ".doc", ".docx", ".txt", ".md", ".csv", ".json",
  ".py", ".js", ".ts", ".tsx", ".jsx", ".go", ".rs",
  ".c", ".cpp", ".h", ".hpp", ".html", ".css",
  ".yaml", ".yml", ".sh", ".sql",
];// Combined accept string for a single input that handles both images and documents
const COMBINED_ACCEPT = [
  "image/png,image/jpeg,image/webp,image/gif",
  ".pdf,.doc,.docx,.txt,.md,.csv,.json",
  ".py,.js,.ts,.tsx,.jsx,.go,.rs,.c,.cpp,.h,.hpp",
  ".html,.css,.yaml,.yml,.sh,.sql",
].join(",");

interface ChatPromptInputProps {
  onSendMessage: (prompt: string, attachments?: AttachmentItem[]) => void;
  disabled?: boolean;
}

export const ChatPromptInput: React.FC<ChatPromptInputProps> = ({ onSendMessage, disabled }) => {
  const [input, setInput] = useState("");
  const [modelDropdownOpen, setModelDropdownOpen] = useState(false);
  const [attachments, setAttachments] = useState<AttachmentItem[]>([]);
  const [isUploading, setIsUploading] = useState(false);

  const dropdownRef = useRef<HTMLDivElement>(null);
  const combinedInputRef = useRef<HTMLInputElement>(null);

  const { selectedModel, setSelectedModel, models } = useAegis();

  // Local flag that flips synchronously on click, before the async hook work begins
  const [isStopping, setIsStopping] = useState(false);

  const { isListening, isTranscribing, startListening, stopListening, cancelListening } =
    useSpeechRecognition({
      onResult: (text) => {
        setInput(text);
        setIsStopping(false);
      },
    });

  // Once hook finishes transcribing (success or error), clear local stopping flag
  useEffect(() => {
    if (!isTranscribing) {
      setIsStopping(false);
    }
  }, [isTranscribing]);

  const showLoader = isStopping || isTranscribing;

  const { pathname } = useLocation();

  const modelList = Array.isArray(models) ? models : [];
  const currentModelObj =
    modelList.find((m) => m.id === selectedModel) || (modelList.length > 0 ? modelList[0] : null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setModelDropdownOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  // Auto-detect file type from extension/mime and validate
  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (!files || files.length === 0) return;

    const fileList = Array.from(files);
    const photoFiles: File[] = [];
    const docFiles: File[] = [];

    for (const file of fileList) {
      const ext = "." + file.name.split(".").pop()?.toLowerCase();
      const isPhoto = ALLOWED_PHOTO_EXTS.includes(ext) || ALLOWED_PHOTO_MIMES.includes(file.type);
      const isDoc = ALLOWED_DOC_EXTS.includes(ext);

      if (isPhoto) {
        if (file.size > 15 * 1024 * 1024) {
          toast.error(`"${file.name}" exceeds the 15MB photo size limit.`);
          continue;
        }
        photoFiles.push(file);
      } else if (isDoc) {
        if (file.size > 30 * 1024 * 1024) {
          toast.error(`"${file.name}" exceeds the 30MB file size limit.`);
          continue;
        }
        docFiles.push(file);
      } else {
        toast.error(`"${file.name}" is not a supported format.`);
      }
    }

    if (photoFiles.length === 0 && docFiles.length === 0) {
      e.target.value = "";
      return;
    }

    setIsUploading(true);
    try {
      const results: AttachmentItem[] = [];

      if (photoFiles.length > 0) {
        const uploaded = await uploadApi.uploadMultiple(photoFiles, "photo");
        results.push(...uploaded.map((u) => ({
          id: u.id,
          filename: u.filename,
          mime_type: u.mime_type,
          file_type: u.file_type,
          size: u.size,
          url: u.url,
          content: u.content,
        })));
      }

      if (docFiles.length > 0) {
        const uploaded = await uploadApi.uploadMultiple(docFiles, "file");
        results.push(...uploaded.map((u) => ({
          id: u.id,
          filename: u.filename,
          mime_type: u.mime_type,
          file_type: u.file_type,
          size: u.size,
          url: u.url,
          content: u.content,
        })));
      }

      setAttachments((prev) => [...prev, ...results]);
      toast.success(
        results.length === 1
          ? `Attached "${results[0].filename}"`
          : `Attached ${results.length} files`
      );
    } catch (err: any) {
      toast.error(err?.response?.data?.message || "Failed to upload files");
    } finally {
      setIsUploading(false);
    }

    e.target.value = "";
  };

  const removeAttachment = (indexToRemove: number) => {
    setAttachments((prev) => prev.filter((_, idx) => idx !== indexToRemove));
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      submit();
    }
  };

  const submit = () => {
    if ((!input.trim() && attachments.length === 0) || disabled || isUploading) return;
    if (isListening) {
      stopListening();
    }
    onSendMessage(input.trim(), attachments);
    setInput("");
    setAttachments([]);
  };

  return (
    <div className="w-full max-w-3xl mx-auto px-3 sm:px-4">
      <div className="relative bg-surface-card/90 border border-surface-border rounded-3xl p-3 sm:p-3.5 shadow-glow-card transition-all focus-within:border-zinc-600 focus-within:ring-1 focus-within:ring-zinc-600">
        {/* Attachment preview chips */}
        {attachments.length > 0 && (
          <div className="flex flex-wrap items-center gap-2 mb-2 px-1">
            {attachments.map((att, idx) => (
              <div
                key={att.id || idx}
                className="group relative flex items-center gap-2 bg-zinc-800/90 border border-zinc-700/80 rounded-xl px-2.5 py-1.5 text-xs text-zinc-200 shadow-sm transition-all hover:border-zinc-600"
              >
                {att.file_type === "image" ? (
                  <div className="w-6 h-6 rounded-lg overflow-hidden bg-zinc-900 shrink-0 border border-zinc-700/50">
                    <img
                      src={
                        att.url ||
                        (att.content ? `data:${att.mime_type};base64,${att.content}` : "")
                      }
                      alt={att.filename}
                      className="w-full h-full object-cover"
                    />
                  </div>
                ) : (
                  <FileTextIcon className="w-4 h-4 text-emerald-400 shrink-0" />
                )}
                <div className="flex flex-col min-w-0 max-w-[130px] sm:max-w-[180px]">
                  <span className="truncate text-xs font-medium">{att.filename}</span>
                  {att.size && (
                    <span className="text-[10px] text-zinc-400">
                      {(att.size / 1024).toFixed(0)} KB
                    </span>
                  )}
                </div>
                <button
                  type="button"
                  onClick={() => removeAttachment(idx)}
                  className="p-1 rounded-full text-zinc-400 hover:text-red-400 hover:bg-zinc-700/60 transition-colors ml-0.5"
                  title="Remove attachment"
                  aria-label="Remove attachment"
                >
                  <XIcon className="w-3.5 h-3.5" />
                </button>
              </div>
            ))}

            {isUploading && (
              <div className="flex items-center gap-1.5 text-xs text-zinc-400 px-2 py-1 bg-zinc-800/40 rounded-xl border border-zinc-700/40">
                <CircleNotchIcon className="w-3.5 h-3.5 animate-spin text-zinc-300" />
                <span>Uploading...</span>
              </div>
            )}
          </div>
        )}

        {/* Main Textarea */}
        <textarea
          rows={2}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={
            showLoader
              ? "Transcribing with Whisper..."
              : isListening
                ? "Listening... Click mic to stop"
                : pathname === "/"
                  ? "How can I help you today?"
                  : "Write a message..."
          }
          disabled={disabled}
          className="w-full bg-transparent text-sm text-zinc-100 placeholder:text-zinc-500 focus:outline-none resize-none px-2 py-1 min-h-[48px]"
        />

        {/* Bottom Bar Controls inside prompt box */}
        <div className="flex items-center justify-between pt-2 border-t border-surface-border/40 px-1 gap-1">
          {/* Left: single "Add photos/files" button */}
          <div>
            {/* Single hidden input that accepts both images and documents */}
            <input
              ref={combinedInputRef}
              type="file"
              accept={COMBINED_ACCEPT}
              multiple
              className="hidden"
              onChange={handleFileSelect}
            />
            <button
              type="button"
              onClick={() => combinedInputRef.current?.click()}
              className="p-1.5 rounded-full text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors shrink-0"
              title="Add photos/files"
              aria-label="Add photos/files"
            >
              <PlusIcon className="w-4 h-4" />
            </button>
          </div>

          {/* Right Action Icons: Model Selector Dropdown, Mic Icon, Send */}
          <div className="flex items-center gap-1.5 sm:gap-2 min-w-0">
            {/* Model Selector Dropdown inside prompt bar */}
            <div className="relative" ref={dropdownRef}>
              <button
                type="button"
                onClick={() => setModelDropdownOpen(!modelDropdownOpen)}
                className="flex items-center gap-1.5 px-2.5 sm:px-3 py-1 rounded-full text-xs font-medium bg-zinc-800/80 hover:bg-zinc-800 text-zinc-200 border border-zinc-700/60 transition-all max-w-[125px] sm:max-w-[200px]"
              >
                <span className="truncate">
                  {currentModelObj ? currentModelObj.name : "Select Model"}
                </span>
                <CaretDownIcon
                  className={`w-3.5 h-3.5 text-zinc-400 transition-transform duration-200 shrink-0 ${
                    modelDropdownOpen ? "rotate-180" : ""
                  }`}
                />
              </button>

              <AnimatePresence>
                {modelDropdownOpen && (
                  <motion.div
                    initial={{ opacity: 0, scale: 0.95, y: 8 }}
                    animate={{ opacity: 1, scale: 1, y: 0 }}
                    exit={{ opacity: 0, scale: 0.95, y: 8 }}
                    transition={{ duration: 0.15, ease: "easeOut" }}
                    className="absolute right-0 bottom-full mb-2 w-72 max-w-[calc(100vw-3rem)] max-h-80 overflow-y-auto bg-surface border border-surface-border rounded-2xl p-1.5 shadow-2xl z-50 space-y-1"
                  >
                    <div className="px-3 py-1.5 text-[10px] uppercase font-bold text-zinc-400 tracking-wider flex items-center justify-between">
                      <span>Select Model</span>
                      <span className="text-[10px] lowercase text-zinc-500 font-normal">
                        {models.length} available
                      </span>
                    </div>
                    {models.length === 0 ? (
                      <div className="px-3 py-4 text-center text-xs text-zinc-400 space-y-1">
                        <p className="font-medium text-zinc-300">No models discovered</p>
                        <p className="text-[11px] text-zinc-500">
                          Add a provider API key in Settings to discover models.
                        </p>
                      </div>
                    ) : (
                      models.map((model) => (
                        <button
                          key={model.id}
                          type="button"
                          onClick={() => {
                            setSelectedModel(model.id);
                            setModelDropdownOpen(false);
                          }}
                          className={`w-full text-left px-3 py-2 rounded-xl transition-all ${
                            selectedModel === model.id
                              ? "bg-zinc-800 text-white font-medium"
                              : "text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200"
                          }`}
                        >
                          <div className="flex items-center justify-between text-xs font-semibold">
                            <span className="flex items-center gap-1.5 truncate">
                              <span className="truncate">{model.name}</span>
                              {selectedModel === model.id && (
                                <CheckIcon className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                              )}
                            </span>
                            <div className="flex items-center gap-1 shrink-0 ml-2">
                              <span className="text-[9px] uppercase px-1 py-0.5 rounded bg-zinc-800/80 text-zinc-400 border border-zinc-700/50">
                                {model.provider}
                              </span>
                              {model.fast && (
                                <span className="text-[10px] text-emerald-400 font-normal">
                                  Fast
                                </span>
                              )}
                            </div>
                          </div>
                        </button>
                      ))
                    )}
                  </motion.div>
                )}
              </AnimatePresence>
            </div>

            {/* Cancel voice input button */}
            <AnimatePresence>
              {(isListening || showLoader) && (
                <motion.button
                  key="cancel-voice-btn"
                  initial={{ opacity: 0, scale: 0.8 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.8 }}
                  transition={{ duration: 0.15 }}
                  type="button"
                  onClick={() => {
                    setIsStopping(false);
                    cancelListening();
                  }}
                  title="Cancel voice input"
                  aria-label="Cancel voice input"
                  className="p-1.5 rounded-full text-zinc-400 hover:text-red-400 hover:bg-red-500/10 transition-colors shrink-0"
                >
                  <XIcon className="w-4 h-4" />
                </motion.button>
              )}
            </AnimatePresence>

            <button
              type="button"
              onClick={() => {
                if (isListening) {
                  // Flip local flag synchronously — spinner renders THIS render tick
                  setIsStopping(true);
                  stopListening();
                } else if (!showLoader) {
                  startListening(input);
                }
              }}
              disabled={disabled || showLoader}
              title={isListening ? "Click to finish recording" : "Use speech to text"}
              aria-label={isListening ? "Stop recording" : "Start speech to text"}
              className={`p-1.5 rounded-full transition-all relative ${
                isListening
                  ? "text-red-400 bg-red-500/20 ring-1 ring-red-500/50 hover:bg-red-500/30"
                  : showLoader
                    ? "text-zinc-200 bg-transparent cursor-default"
                    : "text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800"
              }`}
            >
              {showLoader ? (
                <div className="w-4 h-4 border-2 border-zinc-200 border-t-transparent rounded-full animate-spin bg-transparent" />
              ) : isListening ? (
                <>
                  <span className="absolute -top-0.5 -right-0.5 flex h-2 w-2">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75" />
                    <span className="relative inline-flex rounded-full h-2 w-2 bg-red-500" />
                  </span>
                  <MicrophoneSlashIcon className="w-4 h-4 animate-pulse" />
                </>
              ) : (
                <MicrophoneIcon className="w-4 h-4" />
              )}
            </button>

            <button
              type="button"
              onClick={submit}
              disabled={disabled || (!input.trim() && attachments.length === 0) || isUploading}
              className="w-8 h-8 rounded-full bg-white hover:bg-white text-black flex items-center justify-center transition-all disabled:opacity-40 disabled:hover:bg-white"
            >
              <PaperPlaneRightIcon className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
