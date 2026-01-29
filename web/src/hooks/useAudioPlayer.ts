import { useRef, useState, useCallback, useEffect } from 'react';

export interface AudioPlayerState {
  isPlaying: boolean;
  isLoading: boolean;
  currentTime: number;
  duration: number;
  playbackRate: number;
  error: string | null;
}

export interface AudioPlayerControls {
  play: () => Promise<void>;
  pause: () => void;
  stop: () => void;
  seek: (time: number) => void;
  seekToPosition: (position: number) => void;
  setPlaybackRate: (rate: number) => void;
  loadTrack: (url: string) => Promise<void>;
}

export function useAudioPlayer(): [AudioPlayerState, AudioPlayerControls] {
  const audioContextRef = useRef<AudioContext | null>(null);
  const sourceNodeRef = useRef<AudioBufferSourceNode | null>(null);
  const audioBufferRef = useRef<AudioBuffer | null>(null);
  const gainNodeRef = useRef<GainNode | null>(null);
  const startTimeRef = useRef<number>(0);
  const pausedAtRef = useRef<number>(0);
  const animationFrameRef = useRef<number>(0);

  const [state, setState] = useState<AudioPlayerState>({
    isPlaying: false,
    isLoading: false,
    currentTime: 0,
    duration: 0,
    playbackRate: 1,
    error: null,
  });

  // Initialize audio context lazily (requires user interaction)
  const getAudioContext = useCallback(() => {
    if (!audioContextRef.current) {
      audioContextRef.current = new AudioContext();
      gainNodeRef.current = audioContextRef.current.createGain();
      gainNodeRef.current.connect(audioContextRef.current.destination);
    }
    return audioContextRef.current;
  }, []);

  // Update current time during playback
  const updateTime = useCallback(() => {
    if (state.isPlaying && audioContextRef.current) {
      const elapsed = audioContextRef.current.currentTime - startTimeRef.current;
      const currentTime = pausedAtRef.current + elapsed * state.playbackRate;

      if (currentTime >= state.duration) {
        // Track ended
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
    setState(prev => ({ ...prev, isLoading: true, error: null }));

    try {
      const ctx = getAudioContext();

      // Resume context if suspended (browser autoplay policy)
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

      setState(prev => ({
        ...prev,
        isLoading: false,
        duration: audioBuffer.duration,
        currentTime: 0,
      }));
    } catch (err) {
      setState(prev => ({
        ...prev,
        isLoading: false,
        error: err instanceof Error ? err.message : 'Failed to load audio',
      }));
    }
  }, [getAudioContext]);

  const play = useCallback(async () => {
    if (!audioBufferRef.current) {
      setState(prev => ({ ...prev, error: 'No audio loaded' }));
      return;
    }

    const ctx = getAudioContext();

    // Resume context if suspended
    if (ctx.state === 'suspended') {
      await ctx.resume();
    }

    // Stop any existing playback
    if (sourceNodeRef.current) {
      sourceNodeRef.current.stop();
    }

    // Create new source node
    const source = ctx.createBufferSource();
    source.buffer = audioBufferRef.current;
    source.playbackRate.value = state.playbackRate;
    source.connect(gainNodeRef.current!);

    // Handle track end
    source.onended = () => {
      if (state.isPlaying) {
        setState(prev => ({ ...prev, isPlaying: false }));
      }
    };

    sourceNodeRef.current = source;
    startTimeRef.current = ctx.currentTime;

    // Start from paused position
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
    setState(prev => ({ ...prev, isPlaying: false, currentTime: 0 }));
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
      // Small delay to allow state update
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

  return [
    state,
    {
      play,
      pause,
      stop,
      seek,
      seekToPosition,
      setPlaybackRate,
      loadTrack,
    },
  ];
}
