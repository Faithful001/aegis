import { useState, useRef, useEffect, useCallback } from "react";
import { toast } from "sonner";
import { RawPcmRecorder } from "../utils/audio";
import { transcribeAudioPcm } from "../services/whisper";

interface UseSpeechRecognitionOptions {
  onResult: (text: string) => void;
}

export function useSpeechRecognition({ onResult }: UseSpeechRecognitionOptions) {
  const [isListening, setIsListening] = useState(false);
  const [isTranscribing, setIsTranscribing] = useState(false);
  
  const isListeningRef = useRef(false);
  const isStartingRef = useRef(false);
  const recorderRef = useRef<RawPcmRecorder | null>(null);
  const baseTextRef = useRef<string>("");
  const downloadToastIdRef = useRef<string | number | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  const stopListening = useCallback(async () => {
    console.log("[SpeechRecognition] Stopping recording...");
    isStartingRef.current = false;
    isListeningRef.current = false;
    setIsListening(false);
    setIsTranscribing(true);

    // Yield to React so it can flush the state above and paint the spinner
    // BEFORE we block the JS thread with PCM resampling/normalization
    await Promise.resolve();

    let audioData: Float32Array | null = null;
    try {
      if (recorderRef.current) {
        audioData = recorderRef.current.stop();
        recorderRef.current = null;
      }
    } catch (err) {
      console.error("[SpeechRecognition] Error stopping recorder:", err);
      recorderRef.current = null;
    }

    if (!audioData || audioData.length < 1600) {
      console.log("[SpeechRecognition] Audio too short:", audioData?.length);
      setIsTranscribing(false);
      return;
    }

    const abortController = new AbortController();
    abortControllerRef.current = abortController;

    try {
      console.log("[SpeechRecognition] Transcribing audio...");
      const text = await transcribeAudioPcm(audioData, (progress) => {
        if (progress && progress.status === "progress" && progress.total) {
          const percent = Math.round((progress.loaded / progress.total) * 100);
          if (!downloadToastIdRef.current) {
            downloadToastIdRef.current = toast.loading(
              `Downloading offline Whisper model (${percent}%)...`
            );
          } else {
            toast.loading(`Downloading offline Whisper model (${percent}%)...`, {
              id: downloadToastIdRef.current,
            });
          }
        } else if (progress && progress.status === "done" && downloadToastIdRef.current) {
          toast.dismiss(downloadToastIdRef.current);
          downloadToastIdRef.current = null;
        }
      }, abortController.signal);

      if (downloadToastIdRef.current) {
        toast.dismiss(downloadToastIdRef.current);
        downloadToastIdRef.current = null;
      }

      if (text) {
        const base = baseTextRef.current;
        const separator = base && !base.endsWith(" ") ? " " : "";
        const newText = base ? `${base}${separator}${text}` : text;
        onResult(newText);
      } else {
        toast.info("No words detected in audio. Please try again.");
      }
    } catch (err: any) {
      if (downloadToastIdRef.current) {
        toast.dismiss(downloadToastIdRef.current);
        downloadToastIdRef.current = null;
      }
      if (err?.name === "AbortError") {
        console.log("[SpeechRecognition] Transcription aborted by user.");
      } else {
        console.error("[SpeechRecognition] Transcription error:", err);
        toast.error(err?.message || "Speech transcription failed.");
      }
    } finally {
      setIsTranscribing(false);
      abortControllerRef.current = null;
    }
  }, [onResult]);

  const startListening = useCallback(
    async (currentInput = "") => {
      if (isListeningRef.current || isStartingRef.current) {
        stopListening();
        return;
      }

      if (typeof navigator === "undefined" || !navigator.mediaDevices?.getUserMedia) {
        toast.error("Microphone recording is not supported in this browser.");
        return;
      }

      isStartingRef.current = true;
      baseTextRef.current = currentInput;

      try {
        const recorder = new RawPcmRecorder();
        recorderRef.current = recorder;
        
        await recorder.start();

        // If user called stop while we were waiting for microphone permission
        if (!isStartingRef.current) {
          recorder.stop();
          recorderRef.current = null;
          return;
        }

        isStartingRef.current = false;
        isListeningRef.current = true;
        setIsListening(true);
      } catch (err: any) {
        console.error("[SpeechRecognition] Start recording failed:", err);
        isStartingRef.current = false;
        isListeningRef.current = false;
        setIsListening(false);
        recorderRef.current = null;

        if (err.name === "NotAllowedError" || err.name === "PermissionDeniedError") {
          toast.error("Microphone permission was denied. Please allow microphone access in your browser.");
        } else {
          toast.error("Could not access microphone.");
        }
      }
    },
    [stopListening]
  );

  const cancelListening = useCallback(() => {
    console.log("[SpeechRecognition] Canceling voice input...");

    // Immediately abort any in-flight transcription first
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }

    // Immediately stop microphone hardware
    if (recorderRef.current) {
      try {
        recorderRef.current.stop();
      } catch {
        // Ignore
      }
      recorderRef.current = null;
    }

    // Dismiss any model download toast
    if (downloadToastIdRef.current) {
      toast.dismiss(downloadToastIdRef.current);
      downloadToastIdRef.current = null;
    }

    // Reset all state flags synchronously
    isStartingRef.current = false;
    isListeningRef.current = false;
    setIsListening(false);
    setIsTranscribing(false);

    // Restore input to what it was before voice session started
    onResult(baseTextRef.current);
  }, [onResult]);

  const toggleListening = useCallback(
    (currentInput = "") => {
      if (isListeningRef.current || isStartingRef.current || isTranscribing) {
        stopListening();
      } else {
        startListening(currentInput);
      }
    },
    [startListening, stopListening, isTranscribing]
  );

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      isStartingRef.current = false;
      isListeningRef.current = false;
      if (recorderRef.current) {
        try {
          recorderRef.current.stop();
        } catch {
          // Ignore
        }
        recorderRef.current = null;
      }
      if (downloadToastIdRef.current) {
        toast.dismiss(downloadToastIdRef.current);
      }
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, []);

  return {
    isListening,
    isTranscribing,
    isSupported: true,
    startListening,
    stopListening,
    cancelListening,
    toggleListening,
  };
}
