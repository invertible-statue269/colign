"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { Check, RefreshCw, Sparkles, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { AIStreamingPreview } from "@/components/ai/ai-streaming-preview";
import { streamProposal } from "@/lib/ai";
import { useI18n } from "@/lib/i18n";
import { showError } from "@/lib/toast";

type GeneratorState = "idle" | "loading" | "previewing";

interface ProposalSections {
  problem: string;
  scope: string;
  outOfScope: string;
}

const EMPTY_SECTIONS: ProposalSections = {
  problem: "",
  scope: "",
  outOfScope: "",
};

interface AIProposalGeneratorProps {
  changeId: number | bigint;
  onApply: (sections: ProposalSections) => void;
  hasExistingContent: boolean;
}

export function AIProposalGenerator({
  changeId,
  onApply,
  hasExistingContent,
}: AIProposalGeneratorProps) {
  const { t } = useI18n();
  const [state, setState] = useState<GeneratorState>("idle");
  const [description, setDescription] = useState("");
  const [sections, setSections] = useState<ProposalSections>(EMPTY_SECTIONS);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [showConfirm, setShowConfirm] = useState(false);
  const [regenerateCooldown, setRegenerateCooldown] = useState(false);

  const abortRef = useRef<AbortController | null>(null);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      abortRef.current?.abort();
    };
  }, []);

  const startGeneration = useCallback(async () => {
    abortRef.current?.abort();
    const abort = new AbortController();
    abortRef.current = abort;

    setState("loading");
    setErrorMessage(null);
    setSections(EMPTY_SECTIONS);

    try {
      for await (const chunk of streamProposal(changeId, description, abort.signal)) {
        if (abort.signal.aborted) break;
        const key = chunk.section as keyof ProposalSections;
        if (key in EMPTY_SECTIONS) {
          setSections((prev) => ({ ...prev, [key]: prev[key] + chunk.chunk }));
        }
      }
      if (!abort.signal.aborted) {
        setState("previewing");
      }
    } catch (err) {
      if (abort.signal.aborted) return;
      const msg = err instanceof Error ? err.message : String(err);
      setErrorMessage(msg);
      setState("idle");
      showError(t("ai.connectionError"), err);
    }
  }, [changeId, description, t]);

  function handleGenerate() {
    if (hasExistingContent) {
      setShowConfirm(true);
    } else {
      startGeneration();
    }
  }

  function handleConfirmRegenerate() {
    setShowConfirm(false);
    startGeneration();
  }

  function handleApply() {
    onApply(sections);
    setState("idle");
    setSections(EMPTY_SECTIONS);
    setDescription("");
  }

  function handleCancel() {
    abortRef.current?.abort();
    setState("idle");
    setSections(EMPTY_SECTIONS);
    setErrorMessage(null);
  }

  function handleRegenerate() {
    if (regenerateCooldown) return;
    setRegenerateCooldown(true);
    setTimeout(() => setRegenerateCooldown(false), 3000);
    startGeneration();
  }

  // --- Confirm dialog overlay ---
  if (showConfirm) {
    return (
      <div className="rounded-xl border border-border/40 bg-card/50 p-4 space-y-3">
        <p className="text-sm text-foreground">{t("ai.confirmRegenerate")}</p>
        <div className="flex gap-2">
          <Button size="sm" onClick={handleConfirmRegenerate}>
            <Check className="size-3.5" />
            {t("common.continue")}
          </Button>
          <Button size="sm" variant="ghost" onClick={() => setShowConfirm(false)}>
            <X className="size-3.5" />
            {t("ai.cancel")}
          </Button>
        </div>
      </div>
    );
  }

  // --- Loading / Previewing state ---
  if (state === "loading" || state === "previewing") {
    return (
      <div className="rounded-xl border border-border/40 bg-card/50 p-4 space-y-3">
        <div className="flex items-center gap-2">
          <Sparkles className="size-4 text-primary" />
          <span className="text-sm font-medium text-foreground">
            {state === "loading" ? t("ai.loading") : t("ai.generateProposal")}
          </span>
        </div>

        <AIStreamingPreview sections={sections} isStreaming={state === "loading"} />

        {state === "previewing" && (
          <div className="flex gap-2 pt-1">
            <Button size="sm" onClick={handleApply}>
              <Check className="size-3.5" />
              {t("ai.apply")}
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={handleRegenerate}
              disabled={regenerateCooldown}
            >
              <RefreshCw className="size-3.5" />
              {t("ai.regenerateProposal")}
            </Button>
            <Button size="sm" variant="ghost" onClick={handleCancel}>
              <X className="size-3.5" />
              {t("ai.cancel")}
            </Button>
          </div>
        )}
      </div>
    );
  }

  // --- Idle state with existing content: small regenerate button ---
  if (hasExistingContent) {
    return (
      <div className="flex items-center justify-end">
        {errorMessage && <span className="mr-2 text-xs text-destructive">{errorMessage}</span>}
        <Button size="sm" variant="ghost" onClick={handleGenerate}>
          <RefreshCw className="size-3.5" />
          {t("ai.regenerateProposal")}
        </Button>
      </div>
    );
  }

  // --- Idle state empty: input + generate button ---
  return (
    <div className="rounded-xl border border-dashed border-primary/30 bg-primary/5 p-4 space-y-3">
      <div className="flex items-center gap-2">
        <Sparkles className="size-4 text-primary/70" />
        <span className="text-sm font-medium text-foreground">{t("ai.emptyStateTitle")}</span>
      </div>
      <p className="text-xs text-muted-foreground">{t("ai.emptyStateDescription")}</p>
      <div className="flex gap-2">
        <input
          type="text"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && description.trim()) {
              e.preventDefault();
              startGeneration();
            }
          }}
          placeholder={t("ai.descriptionPlaceholder")}
          className="flex-1 rounded-lg border border-border/30 bg-background/50 px-3 py-1.5 text-sm outline-none placeholder:text-muted-foreground/40 focus:border-primary/50"
        />
        <Button size="sm" onClick={handleGenerate} disabled={!description.trim()}>
          <Sparkles className="size-3.5" />
          {t("ai.generateProposal")}
        </Button>
      </div>
      {errorMessage && (
        <p className="text-xs text-destructive" role="alert">
          {errorMessage}
        </p>
      )}
    </div>
  );
}
