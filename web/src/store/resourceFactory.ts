import { StateCreator } from "zustand";

export type Status = "idle" | "loading" | "error" | "ready";

export interface ResourceSlice<T> {
  items: T[];
  status: Status;
  error?: string | null;
  lastFetched?: number | null;
  fetch: () => Promise<void>;
  refresh: () => Promise<void>;
  create: (payload: any) => Promise<any>;
  update: (id: string, changes: any) => Promise<any>;
  remove: (id: string) => Promise<void>;
  getById: (id: string) => Promise<T | undefined>;
}

export type ResourceFns<T> = {
  list: () => Promise<T[]>;
  create?: (payload: any) => Promise<any>;
  update?: (id: string, changes: any) => Promise<any>;
  remove?: (id: string) => Promise<void>;
  getById?: (id: string) => Promise<T>;
};

export function createResourceSlice<T>(
  name: string,
  fns: ResourceFns<T>,
): StateCreator<any, [], [], ResourceSlice<T>> {
  const { list, create, update, remove, getById } = fns;

  return (set, get) => ({
    items: [] as T[],
    status: "idle" as Status,
    error: null as string | null,
    lastFetched: null as number | null,

    fetch: async () => {
      set({ [name]: { ...get()[name], status: "loading", error: null } });
      try {
        const data = await list();
        set({
          [name]: {
            ...get()[name],
            items: data,
            status: "ready",
            error: null,
            lastFetched: Date.now(),
          },
        });
      } catch (e: any) {
        set({
          [name]: {
            ...get()[name],
            status: "error",
            error: e?.message || String(e),
          },
        });
      }
    },

    refresh: async () => {
      const slice = get()[name] as ResourceSlice<T>;
      if (slice?.status === "loading") return;
      await get()[name].fetch();
    },

    create: async (payload: any) => {
      if (!create) throw new Error(`${name}: create not implemented`);
      await create(payload);
      await get()[name].refresh();
    },

    update: async (id: string, changes: any) => {
      if (!update) throw new Error(`${name}: update not implemented`);
      await update(id, changes);
      await get()[name].refresh();
    },

    remove: async (id: string) => {
      if (!remove) throw new Error(`${name}: remove not implemented`);
      await remove(id);
      await get()[name].refresh();
    },

    getById: async (id: string) => {
      if (!getById) return undefined;
      try {
        const res = await getById(id);
        return res;
      } catch (e) {
        return undefined;
      }
    },
  });
}
