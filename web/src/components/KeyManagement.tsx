import { Button, Code, Group, Paper, Stack, Text } from "@mantine/core";
import { IconAlertTriangle, IconThumbUp } from "@tabler/icons-react";
import { useState } from "react";
import { GPGService, UserInfo } from "../services/gpg";
import UserAvatar from "./avatar/UserAvatar";

export function UserSettings() {
  const [currentUser, setCurrentUser] = useState<UserInfo | null>(null);

  const gpg = GPGService.getInstance();

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  gpg.getCurrentUserInfo().then((user) => {
    setCurrentUser(user);
  });

  const fingerprint = gpg.getCurrentFingerprint();

  if (!fingerprint) {
    return <Text>No key loaded</Text>;
  }

  return (
    <Paper p="md" withBorder>
      <Text size="xl" fw={500} mb="md">
        User Settings
      </Text>

      <Group mb="md">
        <UserAvatar seed={fingerprint} size={100} />
        <Stack style={{ flexGrow: 1 }}>
          {currentUser && (
            <Text>
              {currentUser.name} ({currentUser.email})
            </Text>
          )}
          <Stack gap={0}>
            <Text size="sm" c="dimmed" m={0} p={0}>
              Fingerprint:
            </Text>
            <Group>
              <Code block style={{ flexGrow: 1 }}>
                {gpg.getCurrentFingerprint()}
              </Code>
              <Stack>
                <Button
                  variant="light"
                  onClick={() =>
                    copyToClipboard(gpg.getCurrentFingerprint() || "")
                  }
                >
                  Copy
                </Button>
              </Stack>
            </Group>
          </Stack>
        </Stack>
      </Group>

      <Text size="sm" c="dimmed" mb={0}>
        Public key:
      </Text>
      <Group align="flex-start">
        <Code block style={{ flexGrow: 1, wordBreak: "break-all" }}>
          {gpg.getCurrentPublicKey()}
        </Code>
        <Stack>
          <Button
            variant="light"
            onClick={() => copyToClipboard(gpg.getCurrentPublicKey() || "")}
            leftSection={<IconThumbUp size={16} color="green" />}
          >
            Copy Public
          </Button>
          <Button
            variant="light"
            onClick={async () => {
              copyToClipboard(gpg.getCurrentPrivateKey() || "");
            }}
            leftSection={<IconAlertTriangle size={16} color="red" />}
          >
            Copy Private
          </Button>
        </Stack>
      </Group>
    </Paper>
  );
}
