import { useState, useCallback } from "react";
import {
  Paper,
  Text,
  Button,
  Textarea,
  Stack,
  Group,
  Tabs,
  FileButton,
  Container,
  TextInput,
  SimpleGrid,
  Center,
  Box,
  LoadingOverlay,
} from "@mantine/core";
import { IconRefresh } from "@tabler/icons-react";
import { GPGService } from "../services/gpg";
import UserAvatar from "./avatar/UserAvatar";

type GeneratedKey = {
  privateKey: string;
  publicKey: string;
  fingerprint: string;
  name: string;
  email: string;
};

// Calculate width based on 5 avatars of 100px each, plus gaps and padding
const AVATAR_SIZE = 100;
const AVATAR_COLUMNS = 5;
const GAP_SIZE = 16; // md spacing is typically 16px
const PAPER_PADDING = 32; // xl padding is typically 32px
const PAPER_WIDTH =
  AVATAR_SIZE * AVATAR_COLUMNS +
  (AVATAR_COLUMNS - 1) * GAP_SIZE +
  PAPER_PADDING * 2;

export function KeyGeneration() {
  const [privateKey, setPrivateKey] = useState("");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [generatedKeys, setGeneratedKeys] = useState<GeneratedKey[]>([]);
  const [isGenerating, setIsGenerating] = useState(false);
  const gpg = GPGService.getInstance();

  const handleFileUpload = (file: File | null) => {
    if (!file) return;

    file
      .text()
      .then((text) => {
        setPrivateKey(text);
        handleKeyImport(text);
      })
      .catch((error) => {
        console.error("Failed to read key file:", error);
      });
  };

  const handleKeyImport = async (keyText?: string) => {
    try {
      await gpg.importPrivateKey(keyText || privateKey);
      window.location.reload();
    } catch (error) {
      console.error("Failed to import key:", error);
    }
  };

  const generateKeys = useCallback(async () => {
    if (!name || !email) return;

    setIsGenerating(true);
    try {
      const keys: GeneratedKey[] = [];
      for (let i = 0; i < 50; i++) {
        const key = await gpg.generateKeyWithoutSaving(name, email);
        keys.push(key);
      }
      setGeneratedKeys(keys);
    } catch (error) {
      console.error("Failed to generate keys:", error);
    } finally {
      setIsGenerating(false);
    }
  }, [name, email]);

  const handleSelectKey = async (key: GeneratedKey) => {
    try {
      await gpg.importPrivateKey(key.privateKey);
      window.location.reload();
    } catch (error) {
      console.error("Failed to import selected key:", error);
    }
  };

  return (
    <Box
      style={{
        width: "100vw",
        minHeight: "100vh",
        display: "flex",
        alignItems: "flex-start",
        justifyContent: "center",
        padding: "3rem 0",
      }}
    >
      <Paper p="xl" withBorder style={{ width: PAPER_WIDTH }}>
        <Tabs defaultValue="import">
          <Tabs.List>
            <Tabs.Tab value="import">Import Key</Tabs.Tab>
            <Tabs.Tab value="generate">Generate Key</Tabs.Tab>
          </Tabs.List>

          <Tabs.Panel value="import" pt="xl">
            <Stack>
              <Text>Import your GPG private key:</Text>
              <Group>
                <FileButton
                  onChange={handleFileUpload}
                  accept=".key,.asc,.gpg,text/plain"
                >
                  {(props) => (
                    <Button variant="light" {...props}>
                      Upload Key File
                    </Button>
                  )}
                </FileButton>
                <Text size="sm" c="dimmed">
                  or paste below
                </Text>
              </Group>
              <Textarea
                placeholder="Paste your private key here"
                value={privateKey}
                onChange={(e) => setPrivateKey(e.currentTarget.value)}
                minRows={3}
                autosize
              />
              <Button onClick={() => handleKeyImport()}>Import Key</Button>
            </Stack>
          </Tabs.Panel>

          <Tabs.Panel value="generate" pt="xl">
            <Stack>
              <Group grow>
                <TextInput
                  label="Name"
                  placeholder="Your name"
                  value={name}
                  onChange={(e) => setName(e.currentTarget.value)}
                />
                <TextInput
                  label="Email"
                  placeholder="your@email.com"
                  value={email}
                  onChange={(e) => setEmail(e.currentTarget.value)}
                />
              </Group>

              <Group justify="center">
                <Button
                  onClick={generateKeys}
                  leftSection={<IconRefresh size={16} />}
                  disabled={!name || !email}
                >
                  Generate 50 Keys
                </Button>
              </Group>

              <Box pos="relative">
                <LoadingOverlay visible={isGenerating} />
                {generatedKeys.length > 0 && (
                  <SimpleGrid cols={5} spacing="md">
                    {generatedKeys.map((key) => (
                      <Center key={key.fingerprint}>
                        <Box
                          style={{ cursor: "pointer" }}
                          onClick={() => handleSelectKey(key)}
                        >
                          <UserAvatar seed={key.fingerprint} size={100} />
                          <Text size="xs" ta="center" mt="xs" c="dimmed">
                            {key.fingerprint.slice(0, 8)}
                          </Text>
                        </Box>
                      </Center>
                    ))}
                  </SimpleGrid>
                )}
              </Box>
            </Stack>
          </Tabs.Panel>
        </Tabs>
      </Paper>
    </Box>
  );
}
