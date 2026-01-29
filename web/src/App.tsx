import { useMemo, useState } from 'react';
import { ThemeToggle } from './components/ThemeToggle';
import { LibraryGrid } from './components/LibraryGrid';
import { TrackDetail } from './components/TrackDetail';
import { SetBuilder } from './components/SetBuilder';
import { TransitionRehearsal } from './components/TransitionRehearsal';
import { demoSetPlan, demoTracks } from './mockData';
import type { Track } from './types';

function App() {
  const [query, setQuery] = useState('');
  const [onlyReview, setOnlyReview] = useState(false);
  const [onlyAnalyzed, setOnlyAnalyzed] = useState(false);
  const [highEnergyOnly, setHighEnergyOnly] = useState(false);
  const [selectedId, setSelectedId] = useState(demoTracks[0]?.id);

  const filtered = useMemo(() => {
    return demoTracks.filter((track) => {
      const matchesQuery =
        track.title.toLowerCase().includes(query.toLowerCase()) ||
        track.artist.toLowerCase().includes(query.toLowerCase()) ||
        track.key.toLowerCase().includes(query.toLowerCase());
      const matchesReview = !onlyReview || track.needsReview;
      const matchesAnalyzed = !onlyAnalyzed || track.status === 'analyzed';
      const matchesEnergy = !highEnergyOnly || track.energy >= 7;
      return matchesQuery && matchesReview && matchesAnalyzed && matchesEnergy;
    });
  }, [query, onlyReview, onlyAnalyzed, highEnergyOnly]);

  const trackMap = useMemo(
    () =>
      filtered.reduce<Record<string, Track>>((acc, t) => {
        acc[t.id] = t;
        return acc;
      }, {}),
    [filtered]
  );

  const selected = trackMap[selectedId ?? ''] ?? filtered[0];
  const firstEdge = demoSetPlan.edges[0];
  const fromTrack = trackMap[firstEdge?.from ?? ''];
  const toTrack = trackMap[firstEdge?.to ?? ''];

  const stats = useMemo(() => {
    const analyzed = demoTracks.filter((t) => t.status === 'analyzed').length;
    const pending = demoTracks.filter((t) => t.status === 'pending').length;
    const avgBpm =
      demoTracks.reduce((acc, t) => acc + t.bpm, 0) / Math.max(demoTracks.length, 1);
    const keys = new Set(demoTracks.map((t) => t.key));
    return { analyzed, pending, avgBpm: Math.round(avgBpm * 10) / 10, keys: keys.size };
  }, []);

  return (
    <div className="app">
      <header className="app-header">
        <div className="app-logo">Algiers</div>
        <nav>
          <ThemeToggle />
        </nav>
      </header>

      <main className="app-main">
        <section className="card hero">
          <div>
            <h2>DJ Set Prep Copilot</h2>
            <p className="muted">
              Scan, analyze, propose cue points, and audition transitions before you open your DJ
              app. This demo shows the library, track detail, set builder, and rehearsal views.
            </p>
            <div className="pill-row">
              <span className="pill pill-primary">gRPC engine ready</span>
              <span className="pill pill-secondary">CPU fallback analyzer</span>
              <span className="pill pill-secondary">Generic exports (M3U/JSON/CSV)</span>
            </div>
            <div className="stat-row">
              <div className="stat">
                <div className="stat-label">Analyzed</div>
                <div className="stat-value">{stats.analyzed}</div>
              </div>
              <div className="stat">
                <div className="stat-label">Pending</div>
                <div className="stat-value">{stats.pending}</div>
              </div>
              <div className="stat">
                <div className="stat-label">Avg BPM</div>
                <div className="stat-value">{stats.avgBpm}</div>
              </div>
              <div className="stat">
                <div className="stat-label">Keys</div>
                <div className="stat-value">{stats.keys}</div>
              </div>
            </div>
          </div>
        </section>

        <div className="layout-grid">
          <div className="column">
            <div className="card controls">
              <div className="controls-row">
                <input
                  className="input"
                  placeholder="Search title, artist, or key…"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                />
                <label className="checkbox">
                  <input
                    type="checkbox"
                    checked={onlyReview}
                    onChange={(e) => setOnlyReview(e.target.checked)}
                  />
                  Needs grid review
                </label>
                <label className="checkbox">
                  <input
                    type="checkbox"
                    checked={onlyAnalyzed}
                    onChange={(e) => setOnlyAnalyzed(e.target.checked)}
                  />
                  Analyzed only
                </label>
                <label className="checkbox">
                  <input
                    type="checkbox"
                    checked={highEnergyOnly}
                    onChange={(e) => setHighEnergyOnly(e.target.checked)}
                  />
                  Energy 7+
                </label>
              </div>
              <div className="muted">{filtered.length} tracks shown</div>
            </div>
            <LibraryGrid tracks={filtered} selectedId={selected?.id} onSelect={setSelectedId} />
          </div>

          <div className="column">
            <TrackDetail track={selected} />
          </div>

          <div className="column">
            <SetBuilder plan={demoSetPlan} tracks={trackMap} />
            <TransitionRehearsal from={fromTrack} to={toTrack} edge={firstEdge} />
          </div>
        </div>
      </main>

      <footer className="app-footer">
        <p>Algiers · DJ Set Prep Copilot</p>
      </footer>
    </div>
  );
}

export default App;
