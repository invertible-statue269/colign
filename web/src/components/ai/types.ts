export type AIChatMode = "proposal" | "ac" | "general";

export interface ChatProposalResult {
  problem: string;
  scope: string;
  outOfScope: string;
}

export interface ChatACResult {
  scenario: string;
  steps: { keyword: string; text: string }[];
}

export interface ChatMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  result?: ChatProposalResult | ChatACResult[];
  appliedAt?: string;
}
