"use client";

import { useCallback, useState } from "react";
import { SpecEditor } from "@/components/editor/spec-editor";
import { sddTemplates } from "@/components/editor/templates";

interface DocumentTabProps {
  changeId: bigint;
  docType: "proposal" | "design" | "spec" | "tasks";
  initialContent?: string;
}

export function DocumentTab({ changeId, docType, initialContent }: DocumentTabProps) {
  const [content] = useState(initialContent || sddTemplates[docType] || "");

  const handleSave = useCallback(
    async (newContent: string) => {
      try {
        await fetch(`${process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"}/api/documents`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            change_id: Number(changeId),
            type: docType,
            content: newContent,
          }),
        });
      } catch {
        // retry handled by editor
      }
    },
    [changeId, docType]
  );

  return (
    <div className="py-4">
      <SpecEditor
        initialContent={content}
        placeholder={`Start writing your ${docType}...`}
        onSave={handleSave}
      />
    </div>
  );
}
