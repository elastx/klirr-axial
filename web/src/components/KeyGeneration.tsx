import { useState } from "react";
import { TextInput, Button, Stack, Text, Paper, Group } from "@mantine/core";
import * as openpgp from "openpgp";

interface KeyGenerationProps {
  onKeyGenerated: (privateKey: string) => void;
}

export function KeyGeneration({ onKeyGenerated }: KeyGenerationProps) {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [generating, setGenerating] = useState(false);
  const [generatedKey, setGeneratedKey] = useState<{
    privateKey: string;
    publicKey: string;
    fingerprint: string;
  } | null>(null);

  const generateKey = async () => {
    try {
      setGenerating(true);
      const { privateKey: rawPrivateKey, publicKey: rawPublicKey } =
        await openpgp.generateKey({
          userIDs: [{ name, email }],
          format: "armored",
          type: "rsa",
          rsaBits: 4096,
        });

      // Normalize the key format
      const normalizeKey = (key: string) => {
        return key
          .replace(/\r\n/g, "\n") // Convert Windows line endings
          .replace(/\n\n+/g, "\n\n") // Normalize multiple blank lines
          .trim(); // Remove leading/trailing whitespace
      };

      const privateKey = normalizeKey(rawPrivateKey);
      const publicKey = normalizeKey(rawPublicKey);

      // Verify the private key can be read
      const privKeyObj = await openpgp.readPrivateKey({
        armoredKey: privateKey,
      });
      const fingerprint = privKeyObj.getFingerprint();

      console.log("Generated private key:", privateKey);

      const keyPair = {
        privateKey,
        publicKey,
        fingerprint,
      };
      setGeneratedKey(keyPair);
    } catch (error) {
      console.error("Failed to generate key:", error);
    } finally {
      setGenerating(false);
    }
  };

  const handleUseKey = async () => {
    if (generatedKey) {
      try {
        // Verify the key one more time before using it
        await openpgp.readPrivateKey({
          armoredKey: generatedKey.privateKey,
        });
        console.log("Using key:", generatedKey.privateKey);
        onKeyGenerated(generatedKey.privateKey);
      } catch (error) {
        console.error("Failed to verify key before use:", error);
      }
    }
  };

  return (
    <Stack>
      {!generatedKey ? (
        <>
          <Text>Generate a new GPG key:</Text>
          <TextInput
            label="Name"
            placeholder="Your name"
            value={name}
            onChange={(e) => setName(e.currentTarget.value)}
          />
          <TextInput
            label="Email"
            placeholder="your.email@example.com"
            value={email}
            onChange={(e) => setEmail(e.currentTarget.value)}
          />
          <Button
            onClick={generateKey}
            loading={generating}
            disabled={!name || !email}
          >
            Generate Key
          </Button>
        </>
      ) : (
        <Stack>
          <Paper p="md" withBorder>
            <Stack>
              <Text fw={500}>Your new key has been generated!</Text>
              <Text size="sm">Fingerprint: {generatedKey.fingerprint}</Text>
              <Text size="sm" c="dimmed">
                Make sure to save both your public and private keys somewhere
                safe. Your private key will be needed to decrypt messages sent
                to you.
              </Text>
              <Group>
                <Button onClick={handleUseKey}>Use This Key</Button>
                <Button
                  variant="light"
                  onClick={() => {
                    const blob = new Blob([generatedKey.privateKey], {
                      type: "text/plain",
                    });
                    const url = URL.createObjectURL(blob);
                    const a = document.createElement("a");
                    a.href = url;
                    a.download = "private.key";
                    a.click();
                    URL.revokeObjectURL(url);
                  }}
                >
                  Download Private Key
                </Button>
                <Button
                  variant="light"
                  onClick={() => {
                    const blob = new Blob([generatedKey.publicKey], {
                      type: "text/plain",
                    });
                    const url = URL.createObjectURL(blob);
                    const a = document.createElement("a");
                    a.href = url;
                    a.download = "public.key";
                    a.click();
                    URL.revokeObjectURL(url);
                  }}
                >
                  Download Public Key
                </Button>
              </Group>
            </Stack>
          </Paper>
        </Stack>
      )}
    </Stack>
  );
}
