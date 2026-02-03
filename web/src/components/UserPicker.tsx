import { useEffect, useMemo, useState } from "react";
import { Combobox, Group, Text, useCombobox, TextInput, Avatar, Loader } from "@mantine/core";
import { useDebouncedValue } from "@mantine/hooks";
import { APIService } from "../services/api";
import { StoredUser } from "../types";

interface UserPickerProps {
  value: StoredUser | null;
  onSelect: (user: StoredUser) => void;
  placeholder?: string;
  limit?: number;
}

function isHexFingerprint(input: string): boolean {
  const s = input.trim().replace(/\s+/g, "");
  // Accept hex strings of length >= 16
  return /^[0-9a-fA-F]{16,}$/.test(s);
}

function userLabel(u: StoredUser): string {
  if (u.name && u.email) return `${u.name} <${u.email}>`;
  if (u.name) return u.name;
  if (u.email) return u.email;
  return u.fingerprint;
}

export function UserPicker({ value, onSelect, placeholder = "Search users by fingerprint, name, or email", limit = 20 }: UserPickerProps) {
  const api = APIService.getInstance();
  const [query, setQuery] = useState("");
  const [debouncedQuery] = useDebouncedValue(query, 300);
  const [results, setResults] = useState<StoredUser[]>([]);
  const [loading, setLoading] = useState(false);

  const combobox = useCombobox({
    onDropdownClose: () => combobox.resetSelectedOption(),
    onDropdownOpen: () => combobox.updateSelectedOptionIndex("active"),
  });

  useEffect(() => {
    let cancelled = false;
    async function run() {
      setLoading(true);
      try {
        if (!debouncedQuery) {
          const recent = await api.getRecentUsers(10);
          if (!cancelled) setResults(recent);
        } else {
          const users = await api.searchUsers(debouncedQuery, limit, 0);
          if (!cancelled) setResults(users);
        }
      } catch (e) {
        console.error("User search failed", e);
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    run();
    return () => {
      cancelled = true;
    };
  }, [debouncedQuery, limit]);

  const options = useMemo(() => results.map((u) => ({ key: u.fingerprint, user: u })), [results]);

  const handleSubmit = async () => {
    const input = query.trim();
    if (isHexFingerprint(input)) {
      onSelect({ fingerprint: input, public_key: "", name: undefined, email: undefined });
      combobox.closeDropdown();
      setQuery("");
    }
  };

  return (
    <Combobox store={combobox} onOptionSubmit={(val) => {
      const sel = results.find((u) => u.fingerprint === val);
      if (sel) onSelect(sel);
      combobox.closeDropdown();
      setQuery("");
    }}>
      <Combobox.Target>
        <TextInput
          label="Recipient"
          placeholder={placeholder}
          value={query}
          onChange={(e) => setQuery(e.currentTarget.value)}
          onFocus={() => combobox.openDropdown()}
          onBlur={() => combobox.closeDropdown()}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              e.preventDefault();
              handleSubmit();
            }
          }}
        />
      </Combobox.Target>

      <Combobox.Dropdown>
        <Combobox.Options>
          {loading ? (
            <Group p="sm">
              <Loader size="sm" />
              <Text size="sm">Searching…</Text>
            </Group>
          ) : options.length === 0 ? (
            <Combobox.Empty>Start typing to search users</Combobox.Empty>
          ) : (
            options.map(({ key, user }) => (
              <Combobox.Option value={key} key={key}>
                <Group wrap="nowrap">
                  <Avatar radius="xl" color="blue">{(user.name || user.email || user.fingerprint)[0]}</Avatar>
                  <div style={{ flex: 1 }}>
                    <Text size="sm">{userLabel(user)}</Text>
                    <Text size="xs" c="dimmed">{user.fingerprint}</Text>
                  </div>
                </Group>
              </Combobox.Option>
            ))
          )}
        </Combobox.Options>
      </Combobox.Dropdown>
    </Combobox>
  );
}
