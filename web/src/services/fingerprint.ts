import * as openpgp from "openpgp";

// Compute frontend fingerprint in the same way as GPGService
export async function getFrontendFingerprintFromArmored(armoredKey: string): Promise<string> {
  const publicKey = await openpgp.readKey({ armoredKey });
  // Prefer encryption subkey KeyID as canonical fingerprint
  try {
    const enc = await publicKey.getEncryptionKey();
    if (enc) {
      return enc.getKeyID().toHex().toLowerCase();
    }
  } catch {}
  // Fallback to signing key ID if no encryption key is available
  try {
    const sign = await (publicKey as any).getSigningKey?.();
    if (sign) {
      return sign.getKeyID().toHex().toLowerCase();
    }
  } catch {}
  // Final fallback to primary key ID
  return publicKey.getKeyID().toHex().toLowerCase();
}
