import { MantineProvider } from "@mantine/core";
import { GPGService } from "./services/gpg";
import { Layout } from "./components/Layout";
import "@mantine/core/styles.css";
import "./index.css";
import { BrowserRouter } from "react-router-dom";
import Signup from "./components/Signup";

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
        {!gpg.isKeyLoaded() ? <Signup /> : <Layout />}
      </MantineProvider>
    </BrowserRouter>
  );
}
