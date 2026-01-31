import { useEffect, useCallback } from 'react';
import { useStore } from '../store';

export interface ShortcutConfig {
  key: string;
  ctrl?: boolean;
  meta?: boolean;
  shift?: boolean;
  alt?: boolean;
  action: () => void;
  description: string;
  category: 'navigation' | 'playback' | 'selection' | 'view' | 'general';
}

// Global shortcuts registry
export const SHORTCUTS: ShortcutConfig[] = [];

export function useKeyboardShortcuts() {
  const {
    viewMode,
    setViewMode,
    filteredTracks,
    selectedId,
    setSelectedId,
    batchMode,
    setBatchMode,
    selectAllTracks,
    selectNoneTracks,
    toggleBatchSelection,
  } = useStore();

  const filtered = filteredTracks();

  // Navigate to next track
  const navigateNext = useCallback(() => {
    const currentIndex = filtered.findIndex((t) => t.id === selectedId);
    if (currentIndex < filtered.length - 1) {
      setSelectedId(filtered[currentIndex + 1].id);
    }
  }, [filtered, selectedId, setSelectedId]);

  // Navigate to previous track
  const navigatePrev = useCallback(() => {
    const currentIndex = filtered.findIndex((t) => t.id === selectedId);
    if (currentIndex > 0) {
      setSelectedId(filtered[currentIndex - 1].id);
    }
  }, [filtered, selectedId, setSelectedId]);

  // Toggle batch selection for current track
  const toggleCurrentSelection = useCallback(() => {
    if (selectedId && batchMode) {
      toggleBatchSelection(selectedId);
    }
  }, [selectedId, batchMode, toggleBatchSelection]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Ignore if typing in an input
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLTextAreaElement ||
        (e.target as HTMLElement)?.isContentEditable
      ) {
        return;
      }

      const key = e.key.toLowerCase();
      const meta = e.metaKey || e.ctrlKey;
      const shift = e.shiftKey;

      // === Navigation ===
      // Arrow Down / J - Next track
      if ((key === 'arrowdown' || key === 'j') && !meta && !shift) {
        e.preventDefault();
        navigateNext();
        return;
      }

      // Arrow Up / K - Previous track
      if ((key === 'arrowup' || key === 'k') && !meta && !shift) {
        e.preventDefault();
        navigatePrev();
        return;
      }

      // === View Switching ===
      // 1 - Library view
      if (key === '1' && !meta) {
        e.preventDefault();
        setViewMode('library');
        return;
      }

      // 2 - Set Builder view
      if (key === '2' && !meta) {
        e.preventDefault();
        setViewMode('setBuilder');
        return;
      }

      // 3 - Graph view
      if (key === '3' && !meta) {
        e.preventDefault();
        setViewMode('graph');
        return;
      }

      // 4 - Settings view
      if (key === '4' && !meta) {
        e.preventDefault();
        setViewMode('settings');
        return;
      }

      // 5 - Training view
      if (key === '5' && !meta) {
        e.preventDefault();
        setViewMode('training');
        return;
      }

      // === Batch Mode ===
      // Cmd+A - Select all (only in library view and batch mode)
      if (key === 'a' && meta && viewMode === 'library') {
        e.preventDefault();
        if (!batchMode) {
          setBatchMode(true);
        }
        selectAllTracks();
        return;
      }

      // Cmd+Shift+A - Deselect all
      if (key === 'a' && meta && shift && viewMode === 'library') {
        e.preventDefault();
        selectNoneTracks();
        return;
      }

      // S - Toggle batch selection mode
      if (key === 's' && !meta && viewMode === 'library') {
        e.preventDefault();
        setBatchMode(!batchMode);
        return;
      }

      // Space - Toggle selection of current track (in batch mode)
      if (key === ' ' && batchMode && viewMode === 'library') {
        e.preventDefault();
        toggleCurrentSelection();
        return;
      }

      // Escape - Exit batch mode or clear selection
      if (key === 'escape') {
        e.preventDefault();
        if (batchMode) {
          setBatchMode(false);
        }
        return;
      }

      // ? - Show keyboard shortcuts help
      if (key === '?' || (key === '/' && shift)) {
        e.preventDefault();
        // Dispatch custom event that the help modal can listen to
        window.dispatchEvent(new CustomEvent('show-shortcuts-help'));
        return;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [
    viewMode,
    setViewMode,
    navigateNext,
    navigatePrev,
    batchMode,
    setBatchMode,
    selectAllTracks,
    selectNoneTracks,
    toggleCurrentSelection,
  ]);
}

// Shortcuts list for help display
export const SHORTCUT_LIST = [
  { category: 'Navigation', shortcuts: [
    { keys: ['↓', 'J'], description: 'Next track' },
    { keys: ['↑', 'K'], description: 'Previous track' },
  ]},
  { category: 'Views', shortcuts: [
    { keys: ['1'], description: 'Library' },
    { keys: ['2'], description: 'Set Builder' },
    { keys: ['3'], description: 'Graph' },
    { keys: ['4'], description: 'Settings' },
    { keys: ['5'], description: 'Training' },
  ]},
  { category: 'Selection', shortcuts: [
    { keys: ['S'], description: 'Toggle batch mode' },
    { keys: ['Space'], description: 'Toggle track selection' },
    { keys: ['⌘', 'A'], description: 'Select all tracks' },
    { keys: ['⌘', '⇧', 'A'], description: 'Deselect all' },
    { keys: ['Esc'], description: 'Exit batch mode' },
  ]},
  { category: 'Help', shortcuts: [
    { keys: ['?'], description: 'Show shortcuts' },
  ]},
];
