"use client";

import { useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { isLoggedIn } from "@/lib/auth";
import { useI18n } from "@/lib/i18n";

const GETTING_STARTED_FLAG = "colign:show-getting-started";

export default function AuthCallbackPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { t } = useI18n();

  useEffect(() => {
    // Cookies are set by the server before redirect — just check and navigate.
    if (isLoggedIn()) {
      if (searchParams.get("first") === "1") {
        sessionStorage.setItem(GETTING_STARTED_FLAG, "1");
      }
      const pendingInvite = sessionStorage.getItem("pending_invite_token");
      if (pendingInvite) {
        sessionStorage.removeItem("pending_invite_token");
        router.push(`/invite/${pendingInvite}`);
      } else {
        router.push("/");
      }
    } else {
      router.push("/auth");
    }
  }, [router, searchParams]);

  return (
    <div className="flex min-h-screen items-center justify-center">
      <p>{t("common.authenticating")}</p>
    </div>
  );
}
