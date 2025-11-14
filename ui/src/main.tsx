import React from "react";
import { createRoot } from "react-dom/client";
import AppRouter from "@app/router";

const root = document.getElementById("root");
if (root) {
  createRoot(root).render(<AppRouter />);
}

