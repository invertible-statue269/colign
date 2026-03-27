import type { Metadata } from "next";
import ChangeDetailClient from "./change-detail-client";

const apiUrl =
  process.env.API_INTERNAL_URL ?? process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

type Props = {
  params: Promise<{ slug: string; changeId: string }>;
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug, changeId } = await params;

  const defaultMeta: Metadata = {
    title: "Colign - AI-powered Spec Writing",
    description:
      "Collaborative Spec-Driven Development platform where developers and non-developers write specs with AI.",
  };

  try {
    const res = await fetch(`${apiUrl}/api/og/projects/${slug}/changes/${changeId}`, {
      next: { revalidate: 60 },
    });
    if (!res.ok) return defaultMeta;

    const data = await res.json();
    const title = `${data.changeName} - ${data.projectName} | Colign`;
    const description = `[${data.stage.charAt(0).toUpperCase() + data.stage.slice(1)}] ${data.changeName} in ${data.projectName}`;

    return {
      title,
      description,
      openGraph: {
        title,
        description,
        type: "article",
        siteName: "Colign",
      },
      twitter: {
        card: "summary",
        title,
        description,
      },
    };
  } catch {
    return defaultMeta;
  }
}

export default function ChangeDetailPage() {
  return <ChangeDetailClient />;
}
