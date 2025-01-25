import { useEffect, useState } from "react";
import { Stack, Text, Button, Group, Card } from "@mantine/core";
import { APIService } from "../services/api";
import { GPGService } from "../services/gpg";
import { User } from "../types";
import Avatar from "./avatar/Avatar";

export function UserList() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [isRegistered, setIsRegistered] = useState(false);

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
            <Avatar seed={user.fingerprint} />
            <Text size="sm" c="dimmed">
              Fingerprint: {user.fingerprint}
            </Text>
            {user.name && <Text>Name: {user.name}</Text>}
            {user.email && <Text>Email: {user.email}</Text>}
          </Card>
        ))}
      </Stack>
    </Stack>
  );
}
