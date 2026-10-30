import { useEffect, useSyncExternalStore } from "react"

import { IpcClientError } from "@/lib/tauri-ipc"

export type SnapshotStatus = "idle" | "loading" | "ready" | "error"

export interface VersionedSnapshot {
  bootId: string
  revision: string
}

export interface SnapshotQueryState<T extends VersionedSnapshot> {
  status: SnapshotStatus
  data: T | null
  error: IpcClientError | null
  isRefreshing: boolean
}

export interface SnapshotStore<T extends VersionedSnapshot> {
  subscribe: (listener: () => void) => () => void
  getState: () => SnapshotQueryState<T>
  apply: (snapshot: T) => void
  ensureLoaded: () => Promise<void>
  refresh: () => Promise<void>
}

export interface SnapshotLoaders<T extends VersionedSnapshot> {
  query: () => Promise<T>
  refresh?: () => Promise<T>
}

export function createSnapshotStore<T extends VersionedSnapshot>({
  query,
  refresh = query,
}: SnapshotLoaders<T>): SnapshotStore<T> {
  let state: SnapshotQueryState<T> = {
    status: "idle",
    data: null,
    error: null,
    isRefreshing: false,
  }
  let currentBootId: string | null = null
  let currentRevision = 0n
  let requestGeneration = 0
  let inFlight: Promise<void> | null = null
  let inFlightIsRefresh = false
  let queuedRefresh: Promise<void> | null = null
  const listeners = new Set<() => void>()

  const publish = (next: SnapshotQueryState<T>) => {
    state = next
    listeners.forEach((listener) => listener())
  }

  const applySnapshot = (snapshot: T) => {
    if (currentBootId !== snapshot.bootId) {
      currentBootId = snapshot.bootId
      currentRevision = 0n
    }

    const revision = BigInt(snapshot.revision)
    if (revision < currentRevision) {
      publish({
        ...state,
        status: state.data ? "ready" : "error",
        error: state.data
          ? null
          : new IpcClientError("STALE_SNAPSHOT", "收到旧版本状态快照"),
        isRefreshing: false,
      })
      return
    }

    currentRevision = revision
    publish({
      status: "ready",
      data: snapshot,
      error: null,
      isRefreshing: false,
    })
  }

  const startLoad = (forceRefresh: boolean) => {
    const requestId = ++requestGeneration
    publish({
      ...state,
      status: "loading",
      error: null,
      isRefreshing: state.data !== null,
    })

    const request = (async () => {
      try {
        const snapshot = forceRefresh ? await refresh() : await query()
        if (requestId === requestGeneration) {
          applySnapshot(snapshot)
        }
      } catch (error) {
        if (requestId === requestGeneration) {
          publish({
            status: "error",
            data: state.data,
            error:
              error instanceof IpcClientError
                ? error
                : new IpcClientError(
                    "IPC_TRANSPORT_FAILED",
                    "无法完成本地请求"
                  ),
            isRefreshing: false,
          })
        }
      } finally {
        if (requestId === requestGeneration) {
          inFlight = null
          inFlightIsRefresh = false
        }
      }
    })()

    inFlight = request
    inFlightIsRefresh = forceRefresh
    return request
  }

  const load = (forceRefresh: boolean) => {
    if (!inFlight) {
      return startLoad(forceRefresh)
    }

    if (!forceRefresh || inFlightIsRefresh) {
      return inFlight
    }

    if (!queuedRefresh) {
      queuedRefresh = inFlight
        .then(
          () => startLoad(true),
          () => startLoad(true)
        )
        .finally(() => {
          queuedRefresh = null
        })
    }

    return queuedRefresh
  }

  return {
    subscribe(listener) {
      listeners.add(listener)
      return () => listeners.delete(listener)
    },
    getState() {
      return state
    },
    apply(snapshot) {
      applySnapshot(snapshot)
    },
    ensureLoaded() {
      if (state.status === "idle" || state.status === "error") {
        return load(false)
      }
      return Promise.resolve()
    },
    refresh() {
      return load(true)
    },
  }
}

export function useSnapshotStore<T extends VersionedSnapshot>(
  store: SnapshotStore<T>
) {
  const state = useSyncExternalStore(
    store.subscribe,
    store.getState,
    store.getState
  )

  useEffect(() => {
    void store.ensureLoaded()
  }, [store])

  return {
    state,
    refresh: store.refresh,
  }
}
