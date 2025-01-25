import { useState, useEffect } from "react";
import { Stack, Text, Paper, Button, Group } from "@mantine/core";
import { APIService } from "../services/api";
import { GPGService } from "../services/gpg";
import { User } from "../types";

export function UserList() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [registered, setRegistered] = useState(false);

  const apiService = APIService.getInstance();
  const gpgService = GPGService.getInstance();

  useEffect(() => {
    loadUsers();
  }, []);

  const loadUsers = async () => {
    try {
      const userList = await apiService.getUsers();
      setUsers(userList);

      // Check if current user is registered
      const currentFingerprint = gpgService.getCurrentFingerprint();
      if (currentFingerprint) {
        setRegistered(
          userList.some((u) => u.fingerprint === currentFingerprint)
        );
      }
    } catch {
      // Failed to load users - UI will show empty state
    } finally {
      setLoading(false);
    }
  };

  const handleRegister = async () => {
    try {
      const userInfo = await gpgService.getCurrentUserInfo();
      if (!userInfo) return;

      await apiService.registerUser(userInfo);
      setRegistered(true);
      await loadUsers(); // Reload the user list
    } catch {
      // Registration failed - UI state won't change
    }
  };

  if (loading) {
    return <Text>Loading users...</Text>;
  }

  return (
    <Stack>
      <Group justify="space-between">
        <Text fw={500}>Users</Text>
        {!registered && (
          <Button variant="light" onClick={handleRegister}>
            Register
          </Button>
        )}
      </Group>
      {users.length === 0 ? (
        <Text c="dimmed">No users registered yet</Text>
      ) : (
        users.map((user) => (
          <Paper key={user.fingerprint} p="sm" withBorder>
            <Stack gap="xs">
              <Text fw={500}>{user.fingerprint}</Text>
              <Text size="sm" style={{ wordBreak: "break-all" }}>
                {user.public_key}
              </Text>
            </Stack>
          </Paper>
        ))
      )}
    </Stack>
  );
}
