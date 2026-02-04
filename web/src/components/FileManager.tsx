import { useState, useEffect } from "react";
import {
  Container,
  Title,
  Card,
  Text,
  Button,
  Group,
  Stack,
  FileInput,
  Textarea,
  Grid,
  Badge,
  ActionIcon,
  Modal,
  Alert,
} from "@mantine/core";
import {
  IconUpload,
  IconDownload,
  IconTrash,
  IconFile,
  IconInfoCircle,
} from "@tabler/icons-react";
import { APIService } from "../services/api";
import { GPGService } from "../services/gpg";
import { useAppStore } from "../store";

interface FileMetadata {
  id: string;
  filename: string;
  size: number;
  content_type: string;
  description: string;
  uploader: string;
  content_hash: string;
  encrypted: boolean;
  recipients: string[];
  created_at: string;
}

export function FileManager() {
  const filesSlice = useAppStore((s) => s.files);
  const files = useAppStore((s) => s.files.items) as unknown as FileMetadata[];
  const loading = filesSlice.status === "loading";
  const [uploading, setUploading] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [description, setDescription] = useState("");
  const [modalOpened, setModalOpened] = useState(false);
  const [selectedMetadata, setSelectedMetadata] = useState<FileMetadata | null>(
    null
  );
  const [error, setError] = useState<string | null>(null);

  const api = APIService.getInstance();
  const gpg = GPGService.getInstance();

  useEffect(() => {
    filesSlice.refresh();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const loadFiles = async () => {
    setError(null);
    try {
      await filesSlice.refresh();
    } catch (err) {
      setError("Failed to load files");
      console.error(err);
    }
  };

  const handleUpload = async () => {
    if (!selectedFile) return;

    setUploading(true);
    setError(null);
    try {
      const fingerprint = gpg.getCurrentFingerprint();
      if (!fingerprint) {
        throw new Error("No GPG key loaded");
      }

      await filesSlice.upload(selectedFile, fingerprint, description);
      setSelectedFile(null);
      setDescription("");
      await loadFiles();
    } catch (err: any) {
      setError(err.message || "Failed to upload file");
      console.error(err);
    } finally {
      setUploading(false);
    }
  };

  const handleDownload = async (file: FileMetadata) => {
    setError(null);
    try {
      const blob = await api.downloadFile(file.id);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = file.filename;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (err) {
      setError("Failed to download file");
      console.error(err);
    }
  };

  const handleDelete = async (fileId: string) => {
    if (!confirm("Are you sure you want to delete this file?")) return;

    setError(null);
    try {
      await filesSlice.remove(fileId);
      await loadFiles();
    } catch (err) {
      setError("Failed to delete file");
      console.error(err);
    }
  };

  const showDetails = (file: FileMetadata) => {
    setSelectedMetadata(file);
    setModalOpened(true);
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + " KB";
    return (bytes / (1024 * 1024)).toFixed(2) + " MB";
  };

  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleString();
  };

  return (
    <Container size="lg">
      <Stack gap="lg">
        <Title order={2}>File Manager</Title>

        {error && (
          <Alert color="red" title="Error">
            {error}
          </Alert>
        )}

        {/* Upload Section */}
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Stack gap="md">
            <Title order={3}>Upload File</Title>
            <FileInput
              label="Select file"
              placeholder="Choose a file"
              value={selectedFile}
              onChange={setSelectedFile}
              leftSection={<IconUpload size={14} />}
            />
            <Textarea
              label="Description (optional)"
              placeholder="Enter file description"
              value={description}
              onChange={(e) => setDescription(e.currentTarget.value)}
              minRows={2}
            />
            <Button
              onClick={handleUpload}
              disabled={!selectedFile || uploading}
              loading={uploading}
              leftSection={<IconUpload size={16} />}
            >
              Upload
            </Button>
          </Stack>
        </Card>

        {/* Files List */}
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Stack gap="md">
            <Group justify="space-between">
              <Title order={3}>Files</Title>
              <Button onClick={loadFiles} loading={loading} size="sm">
                Refresh
              </Button>
            </Group>

            {loading && <Text>Loading files...</Text>}

            {!loading && files.length === 0 && (
              <Text c="dimmed">No files uploaded yet.</Text>
            )}

            <Grid>
              {files.map((file) => (
                <Grid.Col span={{ base: 12, md: 6 }} key={file.id}>
                  <Card shadow="xs" padding="md" withBorder>
                    <Stack gap="xs">
                      <Group justify="space-between">
                        <Group gap="xs">
                          <IconFile size={20} />
                          <Text fw={500} lineClamp={1}>
                            {file.filename}
                          </Text>
                        </Group>
                        {file.encrypted && (
                          <Badge color="red" size="sm">
                            Encrypted
                          </Badge>
                        )}
                      </Group>

                      <Text size="sm" c="dimmed">
                        {formatFileSize(file.size)} • {file.content_type}
                      </Text>

                      {file.description && (
                        <Text size="sm" lineClamp={2}>
                          {file.description}
                        </Text>
                      )}

                      <Text size="xs" c="dimmed">
                        Uploaded: {formatDate(file.created_at)}
                      </Text>

                      <Group gap="xs" mt="xs">
                        <ActionIcon
                          variant="light"
                          color="blue"
                          onClick={() => handleDownload(file)}
                          title="Download"
                        >
                          <IconDownload size={16} />
                        </ActionIcon>
                        <ActionIcon
                          variant="light"
                          color="gray"
                          onClick={() => showDetails(file)}
                          title="Details"
                        >
                          <IconInfoCircle size={16} />
                        </ActionIcon>
                        {file.uploader === gpg.getCurrentFingerprint() && (
                          <ActionIcon
                            variant="light"
                            color="red"
                            onClick={() => handleDelete(file.id)}
                            title="Delete"
                          >
                            <IconTrash size={16} />
                          </ActionIcon>
                        )}
                      </Group>
                    </Stack>
                  </Card>
                </Grid.Col>
              ))}
            </Grid>
          </Stack>
        </Card>
      </Stack>

      {/* Details Modal */}
      <Modal
        opened={modalOpened}
        onClose={() => setModalOpened(false)}
        title="File Details"
        size="lg"
      >
        {selectedMetadata && (
          <Stack gap="md">
            <div>
              <Text size="sm" fw={500}>
                Filename
              </Text>
              <Text size="sm">{selectedMetadata.filename}</Text>
            </div>
            <div>
              <Text size="sm" fw={500}>
                Size
              </Text>
              <Text size="sm">{formatFileSize(selectedMetadata.size)}</Text>
            </div>
            <div>
              <Text size="sm" fw={500}>
                Content Type
              </Text>
              <Text size="sm">{selectedMetadata.content_type}</Text>
            </div>
            <div>
              <Text size="sm" fw={500}>
                Content Hash
              </Text>
              <Text size="xs" style={{ wordBreak: "break-all" }}>
                {selectedMetadata.content_hash}
              </Text>
            </div>
            <div>
              <Text size="sm" fw={500}>
                Uploader
              </Text>
              <Text size="xs" style={{ wordBreak: "break-all" }}>
                {selectedMetadata.uploader}
              </Text>
            </div>
            <div>
              <Text size="sm" fw={500}>
                Uploaded At
              </Text>
              <Text size="sm">{formatDate(selectedMetadata.created_at)}</Text>
            </div>
            {selectedMetadata.description && (
              <div>
                <Text size="sm" fw={500}>
                  Description
                </Text>
                <Text size="sm">{selectedMetadata.description}</Text>
              </div>
            )}
            {selectedMetadata.encrypted && (
              <div>
                <Text size="sm" fw={500}>
                  Recipients
                </Text>
                <Text size="xs" style={{ wordBreak: "break-all" }}>
                  {selectedMetadata.recipients.join(", ")}
                </Text>
              </div>
            )}
          </Stack>
        )}
      </Modal>
    </Container>
  );
}
