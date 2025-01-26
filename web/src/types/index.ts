export interface Message {
  id: number;
  topic?: string;
  recipient?: string;
  content: string;
  fingerprint: string;
  created_at: string;
  type: "private" | "bulletin";
  parent_id?: number;
}

export interface StoredUser {
  fingerprint: string;
  public_key: string;
}

export interface User {
  id: number;
  fingerprint: string;
  name: string;
  email: string;
  public_key: string;
}

export interface Topic {
  name: string;
  messages: Message[];
}

export interface KeyPair {
  publicKey: string;
  privateKey: string;
  fingerprint: string;
}
