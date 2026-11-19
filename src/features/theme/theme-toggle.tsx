import { MoonIcon, SunIcon } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

import { applyTheme, getTheme } from "./theme";

export function ThemeToggle() {
  const [theme, setTheme] = useState(getTheme);
  const isDark = theme === "dark";
  const label = isDark ? "切换到浅色模式" : "切换到深色模式";

  function toggleTheme() {
    const nextTheme = isDark ? "light" : "dark";
    applyTheme(nextTheme);
    setTheme(nextTheme);
  }

  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label={label}
            aria-pressed={isDark}
            onClick={toggleTheme}
          />
        }
      >
        {isDark ? (
          <SunIcon data-icon="inline-start" />
        ) : (
          <MoonIcon data-icon="inline-start" />
        )}
      </TooltipTrigger>
      <TooltipContent>{label}</TooltipContent>
    </Tooltip>
  );
}
