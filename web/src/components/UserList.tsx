import { Card, Group, Stack, Text } from "@mantine/core";
import { useEffect } from "react";
import { GPGService } from "../services/gpg";
import { useAppStore } from "../store";
import UserAvatar from "./avatar/UserAvatar";

export function UserList() {
  const users = useAppStore((s) => s.users.items);
  const status = useAppStore((s) => s.users.status);

  if (status === "loading" && users.length === 0) {
    return <Text>Loading users...</Text>;
  }

  return (
    <Stack>
      <Stack gap="md">
        {users.map((user) => (
          <Card key={user.fingerprint} shadow="sm" p="md">
            <Group>
              <UserAvatar seed={user.fingerprint} size={60} />
              <div>
                {user.name && <Text>Name: {user.name}</Text>}
                {user.email && <Text>Email: {user.email}</Text>}
              </div>
            </Group>
          </Card>
        ))}
      </Stack>
    </Stack>
  );
}
