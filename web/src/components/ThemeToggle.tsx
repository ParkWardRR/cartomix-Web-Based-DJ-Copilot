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
