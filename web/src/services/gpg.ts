import * as openpgp from "openpgp";
import { KeyPair } from "../types";

export class GPGService {
  private static instance: GPGService;
  private currentKeyPair: KeyPair | null = null;

  private constructor() {}

  static getInstance(): GPGService {
    if (!GPGService.instance) {
      GPGService.instance = new GPGService();
    }
    return GPGService.instance;
  }

  async importPrivateKey(privateKeyArmored: string): Promise<KeyPair> {
    try {
      const privateKey = await openpgp.readPrivateKey({
        armoredKey: privateKeyArmored,
      });
      const publicKey = privateKey.toPublic();
      const fingerprint = publicKey.getFingerprint();

      this.currentKeyPair = {
        privateKey: privateKeyArmored,
        publicKey: publicKey.armor(),
        fingerprint,
      };

      return this.currentKeyPair;
    } catch (error) {
      throw new Error("Invalid private key");
    }
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
