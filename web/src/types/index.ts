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
  // Canonical signing key ID for display/use in signed messages
  signing_fingerprint?: string;
  // All encryption subkey IDs for resolving recipients
  encryption_fingerprints?: string[];
  name: string;
  email: string;
  public_key: string;
}

export interface Group {
  id: string;
  user: User;
  // Stored raw fingerprints from backend record
  members: string[];
  // Hydrated users resolved by backend (subset of members that exist)
  users: User[];
}

// StoredUser mirrors server data shape for lightweight user entries
export interface StoredUser {
  fingerprint: string;
  public_key: string;
  // Optional identity details when available
  name?: string;
  email?: string;
  // Optional extended fingerprints when provided by backend
  signing_fingerprint?: string;
  encryption_fingerprints?: string[];
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
