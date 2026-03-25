"use client";

import { useMemo, useRef, useState } from "react";
import { cn } from "@/lib/utils";

export interface MentionMember {
  userId: bigint;
  userName: string;
}

interface MentionTextareaProps {
  value: string;
  onChange: (value: string) => void;
  members: MentionMember[];
  placeholder?: string;
  className?: string;
  rows?: number;
  autoFocus?: boolean;
  disabled?: boolean;
  onSubmit?: () => void;
  submitShortcut?: "enter" | "mod-enter";
  onEscape?: () => void;
}

interface MentionMatch {
  query: string;
  start: number;
  end: number;
}

function getMentionMatch(value: string, caret: number): MentionMatch | null {
  const beforeCaret = value.slice(0, caret);
  const match = /(^|\s)@([^\s@]*)$/.exec(beforeCaret);
  if (!match) return null;
  const query = match[2] ?? "";
  const start = beforeCaret.length - query.length - 1;
  return { query, start, end: beforeCaret.length };
}

export function MentionTextarea({
  value,
  onChange,
  members,
  placeholder,
  className,
  rows = 2,
  autoFocus,
  disabled,
  onSubmit,
  submitShortcut,
  onEscape,
}: MentionTextareaProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const [caret, setCaret] = useState(0);
  const [highlightedIndex, setHighlightedIndex] = useState(0);
  const [dismissedMentionKey, setDismissedMentionKey] = useState<string | null>(null);

  const mentionMatch = useMemo(() => getMentionMatch(value, caret), [value, caret]);
  const mentionKey = mentionMatch ? `${mentionMatch.start}:${mentionMatch.end}:${mentionMatch.query}` : null;
  const suggestions = useMemo(() => {
    if (!mentionMatch) return [];
    const query = mentionMatch.query.trim().toLowerCase();
    const uniqueMembers = members.filter(
      (member, index, list) =>
        !!member.userName &&
        list.findIndex((candidate) => candidate.userId === member.userId) === index,
    );

    return uniqueMembers
      .filter((member) => {
        if (!query) return true;
        return member.userName.toLowerCase().includes(query);
      })
      .sort((a, b) => {
        const aStarts = query ? a.userName.toLowerCase().startsWith(query) : false;
        const bStarts = query ? b.userName.toLowerCase().startsWith(query) : false;
        if (aStarts !== bStarts) return aStarts ? -1 : 1;
        return a.userName.localeCompare(b.userName);
      })
      .slice(0, 8);
  }, [members, mentionMatch]);
  const isMentionDismissed = mentionKey !== null && mentionKey === dismissedMentionKey;
  const activeSuggestions = isMentionDismissed ? [] : suggestions;
  const safeHighlightedIndex =
    activeSuggestions.length === 0 ? 0 : Math.min(highlightedIndex, activeSuggestions.length - 1);

  function syncDismissedMention(nextValue: string, nextCaret: number) {
    if (!dismissedMentionKey) return;
    const nextMatch = getMentionMatch(nextValue, nextCaret);
    const nextKey = nextMatch ? `${nextMatch.start}:${nextMatch.end}:${nextMatch.query}` : null;
    if (nextKey !== dismissedMentionKey) {
      setDismissedMentionKey(null);
      setHighlightedIndex(0);
    }
  }

  function syncCaret() {
    const el = textareaRef.current;
    if (!el) return;
    const nextCaret = el.selectionStart ?? 0;
    syncDismissedMention(value, nextCaret);
    setCaret(nextCaret);
  }

  function selectMember(index: number) {
    const match = mentionMatch;
    const member = activeSuggestions[index];
    const el = textareaRef.current;
    if (!match || !member || !el) return;

    const nextValue = `${value.slice(0, match.start)}@${member.userName} ${value.slice(match.end)}`;
    const nextCaret = match.start + member.userName.length + 2;

    setDismissedMentionKey(null);
    onChange(nextValue);
    requestAnimationFrame(() => {
      el.focus();
      el.setSelectionRange(nextCaret, nextCaret);
      setCaret(nextCaret);
    });
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (activeSuggestions.length > 0 && mentionMatch) {
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setHighlightedIndex((prev) => (prev + 1) % activeSuggestions.length);
        return;
      }
      if (e.key === "ArrowUp") {
        e.preventDefault();
        setHighlightedIndex((prev) => (prev - 1 + activeSuggestions.length) % activeSuggestions.length);
        return;
      }
      if (e.key === "Enter" || e.key === "Tab") {
        e.preventDefault();
        selectMember(safeHighlightedIndex);
        return;
      }
      if (e.key === "Escape") {
        e.preventDefault();
        setDismissedMentionKey(mentionKey);
        return;
      }
    }

    if (submitShortcut === "enter" && e.key === "Enter" && !e.shiftKey && !e.metaKey && !e.ctrlKey) {
      e.preventDefault();
      onSubmit?.();
      return;
    }

    if (submitShortcut === "mod-enter" && e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
      e.preventDefault();
      onSubmit?.();
      return;
    }

    if (e.key === "Escape") {
      onEscape?.();
    }
  }

  return (
    <div className="relative">
      <textarea
        ref={textareaRef}
        autoFocus={autoFocus}
        value={value}
        disabled={disabled}
        rows={rows}
        placeholder={placeholder}
        onChange={(e) => {
        onChange(e.target.value);
          const nextCaret = e.target.selectionStart ?? 0;
          syncDismissedMention(e.target.value, nextCaret);
          setCaret(nextCaret);
        }}
        onClick={syncCaret}
        onKeyUp={syncCaret}
        onSelect={syncCaret}
        onKeyDown={handleKeyDown}
        className={className}
      />
      {mentionMatch && activeSuggestions.length > 0 && (
        <div className="absolute left-0 right-0 top-full z-40 mt-1 overflow-hidden rounded-md border border-border/60 bg-popover shadow-lg">
          <div className="max-h-56 overflow-y-auto py-1">
            {activeSuggestions.map((member, index) => (
              <button
                key={String(member.userId)}
                type="button"
                onMouseDown={(e) => {
                  e.preventDefault();
                  selectMember(index);
                }}
                className={cn(
                  "flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-accent",
                  index === safeHighlightedIndex && "bg-accent",
                )}
              >
                <div className="flex size-6 items-center justify-center rounded-full bg-primary/15 text-[10px] font-bold uppercase text-primary">
                  {member.userName.charAt(0)}
                </div>
                <span className="truncate">@{member.userName}</span>
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
