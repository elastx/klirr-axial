import { useEffect, useState } from "react";
import { Stack, Text, Button, Group, Card } from "@mantine/core";
import { IconMessage } from "@tabler/icons-react";
import { APIService } from "../services/api";
import { GPGService } from "../services/gpg";
import { User } from "../types";
import { Messages } from "./Messages";

export function UserList() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [isRegistered, setIsRegistered] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);

  const api = APIService.getInstance();
  const gpg = GPGService.getInstance();

  const loadUsers = async () => {
    try {
      const fetchedUsers = await api.getUsers();
      setUsers(fetchedUsers);

      // Check if current user is registered
      const currentUser = await gpg.getCurrentUserInfo();
      if (currentUser) {
        setIsRegistered(
          fetchedUsers.some((u) => u.fingerprint === currentUser.fingerprint)
        );
      }
    } catch (error) {
      console.error("Failed to load users:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleRegister = async () => {
    try {
      const userInfo = await gpg.getCurrentUserInfo();
      if (!userInfo) {
        throw new Error("No key loaded");
      }

      await api.registerUser(userInfo);
      await loadUsers(); // Refresh the list
    } catch (error) {
      console.error("Failed to register:", error);
    }
  };

  useEffect(() => {
    loadUsers();
  }, []);

  if (loading) {
    return <Text>Loading users...</Text>;
  }

  if (selectedUser) {
    return (
      <Messages
        initialRecipient={selectedUser}
        onClose={() => setSelectedUser(null)}
      />
    );
  }

  return (
    <Stack>
      {!isRegistered && (
        <Group justify="center">
          <Button onClick={handleRegister}>Register</Button>
        </Group>
      )}

      <Stack gap="md">
        {users.map((user) => (
          <Card key={user.id} shadow="sm" p="md">
            <Group justify="space-between">
              <div>
                <Text size="sm" c="dimmed">
                  Fingerprint: {user.fingerprint}
                </Text>
                {user.name && <Text>Name: {user.name}</Text>}
                {user.email && <Text>Email: {user.email}</Text>}
              </div>
              <Button
                variant="light"
                leftSection={<IconMessage size={16} />}
                onClick={() => setSelectedUser(user)}
              >
                Message
              </Button>
            </Group>
          </Card>
        ))}
      </Stack>
    </Stack>
  );
}
