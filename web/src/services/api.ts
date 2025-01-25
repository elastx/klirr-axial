import axios from "axios";
import { Message, User } from "../types";
import { UserInfo } from "./gpg";

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
    const response = await axios.get("/messages");
    return response.data || [];
  }

  async getUsers(): Promise<User[]> {
    const response = await axios.get("/users");
    return response.data || [];
  }

  async registerUser(userInfo: UserInfo): Promise<void> {
    await axios.post("/users", {
      fingerprint: userInfo.fingerprint,
      public_key: userInfo.publicKey,
      name: userInfo.name,
      email: userInfo.email,
    });
  }

  async sendMessage(
    message: Omit<Message, "id" | "created_at">
  ): Promise<void> {
    await axios.post("/messages", message);
  }

  async getTopics(): Promise<string[]> {
    const response = await axios.get("/topics");
    return response.data || [];
  }
}
