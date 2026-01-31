import { useMemo, useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { ThemeToggle } from './components/ThemeToggle';
import { LibraryGrid } from './components/LibraryGrid';
import { TrackDetail } from './components/TrackDetail';
import { SetBuilder } from './components/SetBuilder';
import { TransitionRehearsal } from './components/TransitionRehearsal';
import { WaveformCanvas } from './components/WaveformCanvas';
import { SpectrumAnalyzer } from './components/SpectrumAnalyzer';
import { EnergyArc } from './components/EnergyArc';
import { TransitionGraph } from './components/TransitionGraph';
import { LiveStats } from './components/LiveStats';
import { BPMKeyChart } from './components/BPMKeyChart';
import { ExportPanel } from './components/ExportPanel';
import { SimilarTracks } from './components/SimilarTracks';
import { ModelSettings } from './components/ModelSettings';
import { TrainingScreen } from './components/TrainingScreen';
import { IntroWizard } from './components/IntroWizard';
import { useStore } from './store';

function App() {
  // Get state and actions from store
  const {
    tracks,
    trackMap,
    currentSetPlan,
    filteredTracks,
    selectedTrack,
    selectedId,
    viewMode,
    query,
    onlyReview,
    onlyAnalyzed,
    highEnergyOnly,
    sortMode,
    chartMode,
    isLoading,
    error,
    apiAvailable,
    stats,
    setViewMode,
    setQuery,
    setOnlyReview,
    setOnlyAnalyzed,
    setHighEnergyOnly,
    setSortMode,
    setSelectedId,
    setChartMode,
    fetchTracks,
    checkApiHealth,
    proposeSet,
    fetchTrackDetail,
    hasCompletedOnboarding,
  } = useStore();

  // Local UI state for playback simulation
  const [playheadPosition, setPlayheadPosition] = useState(0.3);
  const [isPlaying, setIsPlaying] = useState(false);

  // Computed values
  const filtered = filteredTracks();
  const selected = selectedTrack();
  const currentStats = stats();

  // Get set plan (from store or generate demo)
  const setPlan = currentSetPlan || {
    mode: 'Peak-time' as const,
    order: filtered.slice(0, 10).map((t) => t.id),
    edges: [],
  };

  const currentSetIndex = useMemo(() => {
    return setPlan.order.findIndex((id) => id === selectedId);
  }, [selectedId, setPlan.order]);

  const currentEdge = useMemo(() => {
    if (currentSetIndex > 0) {
      return setPlan.edges[currentSetIndex - 1];
    }
    return setPlan.edges[0];
  }, [currentSetIndex, setPlan.edges]);

  const fromTrack = trackMap[currentEdge?.from ?? ''];
  const toTrack = trackMap[currentEdge?.to ?? ''];

  const setEnergyValues = useMemo(() => {
    return setPlan.order.map((id) => trackMap[id]?.energy ?? 5);
  }, [trackMap, setPlan.order]);

  const setTrackLabels = useMemo(() => {
    return setPlan.order.map((id) => trackMap[id]?.title ?? id);
  }, [trackMap, setPlan.order]);

  // Initialize: check API and load tracks
  useEffect(() => {
    const init = async () => {
      await checkApiHealth();
      try {
        await fetchTracks();
      } catch {
        // Falls back to mock data automatically
      }
    };
    init();
  }, [checkApiHealth, fetchTracks]);

  // Fetch full track detail when selection changes
  useEffect(() => {
    if (selectedId && apiAvailable) {
      fetchTrackDetail(selectedId);
    }
  }, [selectedId, apiAvailable, fetchTrackDetail]);

  // Auto-propose set when entering set builder with tracks
  const handleSetBuilderEnter = useCallback(async () => {
    setViewMode('setBuilder');
    if (apiAvailable && tracks.length >= 2 && !currentSetPlan) {
      const analyzedTracks = tracks.filter((t) => t.status === 'analyzed');
      if (analyzedTracks.length >= 2) {
        await proposeSet(
          analyzedTracks.slice(0, 15).map((t) => t.id),
          'Peak-time'
        );
      }
    }
  }, [setViewMode, apiAvailable, tracks, currentSetPlan, proposeSet]);

  // Auto-play simulation for GIF capture
  useEffect(() => {
    if (isPlaying) {
      const interval = setInterval(() => {
        setPlayheadPosition((p) => {
          if (p >= 1) {
            setIsPlaying(false);
            return 0;
          }
          return p + 0.003;
        });
      }, 50);
      return () => clearInterval(interval);
    }
  }, [isPlaying]);

  const handlePlay = () => {
    setIsPlaying(!isPlaying);
    if (!isPlaying) {
      setPlayheadPosition(0);
    }
  };

  // Show intro wizard for new users
  if (!hasCompletedOnboarding && tracks.length === 0) {
    return (
      <div className="app">
        <IntroWizard />
      </div>
    );
  }

  // Loading state
  if (isLoading) {
    return (
      <div className="app">
        <div className="loading-screen">
          <div className="loading-spinner" />
          <p>Loading tracks...</p>
          {error && <p className="error-text">{error}</p>}
        </div>
      </div>
    );
  }

  return (
    <div className="app">
      <header className="app-header">
        <div className="header-left">
          <div className="app-logo">
            <span className="logo-icon">◈</span>
            Algiers
          </div>
          <span className="version-badge">alpha</span>
          {!apiAvailable && <span className="demo-badge">demo</span>}
        </div>
        <nav className="header-nav">
          <button
            className={`nav-btn ${viewMode === 'library' ? 'active' : ''}`}
            onClick={() => setViewMode('library')}
          >
            <span className="nav-icon">◫</span>
            Library
          </button>
          <button
            className={`nav-btn ${viewMode === 'setBuilder' ? 'active' : ''}`}
            onClick={handleSetBuilderEnter}
          >
            <span className="nav-icon">≡</span>
            Set Builder
          </button>
          <button
            className={`nav-btn ${viewMode === 'graph' ? 'active' : ''}`}
            onClick={() => setViewMode('graph')}
          >
            <span className="nav-icon">◎</span>
            Graph
          </button>
          <button
            className={`nav-btn ${viewMode === 'settings' ? 'active' : ''}`}
            onClick={() => setViewMode('settings')}
          >
            <span className="nav-icon">⚙</span>
            Settings
          </button>
          <button
            className={`nav-btn ${viewMode === 'training' ? 'active' : ''}`}
            onClick={() => setViewMode('training')}
          >
            <span className="nav-icon">🤖</span>
            Training
          </button>
          <div className="nav-divider" />
          <ThemeToggle />
        </nav>
      </header>

      <main className="app-main">
        <AnimatePresence mode="wait">
          {viewMode === 'library' && (
            <motion.div
              key="library"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
              className="library-view"
            >
              {/* Top Bar - Stats + Charts */}
              <div className="top-bar">
                <LiveStats stats={currentStats} compact />
                <div className="charts-row">
                  <div className="chart-box">
                    <div className="chart-tabs">
                      <button
                        className={`chart-tab ${chartMode === 'bpm' ? 'active' : ''}`}
                        onClick={() => setChartMode('bpm')}
                      >
                        BPM
                      </button>
                      <button
                        className={`chart-tab ${chartMode === 'key' ? 'active' : ''}`}
                        onClick={() => setChartMode('key')}
                      >
                        Keys
                      </button>
                    </div>
                    <BPMKeyChart tracks={tracks} height={100} mode={chartMode} />
                  </div>
                  <div className="chart-box">
                    <div className="chart-label">Spectrum</div>
                    <SpectrumAnalyzer height={100} isActive={isPlaying} style="mirror" />
                  </div>
                </div>
              </div>

              {/* Main Content Grid */}
              <div className="content-grid">
                {/* Left: Library */}
                <div className="panel library-panel">
                  <div className="panel-header">
                    <h3>Library</h3>
                    <span className="count-badge">{filtered.length} tracks</span>
                  </div>
                  <div className="filter-bar">
                    <input
                      className="search-input"
                      placeholder="Search..."
                      value={query}
                      onChange={(e) => setQuery(e.target.value)}
                    />
                    <div className="filter-group">
                      <label className="filter-check">
                        <input
                          type="checkbox"
                          checked={onlyReview}
                          onChange={(e) => setOnlyReview(e.target.checked)}
                        />
                        Review
                      </label>
                      <label className="filter-check">
                        <input
                          type="checkbox"
                          checked={onlyAnalyzed}
                          onChange={(e) => setOnlyAnalyzed(e.target.checked)}
                        />
                        Analyzed
                      </label>
                      <label className="filter-check">
                        <input
                          type="checkbox"
                          checked={highEnergyOnly}
                          onChange={(e) => setHighEnergyOnly(e.target.checked)}
                        />
                        High E
                      </label>
                    </div>
                    <select
                      className="sort-select"
                      value={sortMode}
                      onChange={(e) =>
                        setSortMode(e.target.value as 'bpm-asc' | 'bpm-desc' | 'energy-desc')
                      }
                    >
                      <option value="energy-desc">Energy ↓</option>
                      <option value="bpm-asc">BPM ↑</option>
                      <option value="bpm-desc">BPM ↓</option>
                    </select>
                  </div>
                  <LibraryGrid tracks={filtered} selectedId={selected?.id} onSelect={setSelectedId} />
                </div>

                {/* Center: Track Detail + Waveform */}
                <div className="panel detail-panel">
                  <div className="panel-header">
                    <h3>Track Detail</h3>
                    <div className="transport-controls">
                      <button className="transport-btn" onClick={() => setPlayheadPosition(0)}>
                        ⏮
                      </button>
                      <button className="transport-btn play" onClick={handlePlay}>
                        {isPlaying ? '⏸' : '▶'}
                      </button>
                      <button className="transport-btn" onClick={() => setPlayheadPosition(1)}>
                        ⏭
                      </button>
                    </div>
                  </div>
                  {selected && (
                    <>
                      <TrackDetail track={selected} />
                      <div className="waveform-container">
                        <WaveformCanvas
                          peaks={selected.waveformSummary}
                          sections={selected.sections}
                          cues={selected.cues}
                          playheadPosition={playheadPosition}
                          duration={180}
                          isPlaying={isPlaying}
                          onSeek={setPlayheadPosition}
                          height={90}
                          bpm={selected.bpm}
                        />
                      </div>
                      <SpectrumAnalyzer height={50} isActive={isPlaying} style="bars" colorScheme="accent" bands={32} />
                    </>
                  )}
                </div>

                {/* Right: Set Builder + Energy */}
                <div className="panel set-panel">
                  <div className="panel-header">
                    <h3>Set Order</h3>
                    <span className="pill pill-primary">{setPlan.mode}</span>
                  </div>
                  <SetBuilder plan={setPlan} tracks={trackMap} compact />
                  <div className="energy-section">
                    <div className="section-label">Energy Flow</div>
                    <EnergyArc
                      values={setEnergyValues}
                      labels={setTrackLabels}
                      height={80}
                      highlightIndex={currentSetIndex >= 0 ? currentSetIndex : undefined}
                    />
                  </div>
                  <TransitionRehearsal from={fromTrack} to={toTrack} edge={currentEdge} compact />
                </div>
              </div>
            </motion.div>
          )}

          {viewMode === 'setBuilder' && (
            <motion.div
              key="setBuilder"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
              className="set-builder-view"
            >
              <div className="set-main">
                <div className="panel">
                  <div className="panel-header">
                    <h3>Set Builder</h3>
                    <div className="pill-row">
                      <span className="pill pill-primary">{setPlan.mode}</span>
                      <span className="pill pill-analyzed">{setPlan.order.length} tracks</span>
                    </div>
                  </div>
                  <SetBuilder plan={setPlan} tracks={trackMap} />
                </div>
              </div>
              <div className="set-sidebar">
                <div className="panel">
                  <div className="panel-header">
                    <h3>Energy Journey</h3>
                  </div>
                  <EnergyArc
                    values={setEnergyValues}
                    labels={setTrackLabels}
                    height={120}
                    highlightIndex={currentSetIndex >= 0 ? currentSetIndex : undefined}
                  />
                </div>
                <TransitionRehearsal from={fromTrack} to={toTrack} edge={currentEdge} />
                {fromTrack && toTrack && (
                  <div className="panel">
                    <div className="panel-header">
                      <h3>Transition</h3>
                    </div>
                    <div className="dual-waveform">
                      <div className="waveform-label">
                        <span className="deck-badge a">A</span>
                        {fromTrack.title}
                      </div>
                      <WaveformCanvas
                        peaks={fromTrack.waveformSummary}
                        sections={fromTrack.sections}
                        height={45}
                        playheadPosition={0.85}
                        bpm={fromTrack.bpm}
                      />
                      <div className="waveform-label">
                        <span className="deck-badge b">B</span>
                        {toTrack.title}
                      </div>
                      <WaveformCanvas
                        peaks={toTrack.waveformSummary}
                        sections={toTrack.sections}
                        height={45}
                        playheadPosition={0.1}
                        bpm={toTrack.bpm}
                      />
                    </div>
                  </div>
                )}
                <div className="panel">
                  <ExportPanel trackIds={setPlan.order} playlistName={`${setPlan.mode.toLowerCase().replace(' ', '-')}-set`} />
                </div>
              </div>
            </motion.div>
          )}

          {viewMode === 'graph' && (
            <motion.div
              key="graph"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
              className="graph-view"
            >
              <div className="graph-main">
                <div className="panel graph-panel">
                  <div className="panel-header">
                    <h3>Transition Graph</h3>
                    <span className="muted">D3.js force-directed · click nodes to select</span>
                  </div>
                  <TransitionGraph
                    tracks={trackMap}
                    edges={setPlan.edges}
                    height={480}
                    selectedTrackId={selectedId ?? undefined}
                    onSelectTrack={setSelectedId}
                  />
                </div>
              </div>
              <div className="graph-sidebar">
                {selected && (
                  <div className="panel">
                    <TrackDetail track={selected} />
                    <WaveformCanvas
                      peaks={selected.waveformSummary}
                      sections={selected.sections}
                      cues={selected.cues}
                      height={70}
                      playheadPosition={playheadPosition}
                      bpm={selected.bpm}
                      onSeek={setPlayheadPosition}
                    />
                  </div>
                )}
                <div className="panel">
                  <div className="panel-header">
                    <h3>Set Energy</h3>
                  </div>
                  <EnergyArc
                    values={setEnergyValues}
                    labels={setTrackLabels}
                    height={90}
                    highlightIndex={currentSetIndex >= 0 ? currentSetIndex : undefined}
                  />
                </div>
                <div className="panel">
                  <BPMKeyChart tracks={tracks} height={100} mode="key" />
                </div>
              </div>
            </motion.div>
          )}

          {viewMode === 'settings' && (
            <motion.div
              key="settings"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
              className="settings-view"
            >
              <div className="settings-main">
                <div className="panel">
                  <ModelSettings />
                </div>
              </div>
              <div className="settings-sidebar">
                {selected && (
                  <div className="panel">
                    <div className="panel-header">
                      <h3>Similar to: {selected.title}</h3>
                    </div>
                    <SimilarTracks
                      trackId={selected.id}
                      onSelectTrack={setSelectedId}
                      limit={8}
                    />
                  </div>
                )}
                {!selected && (
                  <div className="panel">
                    <div className="panel-header">
                      <h3>Similar Tracks</h3>
                    </div>
                    <div className="similar-tracks-empty">
                      <span className="empty-icon">🎵</span>
                      <p>Select a track to find similar ones</p>
                    </div>
                  </div>
                )}
              </div>
            </motion.div>
          )}

          {viewMode === 'training' && (
            <motion.div
              key="training"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
              className="training-view"
            >
              <TrainingScreen />
            </motion.div>
          )}
        </AnimatePresence>
      </main>

      <footer className="app-footer">
        <span>Algiers · DJ Set Prep Copilot</span>
        <span className="muted">
          v0.7-beta · Apple Silicon M1–M5
          {apiAvailable ? ' · API connected' : ' · demo mode'}
        </span>
      </footer>
    </div>
  );
}

export default App;
