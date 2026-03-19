"use client";

import { useEffect, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Header } from "@/components/layout/header";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { projectClient } from "@/lib/project";

const stageConfig: Record<string, { label: string; color: string; icon: string; glow: string }> = {
  draft: {
    label: "Draft",
    color: "bg-amber-500/10 text-amber-400 border-amber-500/20",
    icon: "M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10",
    glow: "shadow-amber-500/5",
  },
  design: {
    label: "Design",
    color: "bg-blue-500/10 text-blue-400 border-blue-500/20",
    icon: "M9.53 16.122a3 3 0 00-5.78 1.128 2.25 2.25 0 01-2.4 2.245 4.5 4.5 0 008.4-2.245c0-.399-.078-.78-.22-1.128zm0 0a15.998 15.998 0 003.388-1.62m-5.043-.025a15.994 15.994 0 011.622-3.395m3.42 3.42a15.995 15.995 0 004.764-4.648l3.876-5.814a1.151 1.151 0 00-1.597-1.597L14.146 6.32a15.996 15.996 0 00-4.649 4.763m3.42 3.42a6.776 6.776 0 00-3.42-3.42",
    glow: "shadow-blue-500/5",
  },
  review: {
    label: "Review",
    color: "bg-violet-500/10 text-violet-400 border-violet-500/20",
    icon: "M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178zM15 12a3 3 0 11-6 0 3 3 0 016 0z",
    glow: "shadow-violet-500/5",
  },
  ready: {
    label: "Ready",
    color: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
    icon: "M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z",
    glow: "shadow-emerald-500/5",
  },
};

interface Change {
  id: bigint;
  name: string;
  stage: string;
}

interface ProjectDetail {
  id: bigint;
  name: string;
  slug: string;
  description: string;
}

