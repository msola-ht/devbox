import { FolderOpenIcon, RotateCcwIcon } from "lucide-react";
import { useEffect, useId, useState } from "react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Field, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useProjectDirectory } from "@/lib/project-directory";
import { ProjectDialog } from "@/mygo";

export default function ProjectDirectoryBar() {
  const id = useId();
  const { directory, confirmDirectory, initializing, initialError } = useProjectDirectory();
  const [draft, setDraft] = useState<{ path: string; isDefault: boolean } | null>(null);
  const [operation, setOperation] = useState<"selecting" | "confirming" | "opening" | null>(null);
  const [error, setError] = useState("");
  const [defaultPath, setDefaultPath] = useState("");
  const busy = initializing || operation !== null;

  useEffect(() => {
    const controller = new AbortController();
    void ProjectDialog.defaultDirectory.with({ signal: controller.signal })()
      .then((path) => {
        if (!controller.signal.aborted) setDefaultPath(path);
      })
      .catch(() => {});
    return () => controller.abort();
  }, []);

  async function openDefaultDirectory() {
    setOperation("opening");
    setError("");
    try {
      await ProjectDialog.openDefaultDirectory();
    } catch (error) {
      setError(error instanceof Error ? error.message : "无法打开默认目录。");
    } finally {
      setOperation(null);
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
    setOperation("confirming");
    setError("");
    try {
      let path = draft.path;
      if (draft.isDefault) {
        path = await ProjectDialog.createDefaultDirectory();
      } else {
        await ProjectDialog.validateDirectory(path);
      }
      confirmDirectory(path);
      setDraft(null);
    } catch (error) {
      setError(error instanceof Error ? error.message : "无法准备工作目录。");
    } finally {
      setOperation(null);
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
              <Button type="submit" disabled={busy}>
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
              <Button type="button" variant="outline" disabled={busy} onClick={() => void selectDirectory()}>
                {operation === "selecting" ? "选择中…" : "选择目录"}
              </Button>
              <Tooltip>
                <TooltipTrigger render={
                  <Button type="button" variant="outline" size="icon" disabled={busy}
                    aria-label="打开默认目录" onClick={() => void openDefaultDirectory()}>
                    <FolderOpenIcon aria-hidden />
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
      </Field>
    </form>
  );
}
