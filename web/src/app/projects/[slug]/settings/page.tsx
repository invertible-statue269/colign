"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { workflowClient } from "@/lib/workflow";

const policyOptions = [
  { value: "owner_one", label: "Owner 1 approval" },
  { value: "editor_two", label: "2+ Editor approvals" },
  { value: "all", label: "All members approve" },
  { value: "auto_pass", label: "Auto-pass (no approval)" },
];

export default function ProjectSettingsPage() {
  const params = useParams();
  const slug = params.slug as string;

  const [policy, setPolicy] = useState("owner_one");
  const [minCount, setMinCount] = useState(1);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  async function handleSave() {
    setSaving(true);
    setSaved(false);
    try {
      await workflowClient.setApprovalPolicy({
        projectId: BigInt(1), // TODO: from context
        policy,
        minCount,
      });
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch {
      // handle error
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="min-h-screen">
      <header className="sticky top-0 z-30 border-b border-border/50 bg-background/80 backdrop-blur-md">
        <div className="mx-auto flex max-w-6xl items-center gap-3 px-6 py-4">
          <Link href="/projects" className="text-xl font-bold tracking-tight">
            Co<span className="text-primary">Spec</span>
          </Link>
          <svg className="h-4 w-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8.25 4.5l7.5 7.5-7.5 7.5" />
          </svg>
          <Link href={`/projects/${slug}`} className="text-sm text-muted-foreground hover:text-foreground transition-colors duration-200">
            Project
          </Link>
          <svg className="h-4 w-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8.25 4.5l7.5 7.5-7.5 7.5" />
          </svg>
          <span className="text-sm font-medium">Settings</span>
        </div>
      </header>

      <main className="mx-auto max-w-2xl px-6 py-10">
        <h1 className="mb-8 text-2xl font-semibold tracking-tight">Project Settings</h1>

        <Card className="border-border/50">
          <CardHeader>
            <CardTitle className="text-base">Approval Policy</CardTitle>
            <CardDescription>
              Configure how changes are approved before implementation
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-5">
            <div className="space-y-2">
              <Label>Policy Type</Label>
              <Select value={policy} onValueChange={setPolicy}>
                <SelectTrigger className="cursor-pointer">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {policyOptions.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value} className="cursor-pointer">
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {policy !== "auto_pass" && policy !== "all" && (
              <div className="space-y-2">
                <Label>Minimum Approvals</Label>
                <Input
                  type="number"
                  min={1}
                  max={10}
                  value={minCount}
                  onChange={(e) => setMinCount(Number(e.target.value))}
                />
              </div>
            )}

            <div className="flex items-center gap-3 pt-2">
              <Button onClick={handleSave} disabled={saving} className="cursor-pointer">
                {saving ? "Saving..." : "Save Policy"}
              </Button>
              {saved && (
                <span className="text-sm text-emerald-400">Saved</span>
              )}
            </div>
          </CardContent>
        </Card>
      </main>
    </div>
  );
}
