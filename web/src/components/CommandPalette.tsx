import { useMemo } from 'react';

type Action = {
  group: string;
  label: string;
  hint?: string;
  active?: boolean;
  disabled?: boolean;
  run: () => void;
};

type Props = {
  open: boolean;
  query: string;
  onQueryChange: (q: string) => void;
  onClose: () => void;
  actions: Action[];
};

export function CommandPalette({ open, query, onQueryChange, onClose, actions }: Props) {
  const filtered = useMemo(() => {
    const q = query.toLowerCase().trim();
    if (!q) return actions;
    return actions.filter((a) => a.label.toLowerCase().includes(q) || a.group.toLowerCase().includes(q));
  }, [actions, query]);

  const handleSelect = (action: Action) => {
    if (action.disabled) return;
    action.run();
    onClose();
  };

  if (!open) return null;

  return (
    <div className="cmdp-backdrop" onClick={onClose}>
      <div className="cmdp" onClick={(e) => e.stopPropagation()}>
        <div className="cmdp-header">
          <input
            autoFocus
            value={query}
            onChange={(e) => onQueryChange(e.target.value)}
            placeholder="Type a command or search..."
          />
          <button className="cmdp-close" onClick={onClose} aria-label="Close command palette">×</button>
        </div>
        <div className="cmdp-list">
          {filtered.map((action, idx) => (
            <button
              key={`${action.group}-${action.label}-${idx}`}
              className={`cmdp-item ${action.disabled ? 'disabled' : ''} ${action.active ? 'active' : ''}`}
              onClick={() => handleSelect(action)}
            >
              <div className="cmdp-text">
                <span className="cmdp-label">{action.label}</span>
                <span className="cmdp-group">{action.group}</span>
              </div>
              <div className="cmdp-hint">
                {action.hint && <kbd>{action.hint}</kbd>}
              </div>
            </button>
          ))}
          {!filtered.length && (
            <div className="cmdp-empty">
              <span>Nothing found.</span>
              <small>Try "library", "analyze", or "theme".</small>
            </div>
          )}
        </div>
        <div className="cmdp-footer">
          <span>Press ⎋ to close · ⌘K anytime</span>
        </div>
      </div>
    </div>
  );
}
