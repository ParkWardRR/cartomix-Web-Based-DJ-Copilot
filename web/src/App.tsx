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
  const [selectedId, setSelectedId] = useState(demoTracks[0]?.id);

  const filtered = useMemo(() => {
    return demoTracks.filter((track) => {
      const matchesQuery =
        track.title.toLowerCase().includes(query.toLowerCase()) ||
        track.artist.toLowerCase().includes(query.toLowerCase()) ||
        track.key.toLowerCase().includes(query.toLowerCase());
      const matchesReview = !onlyReview || track.needsReview;
      return matchesQuery && matchesReview;
    });
  }, [query, onlyReview]);

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
