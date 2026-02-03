import { GPGService } from "../services/gpg";

export interface BulletinPost {
  id: string;
  parent_id?: string;
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
  /** Optional message type from backend (e.g., "private") */
  type?: string;
  /** Legacy field in some responses; prefer `sender` */
  author?: string;
}

export interface User {
  id: string;
  fingerprint: string;
  name: string;
  email: string;
  public_key: string;
}

// StoredUser mirrors server data shape for lightweight user entries
export interface StoredUser {
  fingerprint: string;
  public_key: string;
  // Optional identity details when available
  name?: string;
  email?: string;
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

// Hydrate user from public key using gpg.extractUserInfo
export async function hydrateUser(user: User | StoredUser): Promise<User> {
  return GPGService.getInstance()
    .extractUserInfo(user.public_key)
    .then(
      (extraUserInfo) =>
        ({
          ...user,
          ...extraUserInfo,
        }) as User,
    );
}
