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
} from "@mantine/core";
import { GPGService } from "./services/gpg";
import { APIService } from "./services/api";
import { Message, Topic } from "./types";
import { KeyGeneration } from "./components/KeyGeneration";

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
    } catch (error) {
      console.error("Failed to load data:", error);
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
      } catch (error) {
        console.error("Failed to decrypt message:", error);
        decrypted[msg.id] = msg.body;
      }
    }
    setDecryptedMessages(decrypted);
  };

  const handleKeyImport = async () => {
    try {
      await gpgService.importPrivateKey(privateKey);
      setIsKeyLoaded(true);
    } catch (error) {
      console.error("Failed to import key:", error);
    }
  };

  return (
    <MantineProvider>
      <Container size="xl">
        <Box pt="md" pb="md">
          <Group mb="md">
            <Text fz="xl" fw={700}>
              Axial BBS
            </Text>
            {isKeyLoaded && (
              <Text fz="sm" c="dimmed">
                Fingerprint: {gpgService.getCurrentFingerprint()}
              </Text>
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
                        <Textarea
                          placeholder="Paste your private key here"
                          value={privateKey}
                          onChange={(e) => setPrivateKey(e.currentTarget.value)}
                          minRows={3}
                          autosize
                        />
                        <Button onClick={handleKeyImport}>Import Key</Button>
                      </Stack>
                    </Tabs.Panel>

                    <Tabs.Panel value="generate" pt="md">
                      <KeyGeneration
                        onKeyGenerated={(key) => {
                          setPrivateKey(key);
                          handleKeyImport();
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
              {selectedTopic && (
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
              )}
            </Box>
          </Group>
        </Box>
      </Container>
    </MantineProvider>
  );
}

export default App;
