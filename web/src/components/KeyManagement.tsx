import { Text, Paper, Code, Button, Group, Stack } from "@mantine/core";
import { GPGService, UserInfo } from "../services/gpg";
import UserAvatar from "./avatar/UserAvatar";
import { useState } from "react";

export function KeyManagement() {
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
        Key Management
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
              <Button
                variant="light"
                onClick={() =>
                  copyToClipboard(gpg.getCurrentFingerprint() || "")
                }
              >
                Copy
              </Button>
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
        <Button
          variant="light"
          onClick={() => copyToClipboard(gpg.getCurrentPublicKey() || "")}
        >
          Copy
        </Button>
      </Group>
    </Paper>
  );
}
