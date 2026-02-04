import React from "react";
import { KeyGeneration, GeneratedKey } from "./KeyGeneration";
import { GPGService } from "../services/gpg";
import { useAppStore } from "../store";
import { Center } from "@mantine/core";

const Signup: React.FC = () => {
  const gpg = GPGService.getInstance();
  const register = useAppStore((s) => s.users.register);

  const handleRegister = async () => {
    try {
      const userInfo = await gpg.getCurrentUserInfo();
      if (!userInfo) throw new Error("No key loaded");
      await register(userInfo);
      window.location.reload();
    } catch (error) {
      console.error("Failed to register:", error);
    }
  };

  const handleImportPrivateKey = async (armoredKey: string) => {
    await gpg.importPrivateKey(armoredKey);
    await handleRegister();
  };

  const handleSelectGeneratedKey = async (key: GeneratedKey) => {
    await gpg.importPrivateKey(key.privateKey);
    await handleRegister();
  };

  return (
    <Center w={"100%"} mih={"50vh"}>
      <KeyGeneration
        onImportPrivateKey={handleImportPrivateKey}
        onSelectGeneratedKey={handleSelectGeneratedKey}
      />
    </Center>
  );
};

export default Signup;
