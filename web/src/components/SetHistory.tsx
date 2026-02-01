import { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import type { SetSession, SetHistoryStats } from '../types';
import { useStore } from '../store';

// LocalStorage key for set history
const HISTORY_KEY = 'algiers-set-history';

// Generate unique ID
function generateId(): string {
  return `set_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

// Load history from localStorage
function loadHistory(): SetSession[] {
  try {
    const data = localStorage.getItem(HISTORY_KEY);
    return data ? JSON.parse(data) : [];
  } catch {
    return [];
  }
}

// Save history to localStorage
function saveHistory(sessions: SetSession[]): void {
  localStorage.setItem(HISTORY_KEY, JSON.stringify(sessions));
}

// Calculate stats from history
function calculateStats(sessions: SetSession[]): SetHistoryStats {
  if (sessions.length === 0) {
    return {
      totalSessions: 0,
      totalTracks: 0,
      avgTracksPerSet: 0,
      favoriteMode: 'Peak-time',
      lastSessionDate: null,
    };
  }

  const totalTracks = sessions.reduce((sum, s) => sum + s.trackCount, 0);
  const modeCounts: Record<string, number> = {};
  sessions.forEach(s => {
    modeCounts[s.plan.mode] = (modeCounts[s.plan.mode] || 0) + 1;
  });
  const favoriteMode = Object.entries(modeCounts)
    .sort((a, b) => b[1] - a[1])[0][0] as SetHistoryStats['favoriteMode'];

  return {
    totalSessions: sessions.length,
    totalTracks,
    avgTracksPerSet: Math.round(totalTracks / sessions.length),
    favoriteMode,
    lastSessionDate: sessions[0]?.updatedAt || null,
  };
}

interface SetHistoryProps {
  onLoadSession?: (session: SetSession) => void;
}

export function SetHistory({ onLoadSession }: SetHistoryProps) {
  const { currentSetPlan, trackMap } = useStore();
  const [sessions, setSessions] = useState<SetSession[]>([]);
  const [stats, setStats] = useState<SetHistoryStats | null>(null);
  const [showSaveModal, setShowSaveModal] = useState(false);
  const [newSessionName, setNewSessionName] = useState('');
  const [newSessionNotes, setNewSessionNotes] = useState('');
  const [selectedSession, setSelectedSession] = useState<SetSession | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  // Load history on mount
  useEffect(() => {
    const loaded = loadHistory();
    setSessions(loaded);
    setStats(calculateStats(loaded));
  }, []);

  // Save current set as a new session
  const handleSaveSession = useCallback(() => {
    if (!currentSetPlan || !newSessionName.trim()) return;

    const tracksInSet = currentSetPlan.order
      .map(id => trackMap[id])
      .filter(Boolean);

    const totalDuration = tracksInSet.length * 4; // Assume ~4 min per track
    const avgBpm = tracksInSet.reduce((sum, t) => sum + t.bpm, 0) / Math.max(tracksInSet.length, 1);

    const newSession: SetSession = {
      id: generateId(),
      name: newSessionName.trim(),
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      plan: { ...currentSetPlan },
      trackCount: tracksInSet.length,
      totalDuration: Math.round(totalDuration),
      avgBpm: Math.round(avgBpm),
      notes: newSessionNotes.trim() || undefined,
    };

    const updated = [newSession, ...sessions];
    setSessions(updated);
    saveHistory(updated);
    setStats(calculateStats(updated));
    setShowSaveModal(false);
    setNewSessionName('');
    setNewSessionNotes('');
  }, [currentSetPlan, trackMap, newSessionName, newSessionNotes, sessions]);

  // Delete a session
  const handleDeleteSession = useCallback((id: string) => {
    const updated = sessions.filter(s => s.id !== id);
    setSessions(updated);
    saveHistory(updated);
    setStats(calculateStats(updated));
    if (selectedSession?.id === id) {
      setSelectedSession(null);
    }
  }, [sessions, selectedSession]);

  // Load a session
  const handleLoadSession = useCallback((session: SetSession) => {
    if (onLoadSession) {
      onLoadSession(session);
    }
    // Update the session's last accessed time
    const updated = sessions.map(s =>
      s.id === session.id ? { ...s, updatedAt: new Date().toISOString() } : s
    ).sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime());
    setSessions(updated);
    saveHistory(updated);
  }, [sessions, onLoadSession]);

  // Filter sessions by search
  const filteredSessions = sessions.filter(s =>
    s.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    s.notes?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Format date for display
  const formatDate = (iso: string) => {
    const date = new Date(iso);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Yesterday';
    if (diffDays < 7) return `${diffDays} days ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="set-history">
      {/* Stats Header */}
      {stats && stats.totalSessions > 0 && (
        <div className="history-stats">
          <div className="stat-card">
            <span className="stat-value">{stats.totalSessions}</span>
            <span className="stat-label">Sessions</span>
          </div>
          <div className="stat-card">
            <span className="stat-value">{stats.totalTracks}</span>
            <span className="stat-label">Total Tracks</span>
          </div>
          <div className="stat-card">
            <span className="stat-value">{stats.avgTracksPerSet}</span>
            <span className="stat-label">Avg per Set</span>
          </div>
          <div className="stat-card">
            <span className="stat-value">{stats.favoriteMode}</span>
            <span className="stat-label">Favorite Mode</span>
          </div>
        </div>
      )}

      {/* Actions Bar */}
      <div className="history-actions">
        <div className="search-box">
          <input
            type="text"
            placeholder="Search sessions..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
        <button
          className="save-btn"
          onClick={() => setShowSaveModal(true)}
          disabled={!currentSetPlan || currentSetPlan.order.length === 0}
        >
          Save Current Set
        </button>
      </div>

      {/* Sessions List */}
      <div className="sessions-list">
        {filteredSessions.length === 0 ? (
          <div className="empty-state">
            <span className="empty-icon">📚</span>
            <p>No saved sessions yet</p>
            <span className="empty-hint">
              Build a set and click "Save Current Set" to start tracking your history
            </span>
          </div>
        ) : (
          <AnimatePresence>
            {filteredSessions.map((session) => (
              <motion.div
                key={session.id}
                className={`session-card ${selectedSession?.id === session.id ? 'selected' : ''}`}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                onClick={() => setSelectedSession(selectedSession?.id === session.id ? null : session)}
              >
                <div className="session-header">
                  <h4 className="session-name">{session.name}</h4>
                  <span className="session-date">{formatDate(session.updatedAt)}</span>
                </div>

                <div className="session-meta">
                  <span className="meta-item">
                    <span className="meta-icon">🎵</span>
                    {session.trackCount} tracks
                  </span>
                  <span className="meta-item">
                    <span className="meta-icon">⏱</span>
                    ~{session.totalDuration} min
                  </span>
                  <span className="meta-item">
                    <span className="meta-icon">💨</span>
                    {session.avgBpm} BPM
                  </span>
                  <span className={`mode-badge ${session.plan.mode.toLowerCase().replace('-', '')}`}>
                    {session.plan.mode}
                  </span>
                </div>

                {session.notes && (
                  <p className="session-notes">{session.notes}</p>
                )}

                <AnimatePresence>
                  {selectedSession?.id === session.id && (
                    <motion.div
                      className="session-actions"
                      initial={{ opacity: 0, height: 0 }}
                      animate={{ opacity: 1, height: 'auto' }}
                      exit={{ opacity: 0, height: 0 }}
                    >
                      <button
                        className="action-btn load"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleLoadSession(session);
                        }}
                      >
                        Load Set
                      </button>
                      <button
                        className="action-btn delete"
                        onClick={(e) => {
                          e.stopPropagation();
                          if (confirm(`Delete "${session.name}"?`)) {
                            handleDeleteSession(session.id);
                          }
                        }}
                      >
                        Delete
                      </button>
                    </motion.div>
                  )}
                </AnimatePresence>
              </motion.div>
            ))}
          </AnimatePresence>
        )}
      </div>

      {/* Save Modal */}
      <AnimatePresence>
        {showSaveModal && (
          <motion.div
            className="modal-overlay"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => setShowSaveModal(false)}
          >
            <motion.div
              className="save-modal"
              initial={{ scale: 0.9, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.9, opacity: 0 }}
              onClick={(e) => e.stopPropagation()}
            >
              <h3>Save Current Set</h3>

              <div className="form-group">
                <label>Session Name</label>
                <input
                  type="text"
                  placeholder="e.g., Friday Night Set, Club Opening"
                  value={newSessionName}
                  onChange={(e) => setNewSessionName(e.target.value)}
                  autoFocus
                />
              </div>

              <div className="form-group">
                <label>Notes (optional)</label>
                <textarea
                  placeholder="Any notes about this set..."
                  value={newSessionNotes}
                  onChange={(e) => setNewSessionNotes(e.target.value)}
                  rows={3}
                />
              </div>

              {currentSetPlan && (
                <div className="set-preview">
                  <span>{currentSetPlan.order.length} tracks</span>
                  <span>{currentSetPlan.mode} mode</span>
                </div>
              )}

              <div className="modal-actions">
                <button className="cancel-btn" onClick={() => setShowSaveModal(false)}>
                  Cancel
                </button>
                <button
                  className="confirm-btn"
                  onClick={handleSaveSession}
                  disabled={!newSessionName.trim()}
                >
                  Save Session
                </button>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
