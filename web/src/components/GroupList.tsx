import {
  ActionIcon,
  Box,
  Card,
  Group,
  Stack,
  Text,
  Title,
} from "@mantine/core";
import { IconPlus } from "@tabler/icons-react";
import { useState } from "react";
import { useAppStore } from "../store";
import { GPGService } from "../services/gpg";
import { KeyGeneration } from "./KeyGeneration";
import GroupAvatar from "./avatar/GroupAvatar";
import UserAvatar from "./avatar/UserAvatar";

const GroupList: React.FC = () => {
  const { items } = useAppStore().groups;

  const [showCreateForm, setShowCreateForm] = useState(false);

  // Creation is handled in the form via registration + encryption

  return (
    <>
      <Group justify="space-between">
        <Title order={2} mb="md">
          Groups
        </Title>
        <ActionIcon onClick={() => setShowCreateForm(true)}>
          <IconPlus />
        </ActionIcon>
      </Group>
      {showCreateForm && (
        <CreateGroupForm onCancel={() => setShowCreateForm(false)} />
      )}
      {items.length === 0 ? (
        <Text>No groups found.</Text>
      ) : (
        <Stack gap="sm">
          {items.map((group) => (
            <Card key={group.user.fingerprint} shadow="sm" p="md">
              <Group>
                <GroupAvatar seed={group.user.fingerprint} size={60} />
                <div>
                  {group.user.name && <Text>Name: {group.user.name}</Text>}
                  {group.user.email && <Text>Email: {group.user.email}</Text>}
                </div>
              </Group>
              <Stack>
                {group.users && group.users.length > 0
                  ? group.users.map((user) => (
                      <Card key={user.fingerprint} shadow="sm" p="md">
                        <Group>
                          <UserAvatar seed={user.fingerprint} size={60} />
                          <div>
                            {user.name && <Text>Name: {user.name}</Text>}
                            {user.email && <Text>Email: {user.email}</Text>}
                          </div>
                        </Group>
                      </Card>
                    ))
                  : group.members.map((fp) => (
                      <Card key={`${fp}-nouser`} shadow="sm" p="md">
                        <Group>
                          <UserAvatar seed={fp} size={60} />
                          <div>
                            <Text>Unknown User: {fp}</Text>
                          </div>
                        </Group>
                      </Card>
                    ))}
              </Stack>
            </Card>
          ))}
        </Stack>
      )}
    </>
  );
};

const CreateGroupForm: React.FC<{
  onCancel: () => void;
}> = ({ onCancel }) => {
  const store = useAppStore();
  const gpg = GPGService.getInstance();

  return (
    <Box>
      <KeyGeneration
        onSelectGeneratedKey={async (key) => {
          // 1) Register the selected user
          await store.users.register({
            name: key.name,
            email: key.email,
            fingerprint: key.fingerprint,
            publicKey: key.publicKey,
          });

          // 2) Encrypt the generated private key to the current active user's public key
          const currentPublicKey = gpg.getCurrentPublicKey();
          if (!currentPublicKey) {
            throw new Error("No active user key loaded");
          }
          const encryptedPrivateKey = await gpg.encryptMessage(
            key.privateKey,
            currentPublicKey,
          );

          // 3) Create the group with encrypted private key
          await store.groups.create({
            user_id: key.fingerprint,
            private_key: encryptedPrivateKey,
          });

          onCancel();
        }}
      />
    </Box>
  );
};

export default GroupList;
