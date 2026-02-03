import * as openpgp from "openpgp";
import { KeyPair, User } from "../types";
import { APIService } from "./api";

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

  private async computeFingerprintFromKey(
    publicKey: openpgp.Key,
  ): Promise<string> {
    try {
      // Prefer signing subkey to match backend behavior
      const sign = await (publicKey as any).getSigningKey?.();
      if (sign) {
        return sign.getKeyID().toHex().toLowerCase();
      }
    } catch {}
    try {
      const enc = await publicKey.getEncryptionKey();
      if (enc) {
        return enc.getKeyID().toHex().toLowerCase();
      }
    } catch {}
    // Fallback to primary key id
    return publicKey.getKeyID().toHex().toLowerCase();
  }

  private async computeFingerprintFromArmored(
    armoredKey: string,
  ): Promise<string> {
    const key = await openpgp.readKey({ armoredKey });
    return this.computeFingerprintFromKey(key);
  }

  private constructor() {
    // Try to load key from localStorage on initialization
    const savedKey = localStorage.getItem(STORAGE_KEY);
    if (savedKey) {
      try {
        this.currentKeyPair = JSON.parse(savedKey);
        // Migrate fingerprint to backend-compatible 16-hex key ID (uppercase)
        if (this.currentKeyPair?.publicKey) {
          this.computeFingerprintFromArmored(this.currentKeyPair.publicKey)
            .then((keyId) => {
              if (this.currentKeyPair) {
                this.currentKeyPair.fingerprint = keyId;
                localStorage.setItem(
                  STORAGE_KEY,
                  JSON.stringify(this.currentKeyPair),
                );
              }
            })
            .catch(() => {
              // If public key can't be read, keep existing state but clear invalid storage
            });
        }
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
      // Use encryption subkey KeyID when available to match backend
      fingerprint: await this.computeFingerprintFromKey(publicKey),
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
      fingerprint: await this.computeFingerprintFromKey(publicKey),
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
      fingerprint: await this.computeFingerprintFromKey(publicKey),
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
    const fingerprint = await this.computeFingerprintFromKey(publicKey);

    this.currentKeyPair = {
      privateKey: rawPrivateKey,
      publicKey: rawPublicKey,
      fingerprint,
    };

    localStorage.setItem(STORAGE_KEY, JSON.stringify(this.currentKeyPair));
  }

  async generateKeyWithoutSaving(
    name: string,
    email: string,
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
    const fingerprint = await this.computeFingerprintFromKey(publicKey);

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
    const fingerprint = await this.computeFingerprintFromKey(publicKey);

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
    recipientPublicKey: string,
  ): Promise<string> {
    const publicKey = await openpgp.readKey({ armoredKey: recipientPublicKey });

    const encrypted = await openpgp.encrypt({
      message: await openpgp.createMessage({ text: message }),
      encryptionKeys: publicKey,
    });

    return encrypted as string;
  }

  async signMessage(message: string): Promise<string> {
    if (!this.currentKeyPair) {
      throw new Error("No private key loaded");
    }

    const privateKey = await openpgp.readPrivateKey({
      armoredKey: this.currentKeyPair.privateKey,
    });

    const signed = await openpgp.sign({
      message: await openpgp.createMessage({ text: message }),
      signingKeys: privateKey,
      detached: true,
      format: "armored",
    });

    return signed as string;
  }

  async clearSignMessage(message: string): Promise<string> {
    if (!this.currentKeyPair) {
      throw new Error("No private key loaded");
    }

    const privateKey = await openpgp.readPrivateKey({
      armoredKey: this.currentKeyPair.privateKey,
    });

    const cleartext = await openpgp.createCleartextMessage({ text: message });
    const signed = await openpgp.sign({
      message: cleartext,
      signingKeys: privateKey,
      format: "armored",
    });

    return signed as string;
  }

  async verifyClearsignedMessage(
    armoredCleartext: string,
    signerFingerprint: string,
  ): Promise<boolean> {
    const apiService = APIService.getInstance();
    const cleartext = await openpgp.readCleartextMessage({
      cleartextMessage: armoredCleartext,
    });
    const publicKeyArmored = await apiService
      .getUser(signerFingerprint)
      .then((user: User) => user.public_key);
    const verificationKey = await openpgp.readKey({
      armoredKey: publicKeyArmored,
    });
    const result = await openpgp.verify({
      message: cleartext,
      verificationKeys: verificationKey,
    });
    return result.signatures[0].verified;
  }

  async extractClearsignedText(armoredCleartext: string): Promise<string> {
    const cleartext = await openpgp.readCleartextMessage({
      cleartextMessage: armoredCleartext,
    });
    return cleartext.getText();
  }

  async verifyMessageSignature(
    message: string,
    signature: string,
    signerFingerprint: string,
  ): Promise<boolean> {
    const apiService = APIService.getInstance();
    console.log("signature", signature);
    return Promise.all([
      openpgp.createMessage({ text: message }),
      openpgp.readSignature({ armoredSignature: signature }),
      apiService
        .getUser(signerFingerprint)
        .then((user: User) => user.public_key)
        .then((publicKey: string) =>
          openpgp.readKey({ armoredKey: publicKey }),
        ),
    ]).then(([message, signature, verificationKey]) =>
      openpgp
        .verify({
          message: message,
          signature: signature,
          verificationKeys: verificationKey,
        })
        .then((verified) => {
          return verified.signatures[0].verified;
        }),
    );
  }

  getCurrentFingerprint(): string | null {
    return this.currentKeyPair?.fingerprint || null;
  }

  getCurrentPublicKey(): string | null {
    return this.currentKeyPair?.publicKey || null;
  }

  getCurrentPrivateKey(): string | null {
    return this.currentKeyPair?.privateKey || null;
  }

  isKeyLoaded(): boolean {
    return this.currentKeyPair !== null;
  }
}
