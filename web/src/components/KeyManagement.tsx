import { Text, Paper, Code, Button, Group } from "@mantine/core";
import { GPGService } from "../services/gpg";

export function KeyManagement() {
  const gpg = GPGService.getInstance();

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  return (
    <Paper p="md" withBorder>
      <Text size="xl" fw={500} mb="md">
        Key Management
      </Text>

      <Text mb="xs">Your current key fingerprint:</Text>
      <Group mb="md">
        <Code block style={{ flexGrow: 1 }}>
          {gpg.getCurrentFingerprint()}
        </Code>
        <Button
          variant="light"
          onClick={() => copyToClipboard(gpg.getCurrentFingerprint() || "")}
        >
          Copy
        </Button>
      </Group>

      <Text size="sm" c="dimmed" mb="xs">
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
