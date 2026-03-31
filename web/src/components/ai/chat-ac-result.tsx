"use client";

import { useState } from "react";
import { Check, CheckSquare, ListChecks, Square } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useI18n } from "@/lib/i18n";
import type { ChatACResult } from "./types";

// Custom event for applying AC from AI panel
export const AI_APPLY_AC_EVENT = "colign:ai-apply-ac";

export function dispatchApplyAC(results: ChatACResult[]) {
  window.dispatchEvent(
    new CustomEvent(AI_APPLY_AC_EVENT, { detail: results }),
  );
}

const keywordColor: Record<string, string> = {
  Given: "text-blue-400",
  When: "text-amber-400",
  Then: "text-emerald-400",
  And: "text-purple-400",
  But: "text-red-400",
};

interface ChatACResultCardProps {
  results: ChatACResult[];
  appliedAt?: string;
  onApply: (selected: ChatACResult[]) => void;
}

export function ChatACResultCard({ results, appliedAt, onApply }: ChatACResultCardProps) {
  const { t } = useI18n();
  const [selected, setSelected] = useState<boolean[]>(results.map(() => true));

  const allSelected = selected.length > 0 && selected.every(Boolean);
  const selectedCount = selected.filter(Boolean).length;

  function toggleSelect(index: number) {
    setSelected((prev) => prev.map((v, i) => (i === index ? !v : v)));
  }

  function handleApply() {
    const selectedACs = results.filter((_, i) => selected[i]);
    if (selectedACs.length > 0) {
      onApply(selectedACs);
    }
  }

  return (
    <div className="mt-2 rounded-lg border border-primary/20 bg-primary/5 p-3 space-y-2">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-1.5 text-xs font-medium text-primary">
          <ListChecks className="size-3.5" />
          {t("ai.modeAC")} ({results.length})
        </div>
        {!appliedAt && (
          <button
            onClick={() => setSelected(results.map(() => !allSelected))}
            className="flex cursor-pointer items-center gap-1 text-[10px] text-muted-foreground hover:text-foreground transition-colors"
          >
            {allSelected ? (
              <>
                <CheckSquare className="size-3" />
                {t("ai.deselectAll")}
              </>
            ) : (
              <>
                <Square className="size-3" />
                {t("ai.selectAll")}
              </>
            )}
          </button>
        )}
      </div>

      {results.map((ac, index) => {
        const isSelected = selected[index] ?? false;
        return (
          <button
            key={index}
            onClick={() => !appliedAt && toggleSelect(index)}
            disabled={!!appliedAt}
            className={`w-full rounded-md border p-2 text-left transition-all ${
              appliedAt
                ? "border-border/30 opacity-70"
                : isSelected
                  ? "cursor-pointer border-primary/30 bg-background/50"
                  : "cursor-pointer border-border/30 opacity-50 hover:border-border/50"
            }`}
          >
            <div className="flex items-start gap-2">
              {!appliedAt && (
                <div
                  className={`mt-0.5 flex size-3.5 shrink-0 items-center justify-center rounded border transition-colors ${
                    isSelected ? "border-primary bg-primary" : "border-muted-foreground/30"
                  }`}
                >
                  {isSelected && <Check className="size-2 text-primary-foreground" />}
                </div>
              )}
              <div className="min-w-0 flex-1">
                <p className="text-xs font-medium">{ac.scenario}</p>
                <div className="mt-1 space-y-0.5">
                  {ac.steps.map((step, si) => (
                    <div
                      key={si}
                      className={`flex gap-1 text-[10px] ${step.keyword === "And" || step.keyword === "But" ? "pl-3" : ""}`}
                    >
                      <span className={`shrink-0 font-semibold ${keywordColor[step.keyword] || "text-muted-foreground"}`}>
                        {step.keyword}
                      </span>
                      <span className="text-foreground/60">{step.text}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </button>
        );
      })}

      {appliedAt ? (
        <div className="flex items-center gap-1 text-xs text-emerald-500">
          <Check className="size-3" />
          {t("ai.apply")}
        </div>
      ) : (
        <Button
          size="sm"
          onClick={handleApply}
          disabled={selectedCount === 0}
          className="cursor-pointer"
        >
          <Check className="size-3.5" />
          {t("ai.applySelected")} {selectedCount > 0 && `(${selectedCount})`}
        </Button>
      )}
    </div>
  );
}
