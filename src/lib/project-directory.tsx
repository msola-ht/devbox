import { isMyGo } from "mygo-runtime";
import { createContext, useContext, useEffect, useState, type ReactNode } from "react";

import { ProjectDialog } from "@/mygo";
import { getStoredValue, setStoredValue } from "./storage";

const storageKey = "devbox:tool-project";
const ProjectDirectoryContext = createContext<{
  directory: string;
  confirmDirectory: (directory: string) => void;
  initializing: boolean;
  initialError: string;
} | null>(null);

export function ProjectDirectoryProvider({ children }: { children: ReactNode }) {
  const [directory, setDirectory] = useState(() => getStoredValue(storageKey) ?? "");
  const [initializing, setInitializing] = useState(() => isMyGo() && !directory);
  const [initialError, setInitialError] = useState("");

  useEffect(() => {
    if (directory || !isMyGo()) return;
    let active = true;
    void ProjectDialog.createDefaultDirectory()
      .then((path) => {
        if (!active) return;
        setInitializing(false);
        setDirectory(path);
      })
      .catch((error: unknown) => {
        if (!active) return;
        setInitializing(false);
        setInitialError(error instanceof Error ? error.message : "无法准备默认工作目录。");
      });
    return () => { active = false; };
  }, [directory]);

  function confirmDirectory(next: string) {
    setDirectory(next);
    setInitialError("");
    setStoredValue(storageKey, next);
  }

  return (
    <ProjectDirectoryContext value={{ directory, confirmDirectory, initializing, initialError }}>
      {children}
    </ProjectDirectoryContext>
  );
}

export function useProjectDirectory() {
  const context = useContext(ProjectDirectoryContext);
  if (!context) throw new Error("ProjectDirectoryProvider is missing");
  return context;
}
