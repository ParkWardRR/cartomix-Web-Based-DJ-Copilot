import { useEffect } from 'react';
import { motion } from 'framer-motion';
import { useStore } from '../store';

interface ToggleProps {
  enabled: boolean;
  onChange: (enabled: boolean) => void;
  disabled?: boolean;
}

function Toggle({ enabled, onChange, disabled }: ToggleProps) {
  return (
    <button
      className={`toggle ${enabled ? 'enabled' : ''} ${disabled ? 'disabled' : ''}`}
      onClick={() => !disabled && onChange(!enabled)}
      disabled={disabled}
    >
      <motion.div
        className="toggle-knob"
        animate={{ x: enabled ? 20 : 0 }}
        transition={{ type: 'spring', stiffness: 500, damping: 30 }}
      />
    </button>
  );
}

interface SettingRowProps {
  title: string;
  description: string;
  badge?: string;
  badgeColor?: string;
  children: React.ReactNode;
}

function SettingRow({ title, description, badge, badgeColor, children }: SettingRowProps) {
  return (
    <div className="setting-row">
      <div className="setting-info">
        <div className="setting-title">
          {title}
          {badge && (
            <span className="setting-badge" style={{ backgroundColor: badgeColor }}>
              {badge}
            </span>
          )}
        </div>
        <div className="setting-description">{description}</div>
      </div>
      <div className="setting-control">{children}</div>
    </div>
  );
}

export function ModelSettings() {
  const {
    mlSettings,
    mlSettingsLoading,
    fetchMLSettings,
    updateMLSettings,
    apiAvailable,
  } = useStore();

  useEffect(() => {
    if (apiAvailable) {
      fetchMLSettings();
    }
  }, [apiAvailable, fetchMLSettings]);

  if (!apiAvailable) {
    return (
      <div className="model-settings">
        <div className="settings-empty">
          <span className="empty-icon">⚙️</span>
          <p>ML settings require API connection</p>
          <span className="empty-hint">Start the Go engine to configure ML features</span>
        </div>
      </div>
    );
  }

  return (
    <div className="model-settings">
      <div className="settings-header">
        <h3>ML Features</h3>
        <span className="settings-badge local">100% Local</span>
      </div>

      <div className="settings-section">
        <div className="section-header">
          <h4>Embedding Models</h4>
          <span className="section-hint">Audio analysis for similarity matching</span>
        </div>

        <SettingRow
          title="OpenL3 Embeddings"
          description="512-dimensional audio embeddings for vibe-based similarity matching. Runs on Apple Neural Engine."
          badge="ANE"
          badgeColor="#ff9500"
        >
          <Toggle
            enabled={mlSettings.openl3Enabled}
            onChange={(enabled) => updateMLSettings({ openl3Enabled: enabled })}
            disabled={mlSettingsLoading}
          />
        </SettingRow>

        <SettingRow
          title="Apple SoundAnalysis"
          description="Built-in Apple classifier for audio context (music/speech/noise). Generates QA flags for review."
          badge="Built-in"
          badgeColor="#34c759"
        >
          <Toggle
            enabled={mlSettings.soundAnalysisEnabled}
            onChange={(enabled) => updateMLSettings({ soundAnalysisEnabled: enabled })}
            disabled={mlSettingsLoading}
          />
        </SettingRow>

        <SettingRow
          title="Custom DJ Section Model"
          description="User-trained model for Intro/Build/Drop/Break/Outro classification. Requires training data."
          badge="Opt-in"
          badgeColor="#8b5cf6"
        >
          <Toggle
            enabled={mlSettings.customModelEnabled}
            onChange={(enabled) => updateMLSettings({ customModelEnabled: enabled })}
            disabled={mlSettingsLoading}
          />
        </SettingRow>
      </div>

      <div className="settings-section">
        <div className="section-header">
          <h4>Similarity Search</h4>
          <span className="section-hint">Configure how similar tracks are found</span>
        </div>

        <SettingRow
          title="Minimum Similarity"
          description="Only show tracks with similarity score above this threshold"
        >
          <div className="slider-control">
            <input
              type="range"
              min="0"
              max="100"
              value={mlSettings.minSimilarityThreshold * 100}
              onChange={(e) =>
                updateMLSettings({
                  minSimilarityThreshold: parseInt(e.target.value) / 100,
                })
              }
              disabled={mlSettingsLoading}
            />
            <span className="slider-value">
              {Math.round(mlSettings.minSimilarityThreshold * 100)}%
            </span>
          </div>
        </SettingRow>

        <SettingRow
          title="Show Explanations"
          description="Display detailed breakdown of why tracks are similar (vibe, tempo, key, energy)"
        >
          <Toggle
            enabled={mlSettings.showExplanations}
            onChange={(enabled) => updateMLSettings({ showExplanations: enabled })}
            disabled={mlSettingsLoading}
          />
        </SettingRow>
      </div>

      <div className="settings-section">
        <div className="section-header">
          <h4>Hardware Acceleration</h4>
          <span className="section-hint">Apple Silicon optimization status</span>
        </div>

        <div className="hardware-grid">
          <div className="hardware-item">
            <div className="hardware-icon ane">⚡</div>
            <div className="hardware-info">
              <span className="hardware-name">Neural Engine</span>
              <span className="hardware-status active">Active</span>
            </div>
          </div>
          <div className="hardware-item">
            <div className="hardware-icon metal">🔷</div>
            <div className="hardware-info">
              <span className="hardware-name">Metal GPU</span>
              <span className="hardware-status active">Active</span>
            </div>
          </div>
          <div className="hardware-item">
            <div className="hardware-icon accelerate">📊</div>
            <div className="hardware-info">
              <span className="hardware-name">Accelerate vDSP</span>
              <span className="hardware-status active">Active</span>
            </div>
          </div>
          <div className="hardware-item">
            <div className="hardware-icon uma">💾</div>
            <div className="hardware-info">
              <span className="hardware-name">Unified Memory</span>
              <span className="hardware-status active">Zero-copy</span>
            </div>
          </div>
        </div>
      </div>

      <div className="settings-footer">
        <div className="footer-badge">
          <span className="badge-icon">🔒</span>
          <span className="badge-text">All processing is 100% local. No data leaves your Mac.</span>
        </div>
      </div>
    </div>
  );
}
