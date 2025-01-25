export interface Message {
  id: number;
  topic: string;
  content: string;
  fingerprint: string;
  created_at: string;
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
