import { useMemo, useState, useEffect } from 'react';
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
import { demoSetPlan, demoTracks } from './mockData';
import type { Track } from './types';

type ViewMode = 'library' | 'setBuilder' | 'graph';

function App() {
  const [query, setQuery] = useState('');
  const [onlyReview, setOnlyReview] = useState(false);
  const [onlyAnalyzed, setOnlyAnalyzed] = useState(false);
  const [highEnergyOnly, setHighEnergyOnly] = useState(false);
  const [sortMode, setSortMode] = useState<'bpm-asc' | 'bpm-desc' | 'energy-desc'>('energy-desc');
  const [selectedId, setSelectedId] = useState(demoTracks[0]?.id);
  const [viewMode, setViewMode] = useState<ViewMode>('library');
  const [playheadPosition, setPlayheadPosition] = useState(0.3);
  const [isPlaying, setIsPlaying] = useState(false);
  const [chartMode, setChartMode] = useState<'bpm' | 'key'>('bpm');

  const filtered = useMemo(() => {
    const base = demoTracks.filter((track) => {
      const matchesQuery =
        track.title.toLowerCase().includes(query.toLowerCase()) ||
        track.artist.toLowerCase().includes(query.toLowerCase()) ||
        track.key.toLowerCase().includes(query.toLowerCase());
      const matchesReview = !onlyReview || track.needsReview;
      const matchesAnalyzed = !onlyAnalyzed || track.status === 'analyzed';
      const matchesEnergy = !highEnergyOnly || track.energy >= 7;
      return matchesQuery && matchesReview && matchesAnalyzed && matchesEnergy;
    });
    return base.sort((a, b) => {
      switch (sortMode) {
        case 'bpm-asc':
          return a.bpm - b.bpm;
        case 'bpm-desc':
          return b.bpm - a.bpm;
        case 'energy-desc':
        default:
          return b.energy - a.energy || b.bpm - a.bpm;
      }
    });
  }, [query, onlyReview, onlyAnalyzed, highEnergyOnly, sortMode]);

  const trackMap = useMemo(
    () =>
      demoTracks.reduce<Record<string, Track>>((acc, t) => {
        acc[t.id] = t;
        return acc;
      }, {}),
    []
  );

  const selected = trackMap[selectedId ?? ''] ?? filtered[0];

  const currentSetIndex = useMemo(() => {
    return demoSetPlan.order.findIndex((id) => id === selectedId);
  }, [selectedId]);

  const currentEdge = useMemo(() => {
    if (currentSetIndex > 0) {
      return demoSetPlan.edges[currentSetIndex - 1];
    }
    return demoSetPlan.edges[0];
  }, [currentSetIndex]);

  const fromTrack = trackMap[currentEdge?.from ?? ''];
  const toTrack = trackMap[currentEdge?.to ?? ''];

  const setEnergyValues = useMemo(() => {
    return demoSetPlan.order.map((id) => trackMap[id]?.energy ?? 5);
  }, [trackMap]);

  const setTrackLabels = useMemo(() => {
    return demoSetPlan.order.map((id) => trackMap[id]?.title ?? id);
  }, [trackMap]);

  const stats = useMemo(() => {
    const analyzed = demoTracks.filter((t) => t.status === 'analyzed').length;
    const pending = demoTracks.filter((t) => t.status === 'pending').length;
    const avgBpm = demoTracks.reduce((acc, t) => acc + t.bpm, 0) / Math.max(demoTracks.length, 1);
    const avgEnergy = demoTracks.reduce((acc, t) => acc + t.energy, 0) / Math.max(demoTracks.length, 1);
    const keys = new Set(demoTracks.map((t) => t.key));
    const avgEdge =
      demoSetPlan.edges.reduce((acc, e) => acc + e.score, 0) / Math.max(demoSetPlan.edges.length, 1);
    return {
      analyzed,
      pending,
      failed: 0,
      avgBpm: Math.round(avgBpm * 10) / 10,
      avgEnergy: Math.round(avgEnergy * 10) / 10,
      keyCount: keys.size,
      avgEdgeScore: Math.round(avgEdge * 10) / 10,
      totalTracks: demoTracks.length,
    };
  }, []);

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

  return (
    <div className="app">
      <header className="app-header">
        <div className="header-left">
          <div className="app-logo">
            <span className="logo-icon">◈</span>
            Algiers
          </div>
          <span className="version-badge">alpha</span>
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
            onClick={() => setViewMode('setBuilder')}
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
                <LiveStats stats={stats} compact />
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
                    <BPMKeyChart tracks={demoTracks} height={100} mode={chartMode} />
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
                    <span className="pill pill-primary">{demoSetPlan.mode}</span>
                  </div>
                  <SetBuilder plan={demoSetPlan} tracks={trackMap} compact />
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
                      <span className="pill pill-primary">{demoSetPlan.mode}</span>
                      <span className="pill pill-analyzed">{demoSetPlan.order.length} tracks</span>
                    </div>
                  </div>
                  <SetBuilder plan={demoSetPlan} tracks={trackMap} />
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
                    edges={demoSetPlan.edges}
                    height={480}
                    selectedTrackId={selectedId}
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
                  <BPMKeyChart tracks={demoTracks} height={100} mode="key" />
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </main>

      <footer className="app-footer">
        <span>Algiers · DJ Set Prep Copilot</span>
        <span className="muted">v0.1.0-alpha · Apple Silicon M1–M5</span>
      </footer>
    </div>
  );
}

export default App;
