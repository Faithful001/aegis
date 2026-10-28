import { pipeline, env } from "@xenova/transformers";

// Configure transformers.js for in-browser client execution
env.allowLocalModels = false;
env.useBrowserCache = true;

// Prevent hanging on SharedArrayBuffer in browsers without cross-origin isolation
if (env.backends?.onnx?.wasm) {
  env.backends.onnx.wasm.numThreads = 1;
}

let transcriberPromise: Promise<any> | null = null;

async function getTranscriber(onProgress?: (progress: any) => void) {
  if (!transcriberPromise) {
    console.log("[Whisper Worker] Initializing pipeline Xenova/whisper-tiny.en...");
    transcriberPromise = pipeline(
      "automatic-speech-recognition",
      "Xenova/whisper-tiny.en",
      {
        progress_callback: onProgress,
        quantized: true,
      }
    ).catch((err) => {
      transcriberPromise = null;
      throw err;
    });
  }
  return transcriberPromise;
}

self.addEventListener("message", async (e: MessageEvent) => {
  const { type, audio, id } = e.data;

  if (type === "TRANSCRIBE") {
    try {
      console.log("[Whisper Worker] Transcribing audio sample count:", audio?.length);
      const transcriber = await getTranscriber((progress) => {
        self.postMessage({ type: "PROGRESS", progress, id });
      });

      console.log("[Whisper Worker] Running inference...");
      const result = await transcriber(audio, {
        chunk_length_s: 30,
        stride_length_s: 5,
        language: "english",
        task: "transcribe",
      });

      const text = typeof result?.text === "string" ? result.text.trim() : "";
      console.log("[Whisper Worker] Result:", text);
      self.postMessage({ type: "SUCCESS", text, id });
    } catch (err: any) {
      console.error("[Whisper Worker] Error during transcription:", err);
      self.postMessage({ type: "ERROR", error: err?.message || String(err), id });
    }
  }
});
