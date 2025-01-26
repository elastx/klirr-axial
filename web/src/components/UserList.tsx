import { useEffect, useState } from "react";
import { Stack, Text, Button, Group, Card, Avatar } from "@mantine/core";
import { APIService } from "../services/api";
import { GPGService } from "../services/gpg";
import { User } from "../types";
import UserAvatar from "./avatar/UserAvatar";

export function UserList() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [isRegistered, setIsRegistered] = useState(false);

  const api = APIService.getInstance();
  const gpg = GPGService.getInstance();

  const loadUsers = async () => {
    try {
      const fetchedUsers = await api.getUsers();

      const currentUser = await gpg.getCurrentUserInfo();
      if (currentUser) {
        setIsRegistered(
          fetchedUsers.some((u) => u.fingerprint === currentUser.fingerprint)
        );
      }

      // Convert fetchedUsers to User[] by using gpg.getUserInfo
      const enrichedUsers = await Promise.all(
        fetchedUsers.map(async (user) => {
          const userInfo = await gpg.getUserInfo(user.public_key);
          if (!userInfo) {
            return null;
          }
          return {
            id: Math.random(), // Generate random ID since it's not in the stored user
            fingerprint: user.fingerprint,
            public_key: user.public_key,
            name: userInfo.name,
            email: userInfo.email,
          };
        })
      );

      const validUsers = enrichedUsers.filter(
        (user): user is User => user !== null
      );
      setUsers(validUsers);
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
