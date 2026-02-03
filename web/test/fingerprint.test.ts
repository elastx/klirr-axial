import { describe, it, expect } from "vitest";
import * as openpgp from "openpgp";
import { getFrontendFingerprintFromArmored } from "../src/services/fingerprint";
import { execFileSync } from "node:child_process";
import path from "node:path";

describe("backend vs frontend fingerprint parity", () => {
  it("generates same fingerprint for the same public key", async () => {
    const { publicKey: rawPublicKey } = await openpgp.generateKey({
      type: "ecc",
      curve: "curve25519",
      userIDs: [{ name: "Test User", email: "test@example.com" }],
      format: "armored",
    });

    // Frontend-computed fingerprint
    const frontFp = await getFrontendFingerprintFromArmored(rawPublicKey);

    // Backend-computed fingerprint via Go CLI (reads from stdin)
    const repoRoot = path.resolve(process.cwd(), "../src");
    const backFp = execFileSync(
      "go",
      ["run", "./cmd/fingerprint"],
      { input: rawPublicKey, encoding: "utf8", cwd: repoRoot }
    ).trim();

    expect(frontFp).toEqual(backFp);
  });
});
