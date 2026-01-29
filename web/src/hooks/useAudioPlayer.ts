import React, { useRef, useState, useCallback, useEffect } from 'react';

export interface AudioPlayerState {
  isPlaying: boolean;
  isLoading: boolean;
  currentTime: number;
  duration: number;
  playbackRate: number;
  error: string | null;
  // Real-time analysis data from AudioWorklet
  rms: number;
  peak: number;
  waveformData: Float32Array | null;
  // Streaming status
  isStreaming: boolean;
  bufferedTime: number;
  // AudioWorklet status
  workletReady: boolean;
}

export interface AudioPlayerControls {
  play: () => Promise<void>;
  pause: () => void;
  stop: () => void;
  seek: (time: number) => void;
  seekToPosition: (position: number) => void;
  setPlaybackRate: (rate: number) => void;
  loadTrack: (url: string) => Promise<void>;
  loadStreamingTrack: (url: string) => Promise<void>;
  getAnalyserData: () => Uint8Array | null;
}

// Singleton for AudioWorklet registration
let workletRegistered = false;

export function useAudioPlayer(): [AudioPlayerState, AudioPlayerControls] {
  const audioContextRef = useRef<AudioContext | null>(null);
  const sourceNodeRef = useRef<AudioBufferSourceNode | null>(null);
  const audioBufferRef = useRef<AudioBuffer | null>(null);
  const gainNodeRef = useRef<GainNode | null>(null);
  const analyserNodeRef = useRef<AnalyserNode | null>(null);
  const workletNodeRef = useRef<AudioWorkletNode | null>(null);
  const startTimeRef = useRef<number>(0);
  const pausedAtRef = useRef<number>(0);
  const animationFrameRef = useRef<number>(0);
  const analyserDataRef = useRef<Uint8Array<ArrayBuffer> | null>(null);

  const [state, setState] = useState<AudioPlayerState>({
    isPlaying: false,
    isLoading: false,
    currentTime: 0,
    duration: 0,
    playbackRate: 1,
    error: null,
    rms: 0,
    peak: 0,
    waveformData: null,
    isStreaming: false,
    bufferedTime: 0,
    workletReady: false,
  });

  // Initialize audio context with AudioWorklet
  const getAudioContext = useCallback(async () => {
    if (!audioContextRef.current) {
      audioContextRef.current = new AudioContext({ sampleRate: 48000 });

      // Create gain node
      gainNodeRef.current = audioContextRef.current.createGain();

      // Create analyser node for visualization
      analyserNodeRef.current = audioContextRef.current.createAnalyser();
      analyserNodeRef.current.fftSize = 2048;
      analyserNodeRef.current.smoothingTimeConstant = 0.8;
      analyserDataRef.current = new Uint8Array(analyserNodeRef.current.frequencyBinCount) as Uint8Array<ArrayBuffer>;

      // Connect: source -> gain -> analyser -> destination
      gainNodeRef.current.connect(analyserNodeRef.current);
      analyserNodeRef.current.connect(audioContextRef.current.destination);

      // Register AudioWorklet
      if (!workletRegistered) {
        try {
          await audioContextRef.current.audioWorklet.addModule('/audio-worklet-processor.js');
          workletRegistered = true;

          // Create worklet node for real-time analysis
          workletNodeRef.current = new AudioWorkletNode(
            audioContextRef.current,
            'audio-analyzer-processor'
          );

          // Handle messages from worklet
          workletNodeRef.current.port.onmessage = (event) => {
            if (event.data.type === 'analysis') {
              setState(prev => ({
                ...prev,
                rms: event.data.rms,
                peak: event.data.peak,
                waveformData: event.data.buffer,
              }));
            }
          };

          // Insert worklet into chain: gain -> worklet -> analyser
          gainNodeRef.current.disconnect();
          gainNodeRef.current.connect(workletNodeRef.current);
          workletNodeRef.current.connect(analyserNodeRef.current);

          setState(prev => ({ ...prev, workletReady: true }));
        } catch (err) {
          console.warn('AudioWorklet not available, using fallback:', err);
        }
      }
    }
    return audioContextRef.current;
  }, []);

  // Update current time during playback
  const updateTime = useCallback(() => {
    if (state.isPlaying && audioContextRef.current) {
      const elapsed = audioContextRef.current.currentTime - startTimeRef.current;
      const currentTime = pausedAtRef.current + elapsed * state.playbackRate;

      if (currentTime >= state.duration) {
        setState(prev => ({ ...prev, isPlaying: false, currentTime: state.duration }));
        return;
      }

      setState(prev => ({ ...prev, currentTime }));
      animationFrameRef.current = requestAnimationFrame(updateTime);
    }
  }, [state.isPlaying, state.duration, state.playbackRate]);

  useEffect(() => {
    if (state.isPlaying) {
      animationFrameRef.current = requestAnimationFrame(updateTime);
    }
    return () => {
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current);
      }
    };
  }, [state.isPlaying, updateTime]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (sourceNodeRef.current) {
        sourceNodeRef.current.stop();
      }
      if (audioContextRef.current) {
        audioContextRef.current.close();
      }
    };
  }, []);

  const loadTrack = useCallback(async (url: string) => {
    setState(prev => ({ ...prev, isLoading: true, error: null, isStreaming: false }));

    try {
      const ctx = await getAudioContext();

      if (ctx.state === 'suspended') {
        await ctx.resume();
      }

      const response = await fetch(url);
      if (!response.ok) {
        throw new Error(`Failed to load audio: ${response.status}`);
      }

      const arrayBuffer = await response.arrayBuffer();
      const audioBuffer = await ctx.decodeAudioData(arrayBuffer);

      audioBufferRef.current = audioBuffer;
      pausedAtRef.current = 0;

      // Reset worklet state
      if (workletNodeRef.current) {
        workletNodeRef.current.port.postMessage({ type: 'reset' });
      }

      setState(prev => ({
        ...prev,
        isLoading: false,
        duration: audioBuffer.duration,
        currentTime: 0,
        rms: 0,
        peak: 0,
      }));
    } catch (err) {
      setState(prev => ({
        ...prev,
        isLoading: false,
        error: err instanceof Error ? err.message : 'Failed to load audio',
      }));
    }
  }, [getAudioContext]);

  // Load track with streaming (progressive loading)
  const loadStreamingTrack = useCallback(async (url: string) => {
    setState(prev => ({ ...prev, isLoading: true, error: null, isStreaming: true, bufferedTime: 0 }));

    try {
      const ctx = await getAudioContext();

      if (ctx.state === 'suspended') {
        await ctx.resume();
      }

      const response = await fetch(url);
      if (!response.ok) {
        throw new Error(`Failed to load audio: ${response.status}`);
      }

      // For streaming, we still need to buffer the whole file for Web Audio
      // True streaming would require MediaSource Extensions or a different approach
      const contentLength = response.headers.get('content-length');
      const totalBytes = contentLength ? parseInt(contentLength, 10) : 0;

      const reader = response.body?.getReader();
      if (!reader) {
        throw new Error('Streaming not supported');
      }

      const chunks: Uint8Array[] = [];
      let receivedBytes = 0;

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        chunks.push(value);
        receivedBytes += value.length;

        // Update buffered progress
        if (totalBytes > 0) {
          const progress = receivedBytes / totalBytes;
          setState(prev => ({ ...prev, bufferedTime: progress * 100 }));
        }
      }

      // Concatenate chunks
      const arrayBuffer = new ArrayBuffer(receivedBytes);
      const view = new Uint8Array(arrayBuffer);
      let offset = 0;
      for (const chunk of chunks) {
        view.set(chunk, offset);
        offset += chunk.length;
      }

      const audioBuffer = await ctx.decodeAudioData(arrayBuffer);
      audioBufferRef.current = audioBuffer;
      pausedAtRef.current = 0;

      setState(prev => ({
        ...prev,
        isLoading: false,
        duration: audioBuffer.duration,
        currentTime: 0,
        bufferedTime: 100,
      }));
    } catch (err) {
      setState(prev => ({
        ...prev,
        isLoading: false,
        isStreaming: false,
        error: err instanceof Error ? err.message : 'Failed to load audio',
      }));
    }
  }, [getAudioContext]);

  const play = useCallback(async () => {
    if (!audioBufferRef.current) {
      setState(prev => ({ ...prev, error: 'No audio loaded' }));
      return;
    }

    const ctx = await getAudioContext();

    if (ctx.state === 'suspended') {
      await ctx.resume();
    }

    if (sourceNodeRef.current) {
      sourceNodeRef.current.stop();
    }

    const source = ctx.createBufferSource();
    source.buffer = audioBufferRef.current;
    source.playbackRate.value = state.playbackRate;
    source.connect(gainNodeRef.current!);

    source.onended = () => {
      if (state.isPlaying) {
        setState(prev => ({ ...prev, isPlaying: false }));
      }
    };

    sourceNodeRef.current = source;
    startTimeRef.current = ctx.currentTime;

    source.start(0, pausedAtRef.current);

    setState(prev => ({ ...prev, isPlaying: true }));
  }, [getAudioContext, state.playbackRate, state.isPlaying]);

  const pause = useCallback(() => {
    if (sourceNodeRef.current && state.isPlaying) {
      const ctx = audioContextRef.current;
      if (ctx) {
        const elapsed = ctx.currentTime - startTimeRef.current;
        pausedAtRef.current += elapsed * state.playbackRate;
      }
      sourceNodeRef.current.stop();
      sourceNodeRef.current = null;
    }
    setState(prev => ({ ...prev, isPlaying: false }));
  }, [state.isPlaying, state.playbackRate]);

  const stop = useCallback(() => {
    if (sourceNodeRef.current) {
      sourceNodeRef.current.stop();
      sourceNodeRef.current = null;
    }
    pausedAtRef.current = 0;
    setState(prev => ({ ...prev, isPlaying: false, currentTime: 0, rms: 0, peak: 0 }));

    if (workletNodeRef.current) {
      workletNodeRef.current.port.postMessage({ type: 'reset' });
    }
  }, []);

  const seek = useCallback((time: number) => {
    const wasPlaying = state.isPlaying;

    if (sourceNodeRef.current) {
      sourceNodeRef.current.stop();
      sourceNodeRef.current = null;
    }

    pausedAtRef.current = Math.max(0, Math.min(time, state.duration));
    setState(prev => ({ ...prev, currentTime: pausedAtRef.current, isPlaying: false }));

    if (wasPlaying) {
      setTimeout(() => play(), 10);
    }
  }, [state.isPlaying, state.duration, play]);

  const seekToPosition = useCallback((position: number) => {
    seek(position * state.duration);
  }, [seek, state.duration]);

  const setPlaybackRate = useCallback((rate: number) => {
    setState(prev => ({ ...prev, playbackRate: rate }));
    if (sourceNodeRef.current) {
      sourceNodeRef.current.playbackRate.value = rate;
    }
  }, []);

  // Get frequency data for visualization
  const getAnalyserData = useCallback(() => {
    if (analyserNodeRef.current && analyserDataRef.current) {
      analyserNodeRef.current.getByteFrequencyData(analyserDataRef.current);
      return analyserDataRef.current;
    }
    return null;
  }, []);

  // Memoize the controls object to prevent infinite re-render loops
  const controls: AudioPlayerControls = React.useMemo(
    () => ({
      play,
      pause,
      stop,
      seek,
      seekToPosition,
      setPlaybackRate,
      loadTrack,
      loadStreamingTrack,
      getAnalyserData,
    }),
    [play, pause, stop, seek, seekToPosition, setPlaybackRate, loadTrack, loadStreamingTrack, getAnalyserData]
  );

  return [state, controls];
}
