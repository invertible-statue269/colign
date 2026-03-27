import type { Metadata } from "next";
import ProjectDetailClient from "./project-detail-client";

const apiUrl =
  process.env.API_INTERNAL_URL ?? process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

type Props = {
  params: Promise<{ slug: string }>;
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;

  const defaultMeta: Metadata = {
    title: "Colign - AI-powered Spec Writing",
    description:
      "Collaborative Spec-Driven Development platform where developers and non-developers write specs with AI.",
  };

  try {
    const res = await fetch(`${apiUrl}/api/og/projects/${slug}`, {
      next: { revalidate: 60 },
    });
    if (!res.ok) return defaultMeta;

    const data = await res.json();
    const title = `${data.projectName} | Colign`;
    const description = data.description || `${data.projectName} on Colign`;

    return {
      title,
      description,
      openGraph: {
        title,
        description,
        type: "website",
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

export default function ProjectDetailPage() {
  return <ProjectDetailClient />;
}
