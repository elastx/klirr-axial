import { useState } from "react";
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
} from "@mantine/core";
import { GPGService } from "../services/gpg";

export function KeyGeneration() {
  const [privateKey, setPrivateKey] = useState("");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
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

  const handleGenerateKey = async () => {
    try {
      await gpg.generateKey(name, email);
      window.location.reload();
    } catch (error) {
      console.error("Failed to generate key:", error);
    }
  };

  return (
    <Container size="sm" py="xl">
      <Paper p="md" withBorder>
        <Tabs defaultValue="import">
          <Tabs.List>
            <Tabs.Tab value="import">Import Key</Tabs.Tab>
            <Tabs.Tab value="generate">Generate Key</Tabs.Tab>
          </Tabs.List>

          <Tabs.Panel value="import" pt="md">
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

          <Tabs.Panel value="generate" pt="md">
            <Stack>
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
              <Button onClick={handleGenerateKey}>Generate Key</Button>
            </Stack>
          </Tabs.Panel>
        </Tabs>
      </Paper>
    </Container>
  );
}
