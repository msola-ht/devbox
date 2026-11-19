export type Theme = "light" | "dark";

const storageKey = "color-scheme";

export function getTheme(): Theme {
  return document.documentElement.classList.contains("dark") ? "dark" : "light";
}

export function applyTheme(theme: Theme) {
  document.documentElement.classList.toggle("dark", theme === "dark");
  document
    .querySelector('meta[name="color-scheme"]')
    ?.setAttribute("content", theme);

  try {
    localStorage.setItem(storageKey, theme);
  } catch {
    // The selected theme still applies when storage is unavailable.
  }
}
