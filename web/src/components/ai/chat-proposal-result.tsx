"use client";

import { Check, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useI18n } from "@/lib/i18n";
import type { ChatProposalResult } from "./types";

// Custom event for applying proposal from AI panel
export const AI_APPLY_PROPOSAL_EVENT = "colign:ai-apply-proposal";

export function dispatchApplyProposal(result: ChatProposalResult) {
  window.dispatchEvent(
    new CustomEvent(AI_APPLY_PROPOSAL_EVENT, { detail: result }),
  );
}

interface ChatProposalResultCardProps {
  result: ChatProposalResult;
  appliedAt?: string;
  onApply: () => void;
}

const sectionI18nKeys: Record<keyof ChatProposalResult, string> = {
  problem: "proposal.problem",
  scope: "proposal.scope",
  outOfScope: "proposal.outOfScope",
};

export function ChatProposalResultCard({ result, appliedAt, onApply }: ChatProposalResultCardProps) {
  const { t } = useI18n();

  return (
    <div className="mt-2 rounded-lg border border-primary/20 bg-primary/5 p-3 space-y-2">
      <div className="flex items-center gap-1.5 text-xs font-medium text-primary">
        <FileText className="size-3.5" />
        {t("change.proposal")}
      </div>

      {(["problem", "scope", "outOfScope"] as const).map((key) => {
        const content = result[key];
        if (!content) return null;
        return (
          <div key={key}>
            <div className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
              {t(sectionI18nKeys[key])}
            </div>
            <div className="mt-0.5 text-xs text-foreground/80 line-clamp-3">
              {content.replace(/<[^>]*>/g, "")}
            </div>
          </div>
        );
      })}

      {appliedAt ? (
        <div className="flex items-center gap-1 text-xs text-emerald-500">
          <Check className="size-3" />
          {t("ai.apply")}
        </div>
      ) : (
        <Button size="sm" onClick={onApply} className="cursor-pointer">
          <Check className="size-3.5" />
          {t("ai.applyResult")}
        </Button>
      )}
    </div>
  );
}