export default function ProjectDetailPage() {
  const params = useParams();
  const router = useRouter();
  const slug = params.slug as string;

  const [project, setProject] = useState<ProjectDetail | null>(null);
  const [changes, setChanges] = useState<Change[]>([]);
  const [newChangeName, setNewChangeName] = useState("");
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // Rename dialog
  const [renameOpen, setRenameOpen] = useState(false);
  const [renameName, setRenameName] = useState("");
  const [renameDesc, setRenameDesc] = useState("");
  const [renaming, setRenaming] = useState(false);

  // Delete dialog
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleteConfirm, setDeleteConfirm] = useState("");
  const [deleting, setDeleting] = useState(false);

  // Invite member dialog
  const [inviteOpen, setInviteOpen] = useState(false);
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState("member");
  const [inviting, setInviting] = useState(false);
  const [inviteSuccess, setInviteSuccess] = useState("");

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  useEffect(() => {
    async function load() {
      try {
        const projectRes = await projectClient.getProject({ slug });
        if (projectRes.project) {
          setProject({
            id: projectRes.project.id,
            name: projectRes.project.name,
            slug: projectRes.project.slug,
            description: projectRes.project.description,
          });

          const changesRes = await projectClient.listChanges({
            projectId: projectRes.project.id,
          });
          setChanges(
            changesRes.changes.map((c) => ({
              id: c.id,
              name: c.name,
              stage: c.stage,
            })),
          );
        }
      } catch {
        // handle error
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [slug]);

  async function handleCreateChange(e: React.FormEvent) {
    e.preventDefault();
    if (!newChangeName.trim() || !project) return;
    setCreating(true);
    try {
      const res = await projectClient.createChange({
        projectId: project.id,
        name: newChangeName.trim(),
      });
      if (res.change) {
        setChanges((prev) => [
          { id: res.change!.id, name: res.change!.name, stage: res.change!.stage },
          ...prev,
        ]);
        setNewChangeName("");
      }
    } catch {
      // handle error
    } finally {
      setCreating(false);
    }
  }

  async function handleRename() {
    if (!project || !renameName.trim()) return;
    setRenaming(true);
    try {
      const res = await projectClient.updateProject({
        id: project.id,
        name: renameName.trim(),
        description: renameDesc,
      });
      if (res.project) {
        setProject({
          id: res.project.id,
          name: res.project.name,
          slug: res.project.slug,
          description: res.project.description,
        });
        setRenameOpen(false);
        if (res.project.slug !== slug) {
          router.replace(`/projects/${res.project.slug}`);
        }
      }
    } catch {
      // handle error
    } finally {
      setRenaming(false);
    }
  }

  async function handleDelete() {
    if (!project) return;
    setDeleting(true);
    try {
      await projectClient.deleteProject({ id: project.id });
      router.push("/projects");
    } catch {
      setDeleting(false);
    }
  }

  async function handleInvite(e: React.FormEvent) {
    e.preventDefault();
    if (!project || !inviteEmail.trim()) return;
    setInviting(true);
    try {
      await projectClient.inviteMember({
        projectId: project.id,
        email: inviteEmail.trim(),
        role: inviteRole,
      });
      setInviteSuccess(inviteEmail.trim());
      setInviteEmail("");
      setTimeout(() => setInviteSuccess(""), 3000);
    } catch {
      // handle error
    } finally {
      setInviting(false);
    }
  }

  const stageCounts = changes.reduce(
    (acc, c) => {
      const stage = c.stage || "draft";
      acc[stage] = (acc[stage] || 0) + 1;
      return acc;
    },
    {} as Record<string, number>,
  );

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  }

  if (!project) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-muted-foreground">Project not found</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <Header breadcrumbs={[{ label: project.name }]} />

      <main className="mx-auto max-w-5xl px-6 pt-8 pb-16">
        {/* Project Hero */}
        <div className="mb-8">
          <div className="flex items-start justify-between">
            <div>
              <h1 className="text-2xl font-bold tracking-tight">{project.name}</h1>
              {project.description && (
                <p className="mt-1.5 text-sm text-muted-foreground">{project.description}</p>
              )}
            </div>

            {/* Project Menu */}
            <div className="relative" ref={menuRef}>
              <button
                onClick={() => setMenuOpen(!menuOpen)}
                className="cursor-pointer flex h-8 w-8 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
              >
                <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6.75 12a.75.75 0 11-1.5 0 .75.75 0 011.5 0zM12.75 12a.75.75 0 11-1.5 0 .75.75 0 011.5 0zM18.75 12a.75.75 0 11-1.5 0 .75.75 0 011.5 0z"
                  />
                </svg>
              </button>
              {menuOpen && (
                <div className="absolute right-0 top-10 z-50 w-48 rounded-xl border border-border/50 bg-popover p-1.5 shadow-xl animate-in fade-in slide-in-from-top-2 duration-150">
                  <button
                    onClick={() => {
                      setMenuOpen(false);
                      setRenameName(project.name);
                      setRenameDesc(project.description);
                      setRenameOpen(true);
                    }}
                    className="flex w-full cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-foreground transition-colors hover:bg-accent"
                  >
                    <svg
                      className="h-3.5 w-3.5 text-muted-foreground"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={1.5}
                        d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931z"
                      />
                    </svg>
                    Edit project
                  </button>
                  <button
                    onClick={() => {
                      setMenuOpen(false);
                      setInviteOpen(true);
                    }}
                    className="flex w-full cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-foreground transition-colors hover:bg-accent"
                  >
                    <svg
                      className="h-3.5 w-3.5 text-muted-foreground"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={1.5}
                        d="M19 7.5v3m0 0v3m0-3h3m-3 0h-3m-2.25-4.125a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zM4 19.235v-.11a6.375 6.375 0 0112.75 0v.109A12.318 12.318 0 0110.374 21c-2.331 0-4.512-.645-6.374-1.766z"
                      />
                    </svg>
                    Invite member
                  </button>
                  <div className="my-1.5 border-t border-border/50" />
                  <button
                    onClick={() => {
                      setMenuOpen(false);
                      setDeleteConfirm("");
                      setDeleteOpen(true);
                    }}
                    className="flex w-full cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-destructive transition-colors hover:bg-destructive/10"
                  >
                    <svg
                      className="h-3.5 w-3.5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={1.5}
                        d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0"
                      />
                    </svg>
                    Delete project
                  </button>
                </div>
              )}
            </div>
          </div>

          {/* Stage Stats */}
          {changes.length > 0 && (
            <div className="mt-6 grid grid-cols-4 gap-3">
              {(["draft", "design", "review", "ready"] as const).map((stage) => {
                const config = stageConfig[stage];
                const count = stageCounts[stage] || 0;
                return (
                  <div
                    key={stage}
                    className={`rounded-xl border border-border/40 bg-card/50 p-3.5 transition-all duration-200 ${count > 0 ? `shadow-lg ${config.glow}` : "opacity-50"}`}
                  >
                    <div className="flex items-center gap-2">
                      <svg
                        className={`h-3.5 w-3.5 ${config.color.split(" ")[1]}`}
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={1.5}
                          d={config.icon}
                        />
                      </svg>
                      <span className="text-xs font-medium text-muted-foreground">
                        {config.label}
                      </span>
                    </div>
                    <p className="mt-1.5 text-xl font-semibold tabular-nums">{count}</p>
                  </div>
                );
              })}
            </div>
          )}
        </div>

        {/* Create Change */}
        <form onSubmit={handleCreateChange} className="mb-8">
          <div className="flex gap-2">
            <div className="relative flex-1">
              <svg
                className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground/50"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1.5}
                  d="M12 4.5v15m7.5-7.5h-15"
                />
              </svg>
              <Input
                value={newChangeName}
                onChange={(e) => setNewChangeName(e.target.value)}
                placeholder="New change name (e.g., add-user-auth)"
                className="pl-9 bg-card/50 border-border/40 focus:border-primary/50 transition-colors"
              />
            </div>
            <Button
              type="submit"
              disabled={creating || !newChangeName.trim()}
              className="cursor-pointer px-5 transition-all duration-200"
            >
              {creating ? (
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-primary-foreground border-t-transparent" />
              ) : (
                "Create"
              )}
            </Button>
          </div>
        </form>

        {/* Changes List */}
        {changes.length === 0 ? (
          <div className="flex flex-col items-center justify-center rounded-2xl border border-dashed border-border/40 bg-card/30 py-20">
            <div className="rounded-2xl bg-primary/5 p-5 mb-5">
              <svg
                className="h-10 w-10 text-primary/40"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1}
                  d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m3.75 9v6m3-3H9m1.5-12H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z"
                />
              </svg>
            </div>
            <p className="text-sm font-medium text-foreground/70">No changes yet</p>
            <p className="mt-1 text-xs text-muted-foreground">
              Create your first change to start the SDD workflow
            </p>
          </div>
        ) : (
          <div className="space-y-2">
            {changes.map((change) => {
              const config = stageConfig[change.stage] ?? stageConfig.draft;
              return (
                <Link
                  key={String(change.id)}
                  href={`/projects/${project.slug}/changes/${change.id}`}
                >
                  <div
                    className={`group flex items-center justify-between rounded-xl border border-border/40 bg-card/50 px-5 py-4 transition-all duration-200 hover:border-primary/20 hover:bg-card/80 hover:shadow-lg ${config.glow} cursor-pointer`}
                  >
                    <div className="flex items-center gap-4">
                      <div
                        className={`flex h-8 w-8 items-center justify-center rounded-lg border ${config.color}`}
                      >
                        <svg
                          className="h-3.5 w-3.5"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={1.5}
                            d={config.icon}
                          />
                        </svg>
                      </div>
                      <div>
                        <p className="text-sm font-medium group-hover:text-foreground transition-colors">
                          {change.name}
                        </p>
                        <span
                          className={`inline-flex items-center text-xs font-medium mt-0.5 ${config.color.split(" ")[1]}`}
                        >
                          {config.label}
                        </span>
                      </div>
                    </div>
                    <svg
                      className="h-4 w-4 text-muted-foreground/30 transition-all duration-200 group-hover:text-muted-foreground group-hover:translate-x-0.5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={1.5}
                        d="M8.25 4.5l7.5 7.5-7.5 7.5"
                      />
                    </svg>
                  </div>
                </Link>
              );
            })}
          </div>
        )}
      </main>

      {/* Rename Dialog */}
      <Dialog open={renameOpen} onOpenChange={setRenameOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Edit project</DialogTitle>
            <DialogDescription>Update your project name and description.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label htmlFor="rename-name">Name</Label>
              <Input
                id="rename-name"
                value={renameName}
                onChange={(e) => setRenameName(e.target.value)}
                placeholder="Project name"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="rename-desc">Description</Label>
              <Input
                id="rename-desc"
                value={renameDesc}
                onChange={(e) => setRenameDesc(e.target.value)}
                placeholder="Optional description"
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              onClick={handleRename}
              disabled={renaming || !renameName.trim()}
              className="cursor-pointer"
            >
              {renaming ? "Saving..." : "Save changes"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Dialog */}
      <Dialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="text-destructive">Delete project</DialogTitle>
            <DialogDescription>
              This action cannot be undone. All changes, documents, and data in{" "}
              <span className="font-medium text-foreground">{project.name}</span> will be
              permanently deleted.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2 py-2">
            <Label htmlFor="delete-confirm">
              Type <span className="font-mono font-medium text-foreground">{project.name}</span> to
              confirm
            </Label>
            <Input
              id="delete-confirm"
              value={deleteConfirm}
              onChange={(e) => setDeleteConfirm(e.target.value)}
              placeholder={project.name}
            />
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeleteOpen(false)}
              className="cursor-pointer"
            >
              Cancel
            </Button>
            <Button
              onClick={handleDelete}
              disabled={deleting || deleteConfirm !== project.name}
              className="cursor-pointer bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {deleting ? "Deleting..." : "Delete project"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Invite Member Dialog */}
      <Dialog open={inviteOpen} onOpenChange={setInviteOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Invite member</DialogTitle>
            <DialogDescription>
              Send an invitation to collaborate on this project.
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleInvite}>
            <div className="space-y-4 py-2">
              <div className="space-y-2">
                <Label htmlFor="invite-email">Email address</Label>
                <Input
                  id="invite-email"
                  type="email"
                  value={inviteEmail}
                  onChange={(e) => setInviteEmail(e.target.value)}
                  placeholder="colleague@company.com"
                  required
                />
              </div>
              <div className="space-y-2">
                <Label>Role</Label>
                <div className="flex gap-2">
                  {["member", "admin"].map((role) => (
                    <button
                      key={role}
                      type="button"
                      onClick={() => setInviteRole(role)}
                      className={`cursor-pointer rounded-lg border px-3 py-1.5 text-sm transition-colors ${
                        inviteRole === role
                          ? "border-primary bg-primary/10 text-primary"
                          : "border-border/50 text-muted-foreground hover:border-border hover:text-foreground"
                      }`}
                    >
                      {role.charAt(0).toUpperCase() + role.slice(1)}
                    </button>
                  ))}
                </div>
              </div>
              {inviteSuccess && (
                <p className="text-sm text-emerald-400">Invitation sent to {inviteSuccess}</p>
              )}
            </div>
            <DialogFooter>
              <Button
                type="submit"
                disabled={inviting || !inviteEmail.trim()}
                className="cursor-pointer"
              >
                {inviting ? "Sending..." : "Send invitation"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
