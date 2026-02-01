import { useEffect } from 'react';
import { useTheme, type Theme } from '../hooks/useTheme';

const themeLabels: Record<Theme, string> = {
  light: 'Light',
  dark: 'Dark',
  auto: 'Auto',
};

const themeIcons: Record<Theme, string> = {
  light: '\u2600', // sun
  dark: '\u263E', // moon
  auto: '\u25D0', // circle with left half black
};

export function ThemeToggle() {
  const { theme, toggleTheme } = useTheme();

  // Allow global custom event toggling (used by command palette)
  useEffect(() => {
    const handler = () => toggleTheme();
    window.addEventListener('toggle-theme', handler as EventListener);
    return () => window.removeEventListener('toggle-theme', handler as EventListener);
  }, [toggleTheme]);

  return (
    <button
      onClick={toggleTheme}
      className="theme-toggle"
      aria-label={`Current theme: ${themeLabels[theme]}. Click to toggle.`}
      title={`Theme: ${themeLabels[theme]}`}
    >
      <span className="theme-toggle-icon">{themeIcons[theme]}</span>
      <span className="theme-toggle-label">{themeLabels[theme]}</span>
    </button>
  );
}
