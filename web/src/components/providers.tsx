"use client";

import { I18nProvider } from "@/lib/i18n";
import { OrgProvider } from "@/lib/org-context";
import { EventProvider } from "@/lib/events";
import { TooltipProvider } from "@/components/ui/tooltip";
import { SidebarLayout } from "@/components/layout/sidebar-layout";

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <I18nProvider>
      <OrgProvider>
        <EventProvider>
          <TooltipProvider>
            <SidebarLayout>{children}</SidebarLayout>
          </TooltipProvider>
        </EventProvider>
      </OrgProvider>
    </I18nProvider>
  );
}
