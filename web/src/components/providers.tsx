"use client";

import { I18nProvider } from "@/lib/i18n";
import { OrgProvider } from "@/lib/org-context";

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <I18nProvider>
      <OrgProvider>{children}</OrgProvider>
    </I18nProvider>
  );
}
