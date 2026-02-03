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
import { useState, useEffect } from "react";
import { User, Message, StoredUser } from "../types";
import { APIService } from "../services/api";
import { UserPicker } from "./UserPicker";

// TODO: Sign or encrypt messages and store the signed or encrypted result in the contents field.
// Remove the signature field from the model and the API.

interface Conversation {
  user: User;
  messages: Message[];
}

interface NewMessageProps {
  recipient?: User;
  onClose: () => void;
  onSent?: () => void;
}

function NewMessage({ recipient, onClose, onSent }: NewMessageProps) {
  const [message, setMessage] = useState("");
  const [selectedRecipient, setSelectedRecipient] = useState<StoredUser | null>(
    recipient
      ? {
          fingerprint: recipient.fingerprint,
          public_key: recipient.public_key,
          name: recipient.name,
          email: recipient.email,
        }
      : null,
  );
  const api = APIService.getInstance();

  const handleSend = async () => {
    try {
      if (!selectedRecipient) return;
      await api.sendPrivateMessage(selectedRecipient.fingerprint, message);
      setMessage("");
      onSent?.();
      onClose();
    } catch (error) {
      console.error("Failed to send message:", error);
    }
  };

  return (
    <Paper p="md" withBorder>
      <Stack>
        <Group justify="space-between">
          <Text size="xl" fw={500}>
            New Message
          </Text>
        </Group>
        <UserPicker
          value={selectedRecipient}
          onSelect={(u) => setSelectedRecipient(u)}
        />
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
            disabled={!message || !selectedRecipient}
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
  onBack: () => void;
  onMessageSent: () => void;
}

function ConversationView({
  conversation,
  onBack,
  onMessageSent,
}: ConversationViewProps) {
  const [reply, setReply] = useState("");
  const api = APIService.getInstance();

  const handleSend = async () => {
    try {
      await api.sendPrivateMessage(conversation.user.fingerprint, reply);
      setReply("");
      onMessageSent();
    } catch (error) {
      console.error("Failed to send reply:", error);
    }
  };

  return (
    <Paper p="md" withBorder>
      <Stack>
        <Group justify="space-between">
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
          <Button variant="subtle" onClick={onBack}>
            Back to Messages
          </Button>
        </Group>

        <ScrollArea h={400}>
          <Stack gap="xs">
            {conversation.messages.map((msg) => (
              <Paper
                key={msg.id}
                p="sm"
                withBorder
                bg={
                  msg.sender === conversation.user.fingerprint
                    ? undefined
                    : "var(--mantine-color-dark-6)"
                }
                style={{
                  marginLeft:
                    msg.sender === conversation.user.fingerprint ? 0 : "auto",
                  marginRight:
                    msg.sender === conversation.user.fingerprint ? "auto" : 0,
                  maxWidth: "80%",
                }}
              >
                <Text size="sm">{msg.content}</Text>
                <Text
                  size="xs"
                  c="dimmed"
                  ta={
                    msg.sender === conversation.user.fingerprint
                      ? "left"
                      : "right"
                  }
                >
                  {new Date(msg.created_at).toLocaleTimeString()}
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
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [loading, setLoading] = useState(true);
  const api = APIService.getInstance();

  const loadConversations = async () => {
    try {
      const users = await api.getUsers();
      const convs = await api.getConversations();

      // Map conversations to users
      const fullConversations = convs
        .map((conv) => ({
          user: users.find((u) => u.fingerprint === conv.fingerprint)!,
          messages: conv.messages,
        }))
        .filter((conv) => conv.user) as Conversation[]; // Only include conversations where we found the user

      setConversations(fullConversations);
    } catch (error) {
      console.error("Failed to load conversations:", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadConversations();
  }, []);

  if (showNewMessage) {
    return (
      <NewMessage
        recipient={initialRecipient}
        onClose={() => {
          setShowNewMessage(false);
          onClose?.();
        }}
        onSent={loadConversations}
      />
    );
  }

  if (selectedConversation) {
    return (
      <ConversationView
        conversation={selectedConversation}
        onBack={() => setSelectedConversation(null)}
        onMessageSent={loadConversations}
      />
    );
  }

  if (loading) {
    return (
      <Paper p="md" withBorder>
        <Text>Loading conversations...</Text>
      </Paper>
    );
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
                        conv.messages[0].created_at
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
