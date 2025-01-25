import {
  Text,
  Paper,
  Stack,
  Button,
  TextInput,
  Textarea,
  Group,
} from "@mantine/core";
import { IconSend } from "@tabler/icons-react";
import { useState } from "react";

export function BulletinBoard() {
  const [topic, setTopic] = useState("");
  const [message, setMessage] = useState("");

  const handleSubmit = async () => {
    // TODO: Implement sending public message
  };

  return (
    <Stack>
      <Group justify="space-between">
        <Text size="xl" fw={500}>
          Bulletin Board
        </Text>
        <Button
          leftSection={<IconSend size={16} />}
          variant="light"
          onClick={() => {
            setTopic("");
            setMessage("");
          }}
        >
          New Post
        </Button>
      </Group>

      <Paper p="md" withBorder>
        <Stack>
          <TextInput
            label="Topic"
            placeholder="Enter topic"
            value={topic}
            onChange={(e) => setTopic(e.currentTarget.value)}
          />
          <Textarea
            label="Message"
            placeholder="Type your message here"
            minRows={4}
            value={message}
            onChange={(e) => setMessage(e.currentTarget.value)}
          />
          <Button disabled={!topic || !message} onClick={handleSubmit}>
            Post Message
          </Button>
        </Stack>
      </Paper>

      <Paper p="md" withBorder>
        <Text fw={500} mb="md">
          Recent Posts
        </Text>
        <Text c="dimmed">No posts yet</Text>
      </Paper>
    </Stack>
  );
}
