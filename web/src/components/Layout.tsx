import { ReactNode, useState } from "react";
import {
  AppShell,
  Text,
  Burger,
  useMantineTheme,
  Group,
  UnstyledButton,
  ThemeIcon,
  rem,
  Box,
  Title,
} from "@mantine/core";
import {
  IconUsers,
  IconMessages,
  IconKey,
  IconLogout,
} from "@tabler/icons-react";
import { GPGService } from "../services/gpg";
import classes from "./Layout.module.css";

interface MainLinkProps {
  icon: ReactNode;
  color: string;
  label: string;
  active?: boolean;
  onClick?: () => void;
}

function MainLink({ icon, color, label, active, onClick }: MainLinkProps) {
  return (
    <UnstyledButton
      onClick={onClick}
      className={`${classes.mainLink} ${active ? classes.active : ""}`}
    >
      <Group>
        <ThemeIcon color={color} variant="light" size={30}>
          {icon}
        </ThemeIcon>
        <Text size="sm">{label}</Text>
      </Group>
    </UnstyledButton>
  );
}

interface LayoutProps {
  children: ReactNode;
}

export function Layout({ children }: LayoutProps) {
  const theme = useMantineTheme();
  const [opened, setOpened] = useState(false);
  const [activeLink, setActiveLink] = useState("users");
  const gpg = GPGService.getInstance();

  const data = [
    {
      icon: <IconUsers size={rem(18)} />,
      color: "blue",
      label: "Users",
      id: "users",
    },
    {
      icon: <IconMessages size={rem(18)} />,
      color: "teal",
      label: "Messages",
      id: "messages",
    },
    {
      icon: <IconKey size={rem(18)} />,
      color: "violet",
      label: "Key Management",
      id: "key",
    },
  ];

  const handleLogout = () => {
    gpg.clearSavedKey();
    window.location.reload();
  };

  return (
    <AppShell
      header={{ height: 60 }}
      navbar={{
        width: 250,
        breakpoint: "sm",
        collapsed: { mobile: !opened },
      }}
      padding="md"
    >
      <AppShell.Header p="md">
        <Group justify="space-between" h="100%">
          <Group>
            <Burger
              opened={opened}
              onClick={() => setOpened((o) => !o)}
              hiddenFrom="sm"
              size="sm"
            />
            <Title order={3}>Axial BBS</Title>
          </Group>

          <Text size="sm" c="dimmed">
            {gpg.getCurrentFingerprint()?.slice(0, 8)}
          </Text>
        </Group>
      </AppShell.Header>

      <AppShell.Navbar p="md">
        <AppShell.Section grow>
          {data.map((link) => (
            <Box key={link.label} mb="xs">
              <MainLink
                {...link}
                active={activeLink === link.id}
                onClick={() => setActiveLink(link.id)}
              />
            </Box>
          ))}
        </AppShell.Section>

        <AppShell.Section>
          <MainLink
            icon={<IconLogout size={rem(18)} />}
            color="red"
            label="Logout"
            onClick={handleLogout}
          />
        </AppShell.Section>
      </AppShell.Navbar>

      <AppShell.Main>{children}</AppShell.Main>
    </AppShell>
  );
}
