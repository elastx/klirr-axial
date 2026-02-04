import { describe, it, expect } from "vitest";
import * as openpgp from "openpgp";
import { getFrontendFingerprintFromArmored } from "../src/services/fingerprint";
import { execFileSync } from "node:child_process";
import path from "node:path";

describe("backend vs frontend fingerprint parity", () => {
  it("uses encryption subkey KeyID (16-hex, lowercase)", async () => {
    const { publicKey: rawPublicKey } = await openpgp.generateKey({
      type: "ecc",
      curve: "curve25519",
      userIDs: [{ name: "Test User", email: "test@example.com" }],
      format: "armored",
    });

    // Compute expected encryption KeyID via OpenPGP.js
    const pub = await openpgp.readKey({ armoredKey: rawPublicKey });
    const enc = await pub.getEncryptionKey();
    const expectedEncId = enc
      ? enc.getKeyID().toHex().toLowerCase()
      : pub.getKeyID().toHex().toLowerCase();

    // Frontend-computed fingerprint should match encryption subkey KeyID
    const frontFp = await getFrontendFingerprintFromArmored(rawPublicKey);

    // Backend-computed fingerprint via Go CLI (reads from stdin)
    const repoRoot = path.resolve(process.cwd(), "../src");
    const backFp = execFileSync("go", ["run", "./cmd/fingerprint"], {
      input: rawPublicKey,
      encoding: "utf8",
      cwd: repoRoot,
    }).trim();

    // Both frontend and backend must equal the encryption KeyID
    expect(frontFp).toEqual(expectedEncId);
    expect(backFp).toEqual(expectedEncId);
  });
});
