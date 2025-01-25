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
        });

      // Ensure proper armor formatting with exactly one blank line after headers
      const formatArmoredKey = (key: string) => {
        // Split into lines
        const lines = key.split("\n");
        let headerEnd = -1;

        // Find the end of headers
        for (let i = 0; i < lines.length; i++) {
          if (
            lines[i].startsWith("Version:") ||
            lines[i].startsWith("Comment:") ||
            lines[i].startsWith("MessageID:")
          ) {
            headerEnd = i;
          }
        }

        if (headerEnd !== -1) {
          // Ensure exactly one blank line after headers
          const beforeHeaders = lines.slice(0, headerEnd + 1);
          const afterHeaders = lines
            .slice(headerEnd + 1)
            .filter((line) => line !== "");
          return [...beforeHeaders, "", ...afterHeaders].join("\n");
        }

        return key;
      };

      const privateKey = formatArmoredKey(rawPrivateKey);
      const publicKey = formatArmoredKey(rawPublicKey);

      // Verify the private key can be read
      const privKeyObj = await openpgp.readPrivateKey({
        armoredKey: privateKey,
      });
      const fingerprint = privKeyObj.getFingerprint();

      setGeneratedKey({
        privateKey,
        publicKey,
        fingerprint,
      });
    } catch (error) {
      console.error("Failed to generate key:", error);
    } finally {
      setGenerating(false);
    }
  };

  const handleUseKey = () => {
    if (generatedKey) {
      onKeyGenerated(generatedKey.privateKey);
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
