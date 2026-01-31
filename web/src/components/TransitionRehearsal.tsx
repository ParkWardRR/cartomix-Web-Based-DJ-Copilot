import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import type { SetEdge, Track } from '../types';
import { CrossfadePreview } from './CrossfadePreview';

type Props = {
  from?: Track;
  to?: Track;
  edge?: SetEdge;
  compact?: boolean;
};

export function TransitionRehearsal({ from, to, edge, compact }: Props) {
  const [showPreview, setShowPreview] = useState(false);

  if (!from || !to || !edge) {
    return null;
  }

  return (
    <div className={`transition-card ${compact ? 'compact' : ''}`}>
      <div className="transition-header">
        <h4>Transition</h4>
        <span className="pill pill-primary">{edge.score.toFixed(1)}</span>
      </div>

      <div className="transition-decks">
        <div className="deck">
          <div className="deck-label">
            <span className="deck-badge a">A</span>
            <span className="muted">{from.bpm} BPM</span>
          </div>
          <div className="track-title" style={{ fontSize: '0.75rem' }}>{from.title}</div>
        </div>
        <div className="deck">
          <div className="deck-label">
            <span className="deck-badge b">B</span>
            <span className="muted">{to.bpm} BPM</span>
          </div>
          <div className="track-title" style={{ fontSize: '0.75rem' }}>{to.title}</div>
        </div>
      </div>

      <div className="transition-reasons">
        <div className="reason-list">
          <span className="pill pill-secondary">{edge.keyRelation}</span>
          <span className="pill pill-secondary">Δ{edge.tempoDelta} BPM</span>
          <span className="pill pill-secondary">{edge.window}</span>
        </div>
      </div>

      {/* Preview button */}
      <motion.button
        className="preview-transition-btn"
        onClick={() => setShowPreview(!showPreview)}
        whileHover={{ scale: 1.02 }}
        whileTap={{ scale: 0.98 }}
      >
        <span className="preview-icon">{showPreview ? '▼' : '▶'}</span>
        <span>{showPreview ? 'Hide Preview' : 'Preview Crossfade'}</span>
      </motion.button>

      {/* Crossfade preview panel */}
      <AnimatePresence>
        {showPreview && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            transition={{ duration: 0.3 }}
          >
            <CrossfadePreview
              trackA={from}
              trackB={to}
              transitionPoint={0.8}
              crossfadeDuration={16}
              onClose={() => setShowPreview(false)}
            />
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
