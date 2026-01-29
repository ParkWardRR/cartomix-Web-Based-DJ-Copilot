import { useMemo, useState } from 'react';
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

  // Get current transition from set plan
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

  // Energy values for set arc visualization
  const setEnergyValues = useMemo(() => {
    return demoSetPlan.order.map((id) => trackMap[id]?.energy ?? 5);
  }, [trackMap]);

  const setTrackLabels = useMemo(() => {
    return demoSetPlan.order.map((id) => trackMap[id]?.title ?? id);
  }, [trackMap]);

  // Stats calculation
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

  // Simulate playhead animation
  const handlePlay = () => {
    setIsPlaying(!isPlaying);
    if (!isPlaying) {
      const interval = setInterval(() => {
        setPlayheadPosition((p) => {
          if (p >= 1) {
            clearInterval(interval);
            setIsPlaying(false);
            return 0;
          }
          return p + 0.002;
        });
      }, 50);
    }
  };

  return (
    <div className="app">
      <header className="app-header">
        <div className="header-left">
          <div className="app-logo">Algiers</div>
          <span className="version-badge">v0.1.0-alpha</span>
        </div>
        <nav className="header-nav">
          <button
            className={`nav-btn ${viewMode === 'library' ? 'active' : ''}`}
            onClick={() => setViewMode('library')}
          >
            Library
          </button>
          <button
            className={`nav-btn ${viewMode === 'setBuilder' ? 'active' : ''}`}
            onClick={() => setViewMode('setBuilder')}
          >
            Set Builder
          </button>
          <button
            className={`nav-btn ${viewMode === 'graph' ? 'active' : ''}`}
            onClick={() => setViewMode('graph')}
          >
            Graph View
          </button>
          <ThemeToggle />
        </nav>
      </header>

      <main className="app-main">
        {/* Hero Stats Section */}
        <motion.section
          className="hero-section"
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
        >
          <div className="hero-header">
            <div>
              <h1 className="hero-title">DJ Set Prep Copilot</h1>
              <p className="hero-subtitle">
                Scan, analyze, propose cue points, and audition transitions before you open your DJ app.
              </p>
            </div>
            <div className="pill-row">
              <span className="pill pill-primary">gRPC Engine</span>
              <span className="pill pill-analyzed">Apple Silicon Accelerated</span>
              <span className="pill pill-secondary">Local-First</span>
            </div>
          </div>

          <LiveStats stats={stats} />

          {/* Charts Row */}
          <div className="charts-row">
            <div className="chart-container">
              <div className="chart-header">
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
              <BPMKeyChart tracks={demoTracks} height={140} mode={chartMode} />
            </div>
            <div className="chart-container">
              <div className="chart-label">Live Spectrum</div>
              <SpectrumAnalyzer height={140} isActive={isPlaying} style="mirror" />
            </div>
          </div>
        </motion.section>

        {/* Main Content */}
        <AnimatePresence mode="wait">
          {viewMode === 'library' && (
            <motion.div
              key="library"
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: 20 }}
              transition={{ duration: 0.3 }}
              className="layout-grid"
            >
              {/* Library Column */}
              <div className="column">
                <div className="card controls">
                  <div className="controls-row">
                    <input
                      className="input"
                      placeholder="Search title, artist, or key..."
                      value={query}
                      onChange={(e) => setQuery(e.target.value)}
                    />
                    <label className="checkbox">
                      <input
                        type="checkbox"
                        checked={onlyReview}
                        onChange={(e) => setOnlyReview(e.target.checked)}
                      />
                      Needs review
                    </label>
                    <label className="checkbox">
                      <input
                        type="checkbox"
                        checked={onlyAnalyzed}
                        onChange={(e) => setOnlyAnalyzed(e.target.checked)}
                      />
                      Analyzed
                    </label>
                    <label className="checkbox">
                      <input
                        type="checkbox"
                        checked={highEnergyOnly}
                        onChange={(e) => setHighEnergyOnly(e.target.checked)}
                      />
                      High Energy
                    </label>
                  </div>
                  <div className="controls-row">
                    <label className="checkbox">
                      <span>Sort:</span>
                      <select
                        className="select"
                        value={sortMode}
                        onChange={(e) =>
                          setSortMode(e.target.value as 'bpm-asc' | 'bpm-desc' | 'energy-desc')
                        }
                      >
                        <option value="energy-desc">Energy (High→Low)</option>
                        <option value="bpm-asc">BPM (Low→High)</option>
                        <option value="bpm-desc">BPM (High→Low)</option>
                      </select>
                    </label>
                    <span className="muted">{filtered.length} tracks</span>
                  </div>
                </div>
                <LibraryGrid tracks={filtered} selectedId={selected?.id} onSelect={setSelectedId} />
              </div>

              {/* Track Detail Column */}
              <div className="column">
                <div className="card">
                  <TrackDetail track={selected} />

                  {/* Waveform Visualization */}
                  {selected && (
                    <div className="waveform-section">
                      <div className="section-header">
                        <h4>Waveform</h4>
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
                      <WaveformCanvas
                        peaks={selected.waveformSummary}
                        sections={selected.sections}
                        cues={selected.cues}
                        playheadPosition={playheadPosition}
                        duration={180}
                        isPlaying={isPlaying}
                        onSeek={setPlayheadPosition}
                        height={100}
                        bpm={selected.bpm}
                      />
                    </div>
                  )}
                </div>

                {/* Mini Spectrum */}
                <div className="card">
                  <div className="section-header">
                    <h4>Frequency Analysis</h4>
                  </div>
                  <SpectrumAnalyzer height={60} isActive={isPlaying} style="bars" colorScheme="accent" bands={24} />
                </div>
              </div>

              {/* Set Builder Column */}
              <div className="column">
                <SetBuilder plan={demoSetPlan} tracks={trackMap} />

                {/* Energy Arc */}
                <div className="card">
                  <div className="section-header">
                    <h4>Set Energy Flow</h4>
                    <span className="pill pill-primary">{demoSetPlan.mode}</span>
                  </div>
                  <EnergyArc
                    values={setEnergyValues}
                    labels={setTrackLabels}
                    height={100}
                    highlightIndex={currentSetIndex >= 0 ? currentSetIndex : undefined}
                  />
                </div>

                <TransitionRehearsal from={fromTrack} to={toTrack} edge={currentEdge} />
              </div>
            </motion.div>
          )}

          {viewMode === 'setBuilder' && (
            <motion.div
              key="setBuilder"
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: 20 }}
              transition={{ duration: 0.3 }}
              className="set-builder-view"
            >
              <div className="set-builder-main">
                <SetBuilder plan={demoSetPlan} tracks={trackMap} />
              </div>
              <div className="set-builder-sidebar">
                {/* Energy Arc - Full Width */}
                <div className="card">
                  <div className="section-header">
                    <h4>Set Energy Journey</h4>
                    <span className="pill pill-primary">{demoSetPlan.mode}</span>
                  </div>
                  <EnergyArc
                    values={setEnergyValues}
                    labels={setTrackLabels}
                    height={140}
                    highlightIndex={currentSetIndex >= 0 ? currentSetIndex : undefined}
                  />
                </div>

                {/* Transition Preview */}
                <TransitionRehearsal from={fromTrack} to={toTrack} edge={currentEdge} />

                {/* Mini waveforms for current transition */}
                {fromTrack && toTrack && (
                  <div className="card">
                    <div className="section-header">
                      <h4>Transition Preview</h4>
                    </div>
                    <div className="dual-waveform">
                      <div className="waveform-mini-label">
                        <span className="pill pill-pending">A</span>
                        {fromTrack.title}
                      </div>
                      <WaveformCanvas
                        peaks={fromTrack.waveformSummary}
                        sections={fromTrack.sections}
                        height={50}
                        playheadPosition={0.85}
                        bpm={fromTrack.bpm}
                      />
                      <div className="waveform-mini-label">
                        <span className="pill pill-analyzed">B</span>
                        {toTrack.title}
                      </div>
                      <WaveformCanvas
                        peaks={toTrack.waveformSummary}
                        sections={toTrack.sections}
                        height={50}
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
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: 20 }}
              transition={{ duration: 0.3 }}
              className="graph-view"
            >
              <div className="graph-main">
                <div className="card">
                  <div className="section-header">
                    <h4>Transition Graph</h4>
                    <span className="muted">Visualize track connections and flow</span>
                  </div>
                  <TransitionGraph
                    tracks={trackMap}
                    edges={demoSetPlan.edges}
                    height={500}
                    selectedTrackId={selectedId}
                    onSelectTrack={setSelectedId}
                  />
                </div>
              </div>
              <div className="graph-sidebar">
                {/* Selected track info */}
                {selected && (
                  <div className="card">
                    <TrackDetail track={selected} />
                    <WaveformCanvas
                      peaks={selected.waveformSummary}
                      sections={selected.sections}
                      cues={selected.cues}
                      height={80}
                      playheadPosition={playheadPosition}
                      bpm={selected.bpm}
                      onSeek={setPlayheadPosition}
                    />
                  </div>
                )}

                {/* Set energy arc */}
                <div className="card">
                  <div className="section-header">
                    <h4>Set Energy</h4>
                  </div>
                  <EnergyArc
                    values={setEnergyValues}
                    labels={setTrackLabels}
                    height={100}
                    highlightIndex={currentSetIndex >= 0 ? currentSetIndex : undefined}
                  />
                </div>

                {/* Distribution charts */}
                <div className="card">
                  <BPMKeyChart tracks={demoTracks} height={120} mode="key" />
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </main>

      <footer className="app-footer">
        <div className="footer-content">
          <span>Algiers · DJ Set Prep Copilot</span>
          <span className="muted">v0.1.0-alpha · Apple Silicon M1-M5</span>
        </div>
      </footer>
    </div>
  );
}

export default App;
