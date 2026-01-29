import { useState } from 'react';
import { useStore, type ExportFormat } from '../store';

type Props = {
  trackIds: string[];
  playlistName?: string;
};

export function ExportPanel({ trackIds, playlistName = 'dj-set' }: Props) {
  const { apiAvailable, isExporting, exportResult, exportError, exportSet, clearExportResult } = useStore();
  const [selectedFormats, setSelectedFormats] = useState<ExportFormat[]>(['rekordbox']);

  const toggleFormat = (format: ExportFormat) => {
    setSelectedFormats((prev) =>
      prev.includes(format) ? prev.filter((f) => f !== format) : [...prev, format]
    );
  };

  const handleExport = async () => {
    if (trackIds.length === 0) return;
    await exportSet(trackIds, playlistName, selectedFormats);
  };

  if (!apiAvailable) {
    return (
      <div className="export-panel">
        <div className="panel-header">
          <h3>Export</h3>
        </div>
        <div className="export-demo-notice">
          <p>Export is available when connected to the API.</p>
          <p className="muted">Start the backend server to enable exports.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="export-panel">
      <div className="panel-header">
        <h3>Export</h3>
        <span className="count-badge">{trackIds.length} tracks</span>
      </div>

      <div className="export-formats">
        <div className="export-format-label">DJ Software Formats</div>
        <div className="export-format-buttons">
          <button
            className={`export-format-btn ${selectedFormats.includes('rekordbox') ? 'selected' : ''}`}
            onClick={() => toggleFormat('rekordbox')}
          >
            <span className="format-icon">RB</span>
            Rekordbox
          </button>
          <button
            className={`export-format-btn ${selectedFormats.includes('serato') ? 'selected' : ''}`}
            onClick={() => toggleFormat('serato')}
          >
            <span className="format-icon">SJ</span>
            Serato
          </button>
          <button
            className={`export-format-btn ${selectedFormats.includes('traktor') ? 'selected' : ''}`}
            onClick={() => toggleFormat('traktor')}
          >
            <span className="format-icon">TK</span>
            Traktor
          </button>
        </div>
        <p className="export-hint">
          Generic M3U + JSON + cues CSV always included
        </p>
      </div>

      <button
        className="export-btn primary"
        onClick={handleExport}
        disabled={isExporting || trackIds.length === 0}
      >
        {isExporting ? (
          <>
            <span className="export-spinner" />
            Exporting...
          </>
        ) : (
          <>Export Set</>
        )}
      </button>

      {exportError && (
        <div className="export-error">
          <span className="error-icon">!</span>
          {exportError}
        </div>
      )}

      {exportResult && (
        <div className="export-result">
          <div className="export-success">
            <span className="success-icon">✓</span>
            Export complete
          </div>
          <div className="export-files">
            <div className="export-file">
              <span className="file-icon">♪</span>
              <span className="file-name">Playlist (M3U8)</span>
            </div>
            <div className="export-file">
              <span className="file-icon">{ }</span>
              <span className="file-name">Analysis (JSON)</span>
            </div>
            <div className="export-file">
              <span className="file-icon">◇</span>
              <span className="file-name">Cue points (CSV)</span>
            </div>
            {exportResult.vendorExports.map((path, idx) => (
              <div key={idx} className="export-file vendor">
                <span className="file-icon">★</span>
                <span className="file-name">{path.split('/').pop()}</span>
              </div>
            ))}
          </div>
          <div className="export-path">
            <span className="muted">Saved to:</span>
            <code>{exportResult.bundlePath || exportResult.playlistPath.replace(/\/[^/]+$/, '')}</code>
          </div>
          <button className="export-btn secondary" onClick={clearExportResult}>
            Done
          </button>
        </div>
      )}
    </div>
  );
}
