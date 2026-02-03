export interface BulletinPost {
  id: string;
  parent_id: number;
  topic: string;
  sender: string;
  content: string;
  created_at: string;
}

export interface Message {
  id: string;
  sender: string;
  recipient: string;
  content: string;
  created_at: string;
}

export interface User {
  id: string;
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
