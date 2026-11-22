import { FolderOpenIcon, RotateCcwIcon } from "lucide-react";
import { useCallback, useEffect, useId, useRef, useState } from "react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Field, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useLatestRequest } from "@/hooks/use-latest-request";
import { getStoredValue, setStoredValue } from "@/lib/storage";
import { ProjectDialog } from "@/mygo";

const storageKey = "devbox:tool-project";

export default function ProjectDirectoryBar() {
  const id = useId();
  const [directory, setDirectory] = useState(() => getStoredValue(storageKey) ?? "");
  const [initializing, setInitializing] = useState(!directory);
  const [initialError, setInitialError] = useState("");
  const [draft, setDraft] = useState<{ path: string; isDefault: boolean } | null>(null);
  const [operation, setOperation] = useState<"selecting" | "confirming" | "opening" | null>(null);
  const [error, setError] = useState("");
  const [defaultPath, setDefaultPath] = useState("");
  const [defaultPathError, setDefaultPathError] = useState("");
  const startDefaultPathRequest = useLatestRequest();
  const startDirectoryRequest = useLatestRequest();
  const selectButton = useRef<HTMLButtonElement>(null);
  const confirmButton = useRef<HTMLButtonElement>(null);
  const hadDraft = useRef(false);
  const busy = initializing || operation !== null;

  useEffect(() => {
    if (directory) return;
    const signal = startDirectoryRequest();
    void ProjectDialog.createDefaultDirectory.with({ signal })().then((path) => {
      if (!signal.aborted) setDirectory(path);
    }).catch((error: unknown) => {
      if (!signal.aborted) {
        setInitialError(error instanceof Error ? error.message : "无法准备默认工作目录。");
      }
    }).finally(() => {
      if (!signal.aborted) setInitializing(false);
    });
  }, [directory, startDirectoryRequest]);

  const readDefaultPath = useCallback(async () => {
    const signal = startDefaultPathRequest();
    setDefaultPathError("");
    try {
      const path = await ProjectDialog.defaultDirectory.with({ signal })();
      if (!signal.aborted) setDefaultPath(path);
    } catch {
      if (!signal.aborted) setDefaultPathError("无法读取默认目录位置。");
    }
  }, [startDefaultPathRequest]);

  useEffect(() => { void readDefaultPath(); }, [readDefaultPath]);

  useEffect(() => {
    if (draft) confirmButton.current?.focus();
    else if (hadDraft.current) selectButton.current?.focus();
    hadDraft.current = draft !== null;
  }, [draft]);

  async function openDefaultDirectory() {
    const signal = startDirectoryRequest();
    setOperation("opening");
    setError("");
    try {
      const path = await ProjectDialog.openDefaultDirectory.with({ signal })();
      if (signal.aborted) return;
      setDefaultPath(path);
      setDefaultPathError("");
      if (!directory) {
        setDirectory(path);
        setInitialError("");
      }
    } catch (error) {
      if (!signal.aborted) setError(error instanceof Error ? error.message : "无法打开默认目录。");
    } finally {
      if (!signal.aborted) setOperation(null);
    }
  }

  async function selectDirectory() {
    setOperation("selecting");
    setError("");
    try {
      const path = await ProjectDialog.selectDirectory();
      if (path && path !== directory) setDraft({ path, isDefault: false });
    } catch (error) {
      setError(error instanceof Error ? error.message : "无法选择工作目录。");
    } finally {
      setOperation(null);
    }
  }

  async function applyDirectory() {
    if (!draft || busy) return;
    const signal = startDirectoryRequest();
    setOperation("confirming");
    setError("");
    try {
      let path = draft.path;
      if (draft.isDefault) {
        path = await ProjectDialog.createDefaultDirectory.with({ signal })();
      } else {
        await ProjectDialog.validateDirectory.with({ signal })(path);
      }
      if (signal.aborted) return;
      setStoredValue(storageKey, path);
      setDirectory(path);
      setInitialError("");
      if (draft.isDefault) setDefaultPath(path);
      setDraft(null);
    } catch (error) {
      if (!signal.aborted) setError(error instanceof Error ? error.message : "无法准备工作目录。");
    } finally {
      if (!signal.aborted) setOperation(null);
    }
  }

  return (
    <form className="min-w-0" onSubmit={(event) => {
      event.preventDefault();
      void applyDirectory();
    }}>
      <Field>
        <FieldLabel htmlFor={id}>{draft ? "待确认的工作目录" : "项目工作目录"}</FieldLabel>
        <div className="flex flex-wrap items-center gap-2">
          <Input
            id={id}
            name="projectDirectory"
            value={draft?.path ?? directory}
            placeholder={initializing ? "正在准备默认工作目录…" : "尚未选择"}
            readOnly
            className="min-w-40 flex-1"
          />
          {draft ? (
            <>
              <Button ref={confirmButton} type="submit" disabled={busy} focusableWhenDisabled>
                {operation === "confirming" ? "确认中…" : "确认目录"}
              </Button>
              <Button type="button" variant="ghost" disabled={busy} onClick={() => {
                setDraft(null);
                setError("");
              }}>
                取消
              </Button>
            </>
          ) : (
            <div className="flex shrink-0 items-center gap-2">
              <Button ref={selectButton} type="button" variant="outline" disabled={busy} focusableWhenDisabled
                onClick={() => void selectDirectory()}>
                {operation === "selecting" ? "选择中…" : "选择目录"}
              </Button>
              <Tooltip>
                <TooltipTrigger render={
                  <Button type="button" variant="outline" size="icon" disabled={busy} focusableWhenDisabled
                    aria-label="打开默认目录" onClick={() => void openDefaultDirectory()}>
                    {operation === "opening" ? <Spinner aria-hidden /> : <FolderOpenIcon aria-hidden />}
                  </Button>
                } />
                <TooltipContent>打开默认目录</TooltipContent>
              </Tooltip>
              {defaultPath && directory !== defaultPath && (
                <Tooltip>
                  <TooltipTrigger render={
                    <Button type="button" variant="ghost" size="icon" disabled={busy}
                      aria-label="恢复默认目录" onClick={() => {
                        setError("");
                        setDraft({ path: defaultPath, isDefault: true });
                      }}>
                      <RotateCcwIcon aria-hidden />
                    </Button>
                  } />
                  <TooltipContent>恢复默认目录</TooltipContent>
                </Tooltip>
              )}
            </div>
          )}
        </div>
        {(error || initialError) && (
          <Alert variant="destructive"><AlertDescription>{error || initialError}</AlertDescription></Alert>
        )}
        {defaultPathError && !error && !initialError && (
          <Alert variant="destructive">
            <AlertDescription className="flex flex-wrap items-center gap-2">
              {defaultPathError}
              <Button type="button" size="sm" variant="outline" disabled={busy}
                onClick={() => void readDefaultPath()}>
                重试
              </Button>
            </AlertDescription>
          </Alert>
        )}
      </Field>
    </form>
  );
}
