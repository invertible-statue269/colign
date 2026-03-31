"use client";

import { createContext, useCallback, useContext, useEffect, useState } from "react";
import type { AIChatMode } from "./types";

interface AIPanelContextValue {
  isOpen: boolean;
  mode: AIChatMode;
  open: (mode?: AIChatMode) => void;
  close: () => void;
  toggle: () => void;
  setMode: (mode: AIChatMode) => void;
}

const STORAGE_KEY = "colign-ai-panel-open";

const AIPanelContext = createContext<AIPanelContextValue>({
  isOpen: false,
  mode: "general",
  open: () => {},
  close: () => {},
  toggle: () => {},
  setMode: () => {},
});

export function AIPanelProvider({ children }: { children: React.ReactNode }) {
  const [isOpen, setIsOpen] = useState(false);
  const [mode, setMode] = useState<AIChatMode>("general");

  // Restore persisted state on mount
  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored === "true") {
        setIsOpen(true);
      }
    } catch {
      // localStorage unavailable
    }
  }, []);

  // Persist open/close state
  useEffect(() => {
    try {
      localStorage.setItem(STORAGE_KEY, String(isOpen));
    } catch {
      // localStorage unavailable
    }
  }, [isOpen]);

  const open = useCallback((m?: AIChatMode) => {
    if (m) setMode(m);
    setIsOpen(true);
  }, []);

  const close = useCallback(() => {
    setIsOpen(false);
  }, []);

  const toggle = useCallback(() => {
    setIsOpen((prev) => !prev);
  }, []);

  return (
    <AIPanelContext.Provider value={{ isOpen, mode, open, close, toggle, setMode }}>
      {children}
    </AIPanelContext.Provider>
  );
}

export function useAIPanel() {
  return useContext(AIPanelContext);
}
