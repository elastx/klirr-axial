import axios from "axios";
import { Message, User } from "../types";

const API_BASE_URL = "http://localhost:8080/v1";

export class APIService {
  private static instance: APIService;

  private constructor() {
    axios.defaults.baseURL = API_BASE_URL;
  }

  static getInstance(): APIService {
    if (!APIService.instance) {
      APIService.instance = new APIService();
    }
    return APIService.instance;
  }

  async getMessages(): Promise<Message[]> {
    const response = await axios.post("/sync", {
      nodeID: "web-client",
      ranges: [], // We'll need to implement proper sync logic
    });
    return response.data.messages || [];
  }

  async getUsers(): Promise<User[]> {
    const response = await axios.post("/sync", {
      nodeID: "web-client",
      ranges: [{ start: null, end: null }], // This range is used for users
    });
    return response.data.users || [];
  }

  async sendMessage(
    message: Omit<Message, "id" | "created_at">
  ): Promise<void> {
    await axios.put("/sync", message);
  }

  async getTopics(): Promise<string[]> {
    const messages = await this.getMessages();
    const topics = new Set<string>();

    messages.forEach((msg) => {
      // If recipient is not a fingerprint (40 hex chars), it's a topic
      if (!/^[0-9a-f]{40}$/i.test(msg.recipient)) {
        topics.add(msg.recipient);
      }
    });

    return Array.from(topics);
  }
}
