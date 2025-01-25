import { useState, useEffect } from "react";
import {
  MantineProvider,
  Container,
  Box,
  Text,
  Button,
  Textarea,
  Stack,
  Group,
  Paper,
  Tabs,
  FileButton,
} from "@mantine/core";
import { GPGService } from "./services/gpg";
import { APIService } from "./services/api";
import { Message, Topic } from "./types";
import { KeyGeneration } from "./components/KeyGeneration";
import { UserList } from "./components/UserList";

function App() {
  const [privateKey, setPrivateKey] = useState("");
  const [isKeyLoaded, setIsKeyLoaded] = useState(false);
  const [topics, setTopics] = useState<Topic[]>([]);
  const [selectedTopic, setSelectedTopic] = useState<string | null>(null);
  const [decryptedMessages, setDecryptedMessages] = useState<
    Record<string, string>
  >({});

  const gpgService = GPGService.getInstance();
  const apiService = APIService.getInstance();

  const handleFileUpload = (file: File | null) => {
    if (!file) return;

    file
      .text()
      .then((text) => {
        setPrivateKey(text);
        handleKeyImport(text);
      })
      .catch(() => {
        // Silently fail - the UI will show no changes
      });
  };

  // Check for saved key on mount
  useEffect(() => {
    if (gpgService.isKeyLoaded()) {
      setIsKeyLoaded(true);
      loadData();
    }
  }, []);

  useEffect(() => {
    if (isKeyLoaded) {
      loadData();
    }
  }, [isKeyLoaded]);

  useEffect(() => {
    if (selectedTopic) {
      decryptTopicMessages();
    }
  }, [selectedTopic]);

  const loadData = async () => {
    try {
      const allMessages = await apiService.getMessages();

      // Group messages by topic
      const topicsMap = new Map<string, Message[]>();

      allMessages.forEach((msg) => {
        const topic = /^[0-9a-f]{40}$/i.test(msg.recipient)
          ? "Direct Messages"
          : msg.recipient;
        if (!topicsMap.has(topic)) {
          topicsMap.set(topic, []);
        }
        topicsMap.get(topic)?.push(msg);
      });

      const topicsArray: Topic[] = Array.from(topicsMap.entries()).map(
        ([name, messages]) => ({
          name,
          messages,
        })
      );

      setTopics(topicsArray);
    } catch {
      // Failed to load messages - UI will show empty state
    }
  };

  const decryptTopicMessages = async () => {
    if (!selectedTopic) return;

    const topic = topics.find((t) => t.name === selectedTopic);
    if (!topic) return;

    const decrypted: Record<string, string> = {};
    for (const msg of topic.messages) {
      try {
        if (msg.recipient === gpgService.getCurrentFingerprint()) {
          decrypted[msg.id] = await gpgService.decryptMessage(msg.body);
        } else {
          decrypted[msg.id] = msg.body;
        }
      } catch {
        // If decryption fails, show encrypted message
        decrypted[msg.id] = msg.body;
      }
    }
    setDecryptedMessages(decrypted);
  };

  const handleKeyImport = async (keyText?: string) => {
    try {
      await gpgService.importPrivateKey(keyText || privateKey);
      setIsKeyLoaded(true);
      setPrivateKey("");
    } catch {
      // Key import failed - UI state won't change
    }
  };

  const handleLogout = () => {
    gpgService.clearSavedKey();
    setIsKeyLoaded(false);
    setSelectedTopic(null);
    setDecryptedMessages({});
    setTopics([]);
    setPrivateKey("");
  };

  return (
    <MantineProvider>
      <Container size="xl">
        <Box pt="md" pb="md">
          <Group mb="md" justify="space-between">
            <Text fz="xl" fw={700}>
              Axial BBS
            </Text>
            {isKeyLoaded && (
              <Group>
                <Text fz="sm" c="dimmed">
                  Fingerprint: {gpgService.getCurrentFingerprint()}
                </Text>
                <Button variant="subtle" color="red" onClick={handleLogout}>
                  Logout
                </Button>
              </Group>
            )}
          </Group>

          <Group align="flex-start" grow>
            <Box style={{ flexBasis: "300px", flexGrow: 0 }}>
              {!isKeyLoaded ? (
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
                        <Button onClick={() => handleKeyImport()}>
                          Import Key
                        </Button>
                      </Stack>
                    </Tabs.Panel>

                    <Tabs.Panel value="generate" pt="md">
                      <KeyGeneration
                        onKeyGenerated={(key) => {
                          setPrivateKey(key);
                          handleKeyImport(key);
                        }}
                      />
                    </Tabs.Panel>
                  </Tabs>
                </Paper>
              ) : (
                <Stack>
                  <Text fw={500}>Topics:</Text>
                  {topics.map((topic) => (
                    <Button
                      key={topic.name}
                      variant={
                        selectedTopic === topic.name ? "filled" : "light"
                      }
                      onClick={() => setSelectedTopic(topic.name)}
                    >
                      {topic.name}
                    </Button>
                  ))}
                </Stack>
              )}
            </Box>

            <Box>
              {selectedTopic ? (
                <Stack>
                  {topics
                    .find((t) => t.name === selectedTopic)
                    ?.messages.map((msg) => (
                      <Paper key={msg.id} p="md" withBorder>
                        <Text fz="sm" fw={500} mb="xs">
                          From: {msg.sender}
                        </Text>
                        <Text>{decryptedMessages[msg.id] || msg.body}</Text>
                      </Paper>
                    ))}
                </Stack>
              ) : (
                isKeyLoaded && <UserList />
              )}
            </Box>
          </Group>
        </Box>
      </Container>
    </MantineProvider>
  );
}

export default App;
