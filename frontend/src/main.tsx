import "@/styles/fonts.css";
import { AuthProvider } from "@/providers/AuthProvider";
import { Provider } from "@/components/ui/provider";
import { Toaster } from "@/components/ui/toaster";
import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <AuthProvider>
      <Provider>
        <App />
        <Toaster />
      </Provider>
    </AuthProvider>
  </React.StrictMode>,
);
