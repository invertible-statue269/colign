"use client";

import { useRef, useCallback } from "react";
import { SendHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useI18n } from "@/lib/i18n";

interface ChatInputProps {
  onSend: (message: string) => void;
  disabled?: boolean;
}

export function ChatInput({ onSend, disabled }: ChatInputProps) {
  const { t } = useI18n();
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const adjustHeight = useCallback(() => {
    const el = textareaRef.current;
    if (!el) return;
    el.style.height = "auto";
    el.style.height = `${Math.min(el.scrollHeight, 120)}px`;
  }, []);

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  }

  function handleSend() {
    const el = textareaRef.current;
    if (!el) return;
    const value = el.value.trim();
    if (!value || disabled) return;
    onSend(value);
    el.value = "";
    el.style.height = "auto";
  }

  return (
    <div className="border-t border-border/40 px-4 py-3">
      <div className="flex items-end gap-2">
        <textarea
          ref={textareaRef}
          rows={1}
          disabled={disabled}
          placeholder={t("ai.chatPlaceholder")}
          onInput={adjustHeight}
          onKeyDown={handleKeyDown}
          className="flex-1 resize-none rounded-lg border border-border/30 bg-background/50 px-3 py-2 text-sm outline-none placeholder:text-muted-foreground/40 focus:border-primary/50 disabled:opacity-50"
        />
        <Button
          size="sm"
          onClick={handleSend}
          disabled={disabled}
          className="cursor-pointer shrink-0"
        >
          <SendHorizontal className="size-4" />
        </Button>
      </div>
    </div>
  );
}
