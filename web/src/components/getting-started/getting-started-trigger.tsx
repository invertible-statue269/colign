"use client";

import { LifeBuoy } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useI18n } from "@/lib/i18n";

interface GettingStartedTriggerProps {
  onOpen: () => void;
}

export function GettingStartedTrigger({ onOpen }: GettingStartedTriggerProps) {
  const { t } = useI18n();

  return (
    <Button variant="outline" onClick={onOpen}>
      <LifeBuoy className="h-4 w-4" />
      {t("gettingStarted.openButton")}
    </Button>
  );
}
