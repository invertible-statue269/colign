"use client";

import { useI18n } from "@/lib/i18n";

interface AIStreamingPreviewProps {
  sections: { problem?: string; scope?: string; outOfScope?: string };
  isStreaming: boolean;
}

const SECTION_KEYS = ["problem", "scope", "outOfScope"] as const;

const SECTION_I18N: Record<string, string> = {
  problem: "proposal.problem",
  scope: "proposal.scope",
  outOfScope: "proposal.outOfScope",
};

export function AIStreamingPreview({ sections, isStreaming }: AIStreamingPreviewProps) {
  const { t } = useI18n();

  // Find the last non-empty section to place the cursor
  const lastNonEmptyKey = isStreaming
    ? [...SECTION_KEYS].reverse().find((k) => sections[k]?.trim())
    : null;

  return (
    <div className="space-y-3" aria-live="polite">
      {SECTION_KEYS.map((key) => {
        const text = sections[key] || "";
        const showCursor = isStreaming && key === lastNonEmptyKey;

        return (
          <div key={key} className="rounded-lg border border-border/30 bg-card/50 px-4 py-3">
            <div className="mb-1.5 text-xs font-medium text-muted-foreground">
              {t(SECTION_I18N[key])}
            </div>
            <div className="min-h-[2rem] text-sm leading-relaxed text-foreground whitespace-pre-wrap">
              {text}
              {showCursor && (
                <span className="ml-0.5 inline-block h-4 w-0.5 animate-pulse bg-primary align-middle" />
              )}
              {!text && !isStreaming && <span className="text-muted-foreground/40">—</span>}
            </div>
          </div>
        );
      })}
    </div>
  );
}
