import {
  Text,
  Paper,
  Stack,
  Button,
  Group,
  Avatar,
  ScrollArea,
  Textarea,
} from "@mantine/core";
import { IconSend, IconMessage } from "@tabler/icons-react";
import { useState } from "react";
import { User } from "../types";

interface Conversation {
  user: User;
  messages: {
    id: string;
    content: string;
    timestamp: string;
    isOutgoing: boolean;
  }[];
}

interface NewMessageProps {
  recipient?: User;
  onClose: () => void;
}

function NewMessage({ recipient, onClose }: NewMessageProps) {
  const [message, setMessage] = useState("");

  const handleSend = async () => {
    // TODO: Implement sending message
    onClose();
  };

  return (
    <Paper p="md" withBorder>
      <Stack>
        <Group justify="space-between">
          <Text size="xl" fw={500}>
            New Message
          </Text>
          <Text c="dimmed">To: {recipient?.name || "Select recipient"}</Text>
        </Group>
        <Textarea
          placeholder="Type your message here"
          minRows={6}
          value={message}
          onChange={(e) => setMessage(e.currentTarget.value)}
        />
        <Group justify="flex-end">
          <Button variant="subtle" onClick={onClose}>
            Cancel
          </Button>
          <Button
            disabled={!message || !recipient}
            onClick={handleSend}
            leftSection={<IconSend size={16} />}
          >
            Send
          </Button>
        </Group>
      </Stack>
    </Paper>
  );
}

interface ConversationViewProps {
  conversation: Conversation;
}

function ConversationView({ conversation }: ConversationViewProps) {
  const [reply, setReply] = useState("");

  const handleSend = async () => {
    // TODO: Implement sending reply
    setReply("");
  };

  return (
    <Paper p="md" withBorder>
      <Stack>
        <Group>
          <Avatar color="blue" radius="xl">
            {conversation.user.name?.[0] || "?"}
          </Avatar>
          <div>
            <Text fw={500}>{conversation.user.name || "Unknown"}</Text>
            <Text size="sm" c="dimmed">
              {conversation.user.email}
            </Text>
          </div>
        </Group>

        <ScrollArea h={400}>
          <Stack gap="xs">
            {conversation.messages.map((msg) => (
              <Paper
                key={msg.id}
                p="sm"
                withBorder
                bg={msg.isOutgoing ? "var(--mantine-color-dark-6)" : undefined}
                style={{
                  marginLeft: msg.isOutgoing ? "auto" : 0,
                  marginRight: msg.isOutgoing ? 0 : "auto",
                  maxWidth: "80%",
                }}
              >
                <Text size="sm">{msg.content}</Text>
                <Text
                  size="xs"
                  c="dimmed"
                  ta={msg.isOutgoing ? "right" : "left"}
                >
                  {new Date(msg.timestamp).toLocaleTimeString()}
                </Text>
              </Paper>
            ))}
          </Stack>
        </ScrollArea>

        <Group align="flex-end">
          <Textarea
            placeholder="Type your reply..."
            style={{ flex: 1 }}
            value={reply}
            onChange={(e) => setReply(e.currentTarget.value)}
            minRows={2}
          />
          <Button
            disabled={!reply}
            onClick={handleSend}
            leftSection={<IconSend size={16} />}
          >
            Send
          </Button>
        </Group>
      </Stack>
    </Paper>
  );
}

interface MessagesProps {
  initialRecipient?: User;
  onClose?: () => void;
}

export function Messages({ initialRecipient, onClose }: MessagesProps) {
  const [showNewMessage, setShowNewMessage] = useState(!!initialRecipient);
  const [selectedConversation, setSelectedConversation] =
    useState<Conversation | null>(null);

  // TODO: Fetch conversations from API
  const conversations: Conversation[] = [];

  if (showNewMessage) {
    return (
      <NewMessage
        recipient={initialRecipient}
        onClose={() => {
          setShowNewMessage(false);
          onClose?.();
        }}
      />
    );
  }

  if (selectedConversation) {
    return <ConversationView conversation={selectedConversation} />;
  }

  return (
    <Stack>
      <Group justify="space-between">
        <Text size="xl" fw={500}>
          Messages
        </Text>
        <Button
          leftSection={<IconMessage size={16} />}
          onClick={() => setShowNewMessage(true)}
        >
          New Message
        </Button>
      </Group>

      {conversations.length === 0 ? (
        <Paper p="md" withBorder>
          <Stack align="center" py="xl">
            <Text c="dimmed">No conversations yet</Text>
            <Button
              variant="light"
              leftSection={<IconMessage size={16} />}
              onClick={() => setShowNewMessage(true)}
            >
              Start a Conversation
            </Button>
          </Stack>
        </Paper>
      ) : (
        <Stack>
          {conversations.map((conv) => (
            <Paper
              key={conv.user.fingerprint}
              p="md"
              withBorder
              onClick={() => setSelectedConversation(conv)}
              style={{ cursor: "pointer" }}
            >
              <Group>
                <Avatar color="blue" radius="xl">
                  {conv.user.name?.[0] || "?"}
                </Avatar>
                <div style={{ flex: 1 }}>
                  <Group justify="space-between">
                    <Text fw={500}>{conv.user.name || "Unknown"}</Text>
                    <Text size="sm" c="dimmed">
                      {new Date(
                        conv.messages[0].timestamp
                      ).toLocaleDateString()}
                    </Text>
                  </Group>
                  <Text size="sm" lineClamp={1}>
                    {conv.messages[0].content}
                  </Text>
                </div>
              </Group>
            </Paper>
          ))}
        </Stack>
      )}
    </Stack>
  );
}
