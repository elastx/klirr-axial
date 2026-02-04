import { create } from "zustand";
import { APIService } from "../services/api";
import { createResourceSlice, ResourceSlice } from "./resourceFactory";
import type {
  BulletinPost,
  Group,
  Message,
  StoredUser,
  User,
} from "../types";
import type { UserInfo } from "../services/gpg";

// Store shape
export interface AppStore {
  users: ResourceSlice<User> & {
    register: (userInfo: UserInfo) => Promise<void>;
    search: (q: string, limit?: number, offset?: number) => Promise<StoredUser[]>;
    recent: (limit?: number) => Promise<StoredUser[]>;
    // Resolve canonical signing fingerprint from any encryption key ID
    resolveSigningFingerprint: (encryptionId: string) => string | undefined;
  };
  groups: ResourceSlice<Group>;
  messages: ResourceSlice<Message> & {
    send: (recipientFingerprint: string, content: string) => Promise<void>;
  };
  bulletin: ResourceSlice<BulletinPost> & {
    post: (topic: string, content: string, parentId?: string) => Promise<void>;
  };
  topics: ResourceSlice<string>;
  files: ResourceSlice<any> & {
    upload: (
      file: File,
      uploaderFingerprint: string,
      description?: string,
      encrypted?: boolean,
      recipients?: string[],
    ) => Promise<any>;
    remove: (fileId: string) => Promise<void>;
  };

  // Polling controls
  startPolling: () => void;
  stopPolling: () => void;
}

const api = APIService.getInstance();

export const useAppStore = create<AppStore>()((set, get) => {
  // Slices via shared factory (CRUD)
  const usersSlice = createResourceSlice<User>("users", {
    list: () => api.getUsers(),
    create: (userInfo: UserInfo) => api.registerUser(userInfo),
    getById: (fp: string) => api.getUser(fp),
  });

  const groupsSlice = createResourceSlice<Group>("groups", {
    list: () => api.getGroups(),
    create: ({ user_id, private_key }: { user_id: string; private_key: string }) =>
      api.registerGroup(user_id, private_key),
  });

  const messagesSlice = createResourceSlice<Message>("messages", {
    list: () => api.getMessages(),
    create: ({ recipientFingerprint, content }: { recipientFingerprint: string; content: string }) =>
      api.sendPrivateMessage(recipientFingerprint, content),
  });

  const bulletinSlice = createResourceSlice<BulletinPost>("bulletin", {
    list: () => api.getBulletinPosts(),
    create: ({ topic, content, parentId }: { topic: string; content: string; parentId?: string }) =>
      api.sendBulletinPost(topic, content, parentId),
  });

  const topicsSlice = createResourceSlice<string>("topics", {
    list: () => api.getTopics(),
  });

  const filesSlice = createResourceSlice<any>("files", {
    list: () => api.getFiles(undefined, undefined, 100, 0),
    create: ({ file, uploaderFingerprint, description, encrypted, recipients }: { file: File; uploaderFingerprint: string; description?: string; encrypted?: boolean; recipients?: string[] }) =>
      api.uploadFile(file, uploaderFingerprint, description, encrypted, recipients),
    remove: (id: string) => api.deleteFile(id),
    getById: (id: string) => api.getFileMetadata(id),
  });

  // Poller instance stored outside state to avoid re-renders
  let poller: number | null = null;

  return {
    // Attach base resource slices
    users: {
      ...usersSlice(set, get, undefined as any),
      register: async (userInfo: UserInfo) => {
        await get().users.create(userInfo);
      },
      search: (q: string, limit = 20, offset = 0) => api.searchUsers(q, limit, offset),
      recent: (limit = 10) => api.getRecentUsers(limit),
      resolveSigningFingerprint: (encryptionId: string) => {
        const users = get().users.items || [];
        const found = users.find((u) =>
          (u.encryption_fingerprints || []).includes(encryptionId)
        );
        return found
          ? found.signing_fingerprint || found.fingerprint
          : undefined;
      },
    },

    groups: {
      ...groupsSlice(set, get, undefined as any),
    },

    messages: {
      ...messagesSlice(set, get, undefined as any),
      send: async (recipientFingerprint: string, content: string) => {
        await get().messages.create({ recipientFingerprint, content });
      },
    },

    bulletin: {
      ...bulletinSlice(set, get, undefined as any),
      post: async (topic: string, content: string, parentId?: string) => {
        await get().bulletin.create({ topic, content, parentId });
        await get().topics.refresh();
      },
    },

    topics: {
      ...topicsSlice(set, get, undefined as any),
    },

    files: {
      ...filesSlice(set, get, undefined as any),
      upload: async (
        file: File,
        uploaderFingerprint: string,
        description?: string,
        encrypted?: boolean,
        recipients?: string[],
      ) => {
        const res = await get().files.create({ file, uploaderFingerprint, description, encrypted, recipients });
        return res;
      },
      remove: async (fileId: string) => {
        await get().files.remove(fileId);
      },
    },

    // Polling controls
    startPolling: () => {
      if (poller !== null) return;
      // Initial fetch to populate cache quickly
      void get().users.refresh();
      void get().groups.refresh();
      void get().messages.refresh();
      void get().bulletin.refresh();
      void get().topics.refresh();
      void get().files.refresh();

      poller = window.setInterval(() => {
        // Keep data fresh every 10 seconds; backend is the source of truth
        void get().users.refresh();
        void get().groups.refresh();
        void get().messages.refresh();
        void get().bulletin.refresh();
        void get().topics.refresh();
        void get().files.refresh();
      }, 10_000);
    },
    stopPolling: () => {
      if (poller !== null) {
        clearInterval(poller);
        poller = null;
      }
    },
  };
});

// Auto-start polling on first import to honor "periodically ensure up-to-date"
// Consumers can call `stopPolling()` if needed (e.g., during teardown)
useAppStore.getState().startPolling();
