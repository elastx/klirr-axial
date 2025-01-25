export interface Message {
  id: string;
  created_at: string;
  body: string;
  recipient: string;
  sender: string;
  signature: string;
}

export interface User {
  fingerprint: string;
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
