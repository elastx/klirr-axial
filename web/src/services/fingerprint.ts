import * as openpgp from "openpgp";

// Compute frontend fingerprint in the same way as GPGService
export async function getFrontendFingerprintFromArmored(armoredKey: string): Promise<string> {
  const publicKey = await openpgp.readKey({ armoredKey });
  try {
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
  return publicKey.getKeyID().toHex().toLowerCase();
}
