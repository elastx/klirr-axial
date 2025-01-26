import { MantineProvider } from "@mantine/core";
import { GPGService } from "./services/gpg";
import { KeyGeneration } from "./components/KeyGeneration";
import { UserList } from "./components/UserList";
import { Layout } from "./components/Layout";
import "@mantine/core/styles.css";
import { BrowserRouter } from "react-router-dom";

export default function App() {
  const gpg = GPGService.getInstance();

  return (
    <BrowserRouter>
      <MantineProvider
        defaultColorScheme="dark"
        theme={{
          primaryColor: "teal",
          primaryShade: 6,
          colors: {
            dark: [
              "#C1C2C5",
              "#A6A7AB",
              "#909296",
              "#5c5f66",
              "#373A40",
              "#2C2E33",
              "#25262b",
              "#1A1B1E",
              "#141517",
              "#101113",
            ],
          },
        }}
      >
        {!gpg.isKeyLoaded() ? <KeyGeneration /> : <Layout />}
      </MantineProvider>
    </BrowserRouter>
  );
}
