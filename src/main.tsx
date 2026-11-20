import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { HashRouter } from "react-router";

import { Toaster } from "@/components/ui/toast";
import { App } from "./app";
import { restoreLastRoute } from "./app-routes";
import "./style.css";

restoreLastRoute();

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <HashRouter>
      <App />
      <Toaster />
    </HashRouter>
  </StrictMode>,
);
