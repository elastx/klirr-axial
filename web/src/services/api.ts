import axios from "axios";
import { Message, StoredUser, User, BulletinPost, hydrateUser } from "../types";
import { UserInfo } from "./gpg";
import { GPGService } from "./gpg";

const API_BASE_URL = "http://localhost:8080/v1";

export class APIService {
  private static instance: APIService;
  private gpg: GPGService;

  private constructor() {
    axios.defaults.baseURL = API_BASE_URL;
    this.gpg = GPGService.getInstance();
  }

  static getInstance(): APIService {
    if (!APIService.instance) {
      APIService.instance = new APIService();
    }
    return APIService.instance;
  }

  async getMessages(): Promise<Message[]> {
    const response = await axios.get("/messages");
    return response.data || [];
  }

  async getUsers(): Promise<User[]> {
    return axios
      .get("/users")
      .then((res) => res.data)
      .then((users: StoredUser[]) => Promise.all(users.map(hydrateUser)));
  }

  async getUser(fingerprint: string): Promise<User> {
    return axios
      .get(`/users/${fingerprint}`)
      .then((res) => res.data)
      .then(hydrateUser);
  }

  async registerUser(userInfo: UserInfo): Promise<void> {
    await axios.post("/users", {
      fingerprint: userInfo.fingerprint,
      public_key: userInfo.publicKey,
      name: userInfo.name,
      email: userInfo.email,
    });
  }

  async sendPrivateMessage(
    recipientFingerprint: string,
    content: string,
  ): Promise<void> {
    const currentFingerprint = this.gpg.getCurrentFingerprint();
    if (!currentFingerprint) {
      throw new Error("No key loaded");
    }

    await axios.post("/messages", {
      recipient: recipientFingerprint,
      content,
      fingerprint: currentFingerprint,
      type: "private",
    });
  }

  async sendBulletinPost(
    topic: string,
    content: string,
    parentId?: string,
  ): Promise<void> {
    const currentFingerprint = this.gpg.getCurrentFingerprint();
    if (!currentFingerprint) {
      throw new Error("No key loaded");
    }
    const armoredContent = await this.gpg.clearSignMessage(content);

    await axios.post("/bulletin", {
      topic,
      content: armoredContent,
      parent_id: parentId,
    });
  }

  async getTopics(): Promise<string[]> {
    const response = await axios.get("/topics");
    return response.data || [];
  }

  async getConversations(): Promise<
    {
      fingerprint: string;
      messages: Message[];
    }[]
  > {
    const messages = await this.getMessages();
    const users = await this.getUsers();
    const currentFingerprint = this.gpg.getCurrentFingerprint();

    if (!currentFingerprint) {
      throw new Error("No key loaded");
    }

    // Group messages by sender/recipient
    const conversations = new Map<string, Message[]>();

    messages.forEach((msg) => {
      if (msg.type === "private") {
        // If we're the sender, group by recipient
        // If we're the recipient, group by sender
        const counterpart =
          msg.sender === currentFingerprint ? msg.recipient : msg.sender;
        const key = counterpart;
        if (!key) return;
        if (!conversations.has(key)) {
          conversations.set(key, []);
        }
        conversations.get(key)?.push(msg);
      }
    });

    // Sort messages in each conversation by timestamp
    return Array.from(conversations.entries()).map(
      ([fingerprint, messages]) => ({
        fingerprint,
        messages: messages.sort(
          (a, b) =>
            new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
        ),
      }),
    );
  }

  async getBulletinPosts(): Promise<BulletinPost[]> {
    const response = await axios.get("/bulletin");
    const posts: BulletinPost[] = response.data || [];
    return posts.sort(
      (a, b) =>
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
    );
  }

  // Search users via backend, falling back to client-side filtering
  async searchUsers(q: string, limit = 20, offset = 0): Promise<StoredUser[]> {
    try {
      const response = await axios.get("/users/search", {
        params: { q, limit, offset },
      });
      const data = response.data as { users?: StoredUser[] } | StoredUser[];
      const users = Array.isArray(data) ? data : data.users || [];
      return users.slice(0, limit);
    } catch (err) {
      // Fallback: filter locally (non-scalable, but acceptable for dev)
      const all = await this.getUsers();
      const needle = q.toLowerCase();
      const filtered = all.filter((u) =>
        (u.fingerprint || "").toLowerCase().includes(needle),
      );
      return filtered.slice(offset, offset + limit);
    }
  }

  // Recent users via backend, fallback to conversation counterparts
  async getRecentUsers(limit = 10): Promise<StoredUser[]> {
    try {
      const currentFingerprint = this.gpg.getCurrentFingerprint();
      const response = await axios.get("/users/recent", {
        params: { limit },
        headers: currentFingerprint
          ? { "X-User-Fingerprint": currentFingerprint }
          : undefined,
      });
      const users: StoredUser[] = response.data || [];
      return users.slice(0, limit);
    } catch (err) {
      const convs = await this.getConversations();
      const byRecent = convs
        .sort(
          (a, b) =>
            new Date(b.messages[0].created_at).getTime() -
            new Date(a.messages[0].created_at).getTime(),
        )
        .slice(0, limit)
        .map((c) => c.fingerprint);
      const users = await this.getUsers();
      const map = new Map(users.map((u) => [u.fingerprint, u]));
      return byRecent.map((fp) => map.get(fp)).filter(Boolean) as StoredUser[];
    }
  }
}
