/**
 * Resamples Float32Array PCM audio to 16kHz mono.
 */
export function resampleTo16kHz(audioData: Float32Array, origSampleRate: number): Float32Array {
  if (origSampleRate === 16000 || audioData.length === 0) return audioData;
  const ratio = origSampleRate / 16000;
  const newLength = Math.round(audioData.length / ratio);
  if (!newLength) return new Float32Array(0);

  const result = new Float32Array(newLength);
  for (let i = 0; i < newLength; i++) {
    const origIndex = i * ratio;
    const lower = Math.floor(origIndex);
    const upper = Math.min(lower + 1, audioData.length - 1);
    const frac = origIndex - lower;
    result[i] = audioData[lower] * (1 - frac) + audioData[upper] * frac;
  }
  return result;
}

/**
 * 1st-order IIR high-pass filter (cutoff ~80Hz at 16kHz) to filter out
 * desk vibration, fan noise, and DC air rumble, preserving vocal formants.
 */
export function removeLowFrequencyRumble(audio: Float32Array): Float32Array {
  if (audio.length === 0) return audio;
  const alpha = 0.969;
  let prevIn = audio[0];
  let prevOut = 0;

  for (let i = 0; i < audio.length; i++) {
    const currIn = audio[i];
    const currOut = alpha * (prevOut + currIn - prevIn);
    prevIn = currIn;
    prevOut = currOut;
    audio[i] = currOut;
  }
  return audio;
}

/**
 * Removes rumble and DC offset and applies digital gain normalization so peak amplitude is ~0.92.
 * This ensures Whisper receives loud, crisp audio even when speaking softly or at a distance.
 */
export function normalizeAudio(audio: Float32Array): Float32Array {
  if (audio.length === 0) return audio;

  // 1. High-pass filter to strip non-vocal rumble (< 80 Hz)
  removeLowFrequencyRumble(audio);

  // 2. Remove DC offset
  let sum = 0;
  for (let i = 0; i < audio.length; i++) {
    sum += audio[i];
  }
  const mean = sum / audio.length;

  let maxAbs = 0;
  let sumSq = 0;
  for (let i = 0; i < audio.length; i++) {
    const val = audio[i] - mean;
    audio[i] = val;
    const abs = Math.abs(val);
    if (abs > maxAbs) {
      maxAbs = abs;
    }
    sumSq += val * val;
  }

  const rms = Math.sqrt(sumSq / audio.length);
  console.log(`[Audio Energy] Raw RMS: ${rms.toFixed(5)}, Raw Peak: ${maxAbs.toFixed(5)}`);

  // 3. Normalize peak amplitude to 0.92 for maximum Whisper sensitivity
  if (maxAbs > 0.0001) {
    const targetPeak = 0.92;
    const boost = Math.min(targetPeak / maxAbs, 80.0); // Allow up to 80x amplification for soft voices
    console.log(
      `[Audio Normalization] Applying ${boost.toFixed(1)}x digital boost (Target peak: ${targetPeak}).`
    );
    for (let i = 0; i < audio.length; i++) {
      audio[i] *= boost;
    }
  }

  return audio;
}

/**
 * Captures raw Float32Array PCM samples directly from the microphone
 * via AudioContext, keeping the audio graph active and preventing
 * Chromium from optimizing out silent nodes.
 */
export class RawPcmRecorder {
  private stream: MediaStream | null = null;
  private audioContext: AudioContext | null = null;
  private processor: ScriptProcessorNode | null = null;
  private source: MediaStreamAudioSourceNode | null = null;
  private gainNode: GainNode | null = null;
  private pcmChunks: Float32Array[] = [];

  async start(): Promise<void> {
    this.pcmChunks = [];
    const stream = await navigator.mediaDevices.getUserMedia({
      audio: {
        channelCount: 1,
        echoCancellation: true,
        noiseSuppression: false, // Do not mute soft speech
        autoGainControl: true,   // Maximize hardware microphone sensitivity
      },
    });
    this.stream = stream;

    const AudioCtx =
      window.AudioContext ||
      (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;

    // Prefer native 16000Hz sampling if supported by the browser hardware pipeline
    let ctx: AudioContext;
    try {
      ctx = new AudioCtx({ sampleRate: 16000 });
    } catch {
      ctx = new AudioCtx();
    }
    this.audioContext = ctx;

    if (ctx.state === "suspended") {
      await ctx.resume();
    }

    const source = ctx.createMediaStreamSource(stream);

    // Buffer size 4096 gives ~92ms audio slices at 44.1kHz or ~256ms at 16kHz
    const processor = ctx.createScriptProcessor(4096, 1, 1);
    const gainNode = ctx.createGain();

    // CRITICAL: Gain MUST be a tiny non-zero value (1e-6, ~ -120dB).
    // If gain is exactly 0, Chromium's silence optimizer prunes the audio branch
    // and completely stops calling onaudioprocess!
    gainNode.gain.value = 0.000001;

    processor.onaudioprocess = (e) => {
      const channelData = e.inputBuffer.getChannelData(0);
      this.pcmChunks.push(new Float32Array(channelData));
    };

    source.connect(processor);
    processor.connect(gainNode);
    gainNode.connect(ctx.destination);

    this.source = source;
    this.processor = processor;
    this.gainNode = gainNode;

    // Retain global reference so V8 GC does not collect active processor
    (window as unknown as { __aegis_active_processor?: unknown }).__aegis_active_processor = processor;
  }

  stop(): Float32Array {
    try {
      delete (window as unknown as { __aegis_active_processor?: unknown }).__aegis_active_processor;

      if (this.stream) {
        this.stream.getTracks().forEach((track) => {
          track.stop();
          track.enabled = false;
        });
        this.stream = null;
      }

      if (this.processor) {
        this.processor.onaudioprocess = null;
        this.processor.disconnect();
        this.processor = null;
      }
      if (this.gainNode) {
        this.gainNode.disconnect();
        this.gainNode = null;
      }
      if (this.source) {
        this.source.disconnect();
        this.source = null;
      }

      const sampleRate = this.audioContext?.sampleRate || 16000;
      if (this.audioContext && this.audioContext.state !== "closed") {
        this.audioContext.close().catch(() => {});
        this.audioContext = null;
      }

      const totalLength = this.pcmChunks.reduce((acc, c) => acc + c.length, 0);
      console.log("[RawPcmRecorder] Captured total samples:", totalLength, "at native rate:", sampleRate);
      const merged = new Float32Array(totalLength);
      let offset = 0;
      for (const chunk of this.pcmChunks) {
        merged.set(chunk, offset);
        offset += chunk.length;
      }
      this.pcmChunks = [];

      const resampled = resampleTo16kHz(merged, sampleRate);
      return normalizeAudio(resampled);
    } catch (err) {
      console.error("[RawPcmRecorder] Stop error:", err);
      return new Float32Array(0);
    }
  }
}
