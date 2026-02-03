import axios from "axios";
import { Message, StoredUser, User } from "../types";
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

  async getUsers(): Promise<StoredUser[]> {
    const response = await axios.get("/users");
    return response.data || [];
  }

  async getUser(fingerprint: string): Promise<User> {
    const response = await axios.get(`/users/${fingerprint}`);
    return response.data;
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
    content: string
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
    signature: string,
    parentId?: number
  ): Promise<void> {
    const currentFingerprint = this.gpg.getCurrentFingerprint();
    if (!currentFingerprint) {
      throw new Error("No key loaded");
    }

    await axios.post("/messages", {
      topic,
      content,
      author: currentFingerprint,
      signature,
      type: "bulletin",
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
        const key =
          msg.author === currentFingerprint ? msg.recipient! : msg.author;
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
            new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
        ),
      })
    );
  }

  async getBulletinPosts(): Promise<Message[]> {
    const messages = await this.getMessages();
    return messages
      .filter((msg) => msg.type === "bulletin")
      .sort(
        (a, b) =>
          new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      );
  }
}
