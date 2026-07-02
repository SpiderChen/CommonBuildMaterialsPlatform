import React from "react";
import ReactDOM from "react-dom/client";
import "./components/ui/ui.css";
import "./styles/app.css";
import "./styles.css";
import { App } from "./App";
import { MessageBoxProvider, MessageProvider } from "./components";
import { disableNativeContextMenu, disablePageZoom } from "./disablePageZoom";

const runtimeWindow = window as Window & { runtime?: unknown };
document.documentElement.dataset.runtime = runtimeWindow.runtime ? "wails" : "web";
disablePageZoom();
disableNativeContextMenu();

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <MessageBoxProvider>
    <MessageProvider>
      <App />
    </MessageProvider>
  </MessageBoxProvider>
);
