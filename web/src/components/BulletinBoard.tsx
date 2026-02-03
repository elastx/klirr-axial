import {
  Button,
  Group,
  Paper,
  Stack,
  Text,
  TextInput,
  Textarea,
} from "@mantine/core";
import {
  IconArrowBack,
  IconCheck,
  IconMessageReply,
  IconQuestionMark,
  IconSend,
  IconX,
} from "@tabler/icons-react";
import { useEffect, useState } from "react";
import { APIService } from "../services/api";
import { Message } from "../types";
import UserAvatar from "./avatar/UserAvatar";
import { AxiosError } from "axios";
import { GPGService } from "../services/gpg";

interface NewPostProps {
  onSubmit: () => void;
  parentId?: number;
  initialTopic?: string;
  onCancel?: () => void;
}

function NewPost({ onSubmit, parentId, initialTopic, onCancel }: NewPostProps) {
  const [errorMessage, setErrorMessage] = useState("");
  const [topic, setTopic] = useState(initialTopic || "");
  const [content, setContent] = useState("");
  const api = APIService.getInstance();

  const handleSubmit = async () => {
    GPGService.getInstance()
      .signMessage(content)
      .then(async (signature) => {
        return api
          .sendBulletinPost(topic, content, signature, parentId)
          .then(() => {
            setTopic("");
            setContent("");
            onSubmit();
          })
          .catch((error: AxiosError) => {
            let errorText = error.message;
            const serverError = error.response?.data as string;
            if (serverError) {
              errorText = serverError;
            }
            setErrorMessage(errorText);
          });
      });
  };

  return (
    <Paper p="md" withBorder>
      <Stack>
        {errorMessage && (
          <Text color="red" size="sm">
            {errorMessage}
          </Text>
        )}
        {!parentId && (
          <TextInput
            label="Topic"
            placeholder="Enter topic"
            value={topic}
            onChange={(e) => setTopic(e.currentTarget.value)}
          />
        )}
        <Textarea
          label="Message"
          placeholder="Type your message here"
          minRows={4}
          value={content}
          onChange={(e) => setContent(e.currentTarget.value)}
        />
        <Group justify="flex-end">
          {onCancel && (
            <Button variant="subtle" onClick={onCancel}>
              Cancel
            </Button>
          )}
          <Button
            disabled={(!parentId && !topic) || !content}
            onClick={handleSubmit}
            leftSection={<IconSend size={16} />}
          >
            {parentId ? "Reply" : "Post Message"}
          </Button>
        </Group>
      </Stack>
    </Paper>
  );
}

interface PostProps {
  post: Message;
  posts: Message[];
  onReply: (post: Message) => void;
}

function Post({ post, posts, onReply }: PostProps) {
  // Find replies to this post
  const replies = posts.filter((p) => p.parent_id === post.id);
  const [verificationStatus, setVerificationStatus] = useState<boolean | null>(
    null
  );

  useEffect(() => {
    GPGService.getInstance()
      .verifyMessageSignature(post.content, post.signature, post.author)
      .then((verified) => {
        setVerificationStatus(verified);
      });
  }, [post]);

  return (
    <Paper p="md" withBorder>
      <Stack gap="xs">
        <Group justify="space-between">
          <Group>
            <UserAvatar seed={post.author} size={50} />
            <div>
              <Text fw={500}>{post.topic}</Text>
              <Text size="sm" c="dimmed">
                {new Date(post.created_at).toLocaleString()}
              </Text>
            </div>
          </Group>
          <Group>
            {verificationStatus === true && (
              <IconCheck size={16} color="green" />
            )}
            {verificationStatus === false && <IconX size={16} color="red" />}
            {verificationStatus === null && (
              <IconQuestionMark size={16} color="gray" />
            )}
            <Button
              variant="subtle"
              size="sm"
              leftSection={<IconMessageReply size={16} />}
              onClick={() => onReply(post)}
            >
              Reply
            </Button>
          </Group>
        </Group>
        <Text>{post.content}</Text>
        {replies.length > 0 && (
          <Stack ml={40} mt="sm">
            {replies.map((reply) => (
              <Post
                key={reply.id}
                post={reply}
                posts={posts}
                onReply={onReply}
              />
            ))}
          </Stack>
        )}
      </Stack>
    </Paper>
  );
}

export function BulletinBoard() {
  const [posts, setPosts] = useState<Message[]>([]);
  const [loading, setLoading] = useState(true);
  const [replyTo, setReplyTo] = useState<Message | null>(null);
  const [showNewPost, setShowNewPost] = useState(true);
  const api = APIService.getInstance();

  const loadPosts = async () => {
    try {
      const posts = await api.getBulletinPosts();
      setPosts(posts);
    } catch (error) {
      console.error("Failed to load posts:", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadPosts();
  }, []);

  const handlePostSubmitted = async () => {
    setReplyTo(null);
    setShowNewPost(true);
    await loadPosts();
  };

  const handleReply = (post: Message) => {
    setReplyTo(post);
    setShowNewPost(false);
  };

  return (
    <Stack>
      <Group justify="space-between">
        <Text size="xl" fw={500}>
          Bulletin Board
        </Text>
        {!showNewPost && (
          <Button
            leftSection={<IconArrowBack size={16} />}
            variant="subtle"
            onClick={() => {
              setShowNewPost(true);
              setReplyTo(null);
            }}
          >
            Back to Posts
          </Button>
        )}
      </Group>

      {showNewPost && !replyTo && <NewPost onSubmit={handlePostSubmitted} />}

      {replyTo && (
        <>
          <Paper p="md" withBorder bg="var(--mantine-color-dark-6)">
            <Text size="sm" fw={500} mb="xs">
              Replying to:
            </Text>
            <Text>{replyTo.content}</Text>
          </Paper>
          <NewPost
            onSubmit={handlePostSubmitted}
            parentId={replyTo.id}
            initialTopic={replyTo.topic}
            onCancel={() => setReplyTo(null)}
          />
        </>
      )}

      <Paper p="md" withBorder>
        <Text fw={500} mb="md">
          Recent Posts
        </Text>
        {loading ? (
          <Text c="dimmed">Loading posts...</Text>
        ) : posts.length === 0 ? (
          <Text c="dimmed">No posts yet</Text>
        ) : (
          <Stack>
            {posts
              .filter((post) => !post.parent_id) // Only show top-level posts
              .map((post) => (
                <Post
                  key={post.id}
                  post={post}
                  posts={posts}
                  onReply={handleReply}
                />
              ))}
          </Stack>
        )}
      </Paper>
    </Stack>
  );
}
