import { pipeline, env } from "@xenova/transformers";

// Configure transformers.js for client execution
env.allowLocalModels = false;
env.useBrowserCache = true;

// Prevent hanging on SharedArrayBuffer in browsers without cross-origin isolation
if (env.backends?.onnx?.wasm) {
  env.backends.onnx.wasm.numThreads = 1;
}

const MODEL_ID = "Xenova/whisper-base.en";
const CACHE_STORAGE_KEY = "aegis_whisper_base_v1";

export type WhisperModelStatus = "idle" | "loading" | "ready" | "error";

let modelStatus: WhisperModelStatus = "idle";
let transcriberPromise: Promise<any> | null = null;
const statusListeners = new Set<(status: WhisperModelStatus) => void>();

function setStatus(status: WhisperModelStatus) {
  modelStatus = status;
  statusListeners.forEach((listener) => {
    try {
      listener(status);
    } catch {
      // Ignore listener error
    }
  });
}

export function subscribeModelStatus(listener: (status: WhisperModelStatus) => void): () => void {
  statusListeners.add(listener);
  listener(modelStatus);
  return () => {
    statusListeners.delete(listener);
  };
}

export function isModelReady(): boolean {
  return modelStatus === "ready";
}

export function isModelCached(): boolean {
  try {
    return localStorage.getItem(CACHE_STORAGE_KEY) === "true";
  } catch {
    return false;
  }
}

/**
 * Initializes and caches the Whisper model.
 * Reuses the existing promise if already loading or ready.
 */
export async function getTranscriber(onProgress?: (progress: any) => void) {
  if (transcriberPromise) {
    return transcriberPromise;
  }

  setStatus("loading");
  const alreadyCached = isModelCached();
  console.log(
    `[Whisper] Initializing ${MODEL_ID} (${alreadyCached ? "cached in browser" : "first-time download"})...`
  );

  transcriberPromise = pipeline("automatic-speech-recognition", MODEL_ID, {
    progress_callback: onProgress,
    quantized: true,
  })
    .then((transcriber) => {
      setStatus("ready");
      try {
        localStorage.setItem(CACHE_STORAGE_KEY, "true");
      } catch {
        // Ignore quota/private mode errors
      }
      console.log("[Whisper] Model is ready for speech-to-text.");
      return transcriber;
    })
    .catch((err) => {
      transcriberPromise = null;
      setStatus("error");
      console.error("[Whisper] Model initialization failure:", err);
      throw err;
    });

  return transcriberPromise;
}

/**
 * Preloads the Whisper model in the background once the application mounts.
 * Runs non-blockingly during browser idle time.
 */
export function preloadWhisperModel(): void {
  if (modelStatus === "ready" || modelStatus === "loading") {
    return;
  }

  const runPreload = () => {
    // If browser is offline and model was never cached, skip background download until online
    if (typeof navigator !== "undefined" && !navigator.onLine && !isModelCached()) {
      console.log("[Whisper] Offline and model not cached; postponing background preload.");
      return;
    }

    console.log("[Whisper] Starting non-blocking background model preload...");
    getTranscriber().catch((err) => {
      // Catch background errors quietly so app boot is never interrupted
      console.warn("[Whisper] Background preload error (will retry on mic click):", err);
    });
  };

  // Schedule during browser idle time so main UI render and initial network requests take priority
  if (typeof window !== "undefined" && "requestIdleCallback" in window) {
    (window as any).requestIdleCallback(() => runPreload(), { timeout: 3000 });
  } else {
    setTimeout(runPreload, 1500);
  }
}

export async function transcribeAudioPcm(
  audio: Float32Array,
  onProgress?: (progress: any) => void,
  signal?: AbortSignal
): Promise<string> {
  console.log("[Whisper] Transcribing PCM samples:", audio.length);

  if (signal?.aborted) {
    throw new DOMException("Aborted", "AbortError");
  }

  const abortPromise = new Promise<never>((_, reject) => {
    if (signal) {
      if (signal.aborted) {
        reject(new DOMException("Aborted", "AbortError"));
      } else {
        signal.addEventListener("abort", () => {
          reject(new DOMException("Aborted", "AbortError"));
        });
      }
    }
  });

  // Set a 35-second safety timeout so it never hangs indefinitely
  const timeoutPromise = new Promise<never>((_, reject) => {
    setTimeout(() => {
      reject(new Error("Transcription timed out. Check your network connection."));
    }, 35000);
  });

  const runTranscription = async (): Promise<string> => {
    if (signal?.aborted) throw new DOMException("Aborted", "AbortError");

    const transcriber = await getTranscriber(onProgress);

    if (signal?.aborted) throw new DOMException("Aborted", "AbortError");

    const result = await transcriber(audio, {
      return_timestamps: false,
      chunk_length_s: 30,
      // Beam search (evaluates multiple token candidate paths instead of greedy 1-best)
      num_beams: 2,
      // High-sensitivity decoding settings:
      no_speech_threshold: 0.15,
      condition_on_previous_text: false,
      temperature: 0.0,
    });

    if (signal?.aborted) throw new DOMException("Aborted", "AbortError");

    let text = typeof result?.text === "string" ? result.text.trim() : "";
    // Filter out Whisper silence hallucination tokens
    text = text.replace(/\[BLANK_AUDIO\]|\[MUSIC\]|\[APPLAUSE\]/gi, "").trim();
    console.log("[Whisper] Transcribed text:", text);
    return text;
  };

  return Promise.race([runTranscription(), timeoutPromise, abortPromise]);
}
