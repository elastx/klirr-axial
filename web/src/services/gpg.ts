import * as openpgp from "openpgp";
import { KeyPair } from "../types";

const STORAGE_KEY = "axial_gpg_key";

export interface UserInfo {
  name: string;
  email: string;
  fingerprint: string;
  publicKey: string;
}

export class GPGService {
  private static instance: GPGService;
  private currentKeyPair: KeyPair | null = null;

  private constructor() {
    // Try to load key from localStorage on initialization
    const savedKey = localStorage.getItem(STORAGE_KEY);
    if (savedKey) {
      try {
        this.currentKeyPair = JSON.parse(savedKey);
      } catch {
        localStorage.removeItem(STORAGE_KEY);
      }
    }
  }

  static getInstance(): GPGService {
    if (!GPGService.instance) {
      GPGService.instance = new GPGService();
    }
    return GPGService.instance;
  }

  async getUserInfo(publicKeyString: string): Promise<UserInfo | null> {
    const publicKey = await openpgp.readKey({
      armoredKey: publicKeyString,
    });
    const user = publicKey.users[0];
    if (!user) return null;

    return {
      name: user.userID?.name || "",
      email: user.userID?.email || "",
      fingerprint: publicKey.getFingerprint(),
      publicKey: publicKey.armor(),
    };
  }

  async getCurrentUserInfo(): Promise<UserInfo | null> {
    if (!this.currentKeyPair) return null;

    const publicKey = await openpgp.readKey({
      armoredKey: this.currentKeyPair.publicKey,
    });
    const user = publicKey.users[0];
    if (!user) return null;

    return {
      name: user.userID?.name || "",
      email: user.userID?.email || "",
      fingerprint: publicKey.getFingerprint(),
      publicKey: this.currentKeyPair.publicKey,
    };
  }

  async extractUserInfo(publicKeyArmored: string): Promise<UserInfo> {
    const publicKey = await openpgp.readKey({ armoredKey: publicKeyArmored });
    const user = publicKey.users[0];
    if (!user) {
      throw new Error("No user information found in key");
    }

    return {
      name: user.userID?.name || "",
      email: user.userID?.email || "",
      fingerprint: publicKey.getFingerprint(),
      publicKey: publicKeyArmored,
    };
  }

  async generateKey(name: string, email: string): Promise<void> {
    const { privateKey: rawPrivateKey, publicKey: rawPublicKey } =
      await openpgp.generateKey({
        type: "ecc",
        curve: "curve25519",
        userIDs: [{ name, email }],
        format: "armored",
      });

    const publicKey = await openpgp.readKey({ armoredKey: rawPublicKey });
    const fingerprint = publicKey.getFingerprint();

    this.currentKeyPair = {
      privateKey: rawPrivateKey,
      publicKey: rawPublicKey,
      fingerprint,
    };

    localStorage.setItem(STORAGE_KEY, JSON.stringify(this.currentKeyPair));
  }

  async generateKeyWithoutSaving(
    name: string,
    email: string
  ): Promise<{
    privateKey: string;
    publicKey: string;
    fingerprint: string;
    name: string;
    email: string;
  }> {
    const { privateKey: rawPrivateKey, publicKey: rawPublicKey } =
      await openpgp.generateKey({
        type: "ecc",
        curve: "curve25519",
        userIDs: [{ name, email }],
        format: "armored",
      });

    const publicKey = await openpgp.readKey({ armoredKey: rawPublicKey });
    const fingerprint = publicKey.getFingerprint();

    return {
      privateKey: rawPrivateKey,
      publicKey: rawPublicKey,
      fingerprint,
      name,
      email,
    };
  }

  async importPrivateKey(privateKeyArmored: string): Promise<KeyPair> {
    if (!privateKeyArmored || typeof privateKeyArmored !== "string") {
      throw new Error("No key provided");
    }

    // Normalize line endings and ensure proper armor format
    const normalizedKey = privateKeyArmored
      .replace(/\r\n/g, "\n") // Convert Windows line endings
      .replace(/\n\n+/g, "\n\n") // Normalize multiple blank lines
      .trim(); // Remove leading/trailing whitespace

    // Verify it's a private key
    if (!normalizedKey.includes("-----BEGIN PGP PRIVATE KEY BLOCK-----")) {
      throw new Error("Not a PGP private key");
    }

    // Try to read the private key
    const privateKey = await openpgp.readPrivateKey({
      armoredKey: normalizedKey,
    });

    if (!privateKey) {
      throw new Error("Failed to read private key");
    }

    // Get the public key and fingerprint
    const publicKey = privateKey.toPublic();
    const fingerprint = publicKey.getFingerprint();

    this.currentKeyPair = {
      privateKey: normalizedKey,
      publicKey: publicKey.armor(),
      fingerprint,
    };

    // Save to localStorage
    localStorage.setItem(STORAGE_KEY, JSON.stringify(this.currentKeyPair));

    return this.currentKeyPair;
  }

  clearSavedKey(): void {
    this.currentKeyPair = null;
    localStorage.removeItem(STORAGE_KEY);
  }

  async decryptMessage(encryptedMessage: string): Promise<string> {
    if (!this.currentKeyPair) {
      throw new Error("No private key loaded");
    }

    const privateKey = await openpgp.readPrivateKey({
      armoredKey: this.currentKeyPair.privateKey,
    });
    const message = await openpgp.readMessage({
      armoredMessage: encryptedMessage,
    });

    const decrypted = await openpgp.decrypt({
      message,
      decryptionKeys: privateKey,
    });

    return decrypted.data as string;
  }

  async encryptMessage(
    message: string,
    recipientPublicKey: string
  ): Promise<string> {
    const publicKey = await openpgp.readKey({ armoredKey: recipientPublicKey });

    const encrypted = await openpgp.encrypt({
      message: await openpgp.createMessage({ text: message }),
      encryptionKeys: publicKey,
    });

    return encrypted as string;
  }

  getCurrentFingerprint(): string | null {
    return this.currentKeyPair?.fingerprint || null;
  }

  getCurrentPublicKey(): string | null {
    return this.currentKeyPair?.publicKey || null;
  }

  isKeyLoaded(): boolean {
    return this.currentKeyPair !== null;
  }
}
