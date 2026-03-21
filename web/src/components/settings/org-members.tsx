"use client";

import { useState, useEffect, useCallback } from "react";
import { orgClient } from "@/lib/organization";
import { useOrg } from "@/lib/org-context";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { UserPlus, Trash2, Shield, User, Crown } from "lucide-react";

type Member = {
  id: bigint;
  userId: bigint;
  userName: string;
  userEmail: string;
  role: string;
};

const roleIcons: Record<string, typeof Crown> = {
  owner: Crown,
  admin: Shield,
  member: User,
};

const roleLabels: Record<string, string> = {
  owner: "Owner",
  admin: "Admin",
  member: "Member",
};

export function OrgMembers() {
  const { currentOrg } = useOrg();
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState("member");
  const [inviting, setInviting] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const fetchMembers = useCallback(async () => {
    try {
      const res = await orgClient.listMembers({});
      setMembers(
        res.members.map((m) => ({
          id: m.id,
          userId: m.userId,
          userName: m.userName,
          userEmail: m.userEmail,
          role: m.role,
        })),
      );
    } catch (err) {
      console.error("Failed to load members:", err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchMembers();
  }, [fetchMembers]);

  async function handleInvite(e: React.FormEvent) {
    e.preventDefault();
    if (!inviteEmail.trim()) return;
    setInviting(true);
    setError("");
    setSuccess("");
    try {
      await orgClient.inviteOrgMember({
        email: inviteEmail.trim(),
        role: inviteRole,
      });
      setSuccess(`${inviteEmail.trim()} invited successfully`);
      setInviteEmail("");
      fetchMembers();
      setTimeout(() => setSuccess(""), 3000);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to invite");
    } finally {
      setInviting(false);
    }
  }

  async function handleRemove(userId: bigint, name: string) {
    if (!confirm(`Remove ${name} from the organization?`)) return;
    try {
      await orgClient.removeOrgMember({ userId });
      fetchMembers();
    } catch (err) {
      console.error("Failed to remove member:", err);
    }
  }

  async function handleRoleChange(userId: bigint, role: string) {
    try {
      await orgClient.updateOrgMemberRole({ userId, role });
      fetchMembers();
    } catch (err) {
      console.error("Failed to update role:", err);
    }
  }

  if (loading) {
    return (
      <Card className="border-border/50">
        <CardContent className="flex items-center justify-center py-12">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <Card className="border-border/50">
        <CardHeader>
          <CardTitle>Organization Members</CardTitle>
          <CardDescription>
            Manage members of {currentOrg?.name ?? "your organization"}. Members can be added to
            projects.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Invite form */}
          <form onSubmit={handleInvite} className="flex gap-2">
            <Input
              type="email"
              placeholder="Email address"
              value={inviteEmail}
              onChange={(e) => setInviteEmail(e.target.value)}
              className="flex-1"
            />
            <select
              value={inviteRole}
              onChange={(e) => setInviteRole(e.target.value)}
              className="rounded-md border border-border bg-background px-3 py-2 text-sm"
            >
              <option value="member">Member</option>
              <option value="admin">Admin</option>
            </select>
            <Button
              type="submit"
              disabled={inviting || !inviteEmail.trim()}
              className="cursor-pointer"
            >
              <UserPlus className="mr-1.5 size-4" />
              {inviting ? "Inviting..." : "Invite"}
            </Button>
          </form>

          {error && <p className="text-sm text-destructive">{error}</p>}
          {success && <p className="text-sm text-emerald-400">{success}</p>}

          {/* Members list */}
          <div className="divide-y divide-border/50 rounded-lg border border-border/50">
            {members.map((member) => {
              const RoleIcon = roleIcons[member.role] ?? User;
              return (
                <div
                  key={String(member.id)}
                  className="flex items-center justify-between px-4 py-3"
                >
                  <div className="flex items-center gap-3">
                    <div className="flex size-8 items-center justify-center rounded-full bg-accent text-xs font-medium">
                      {member.userName?.[0]?.toUpperCase() ??
                        member.userEmail?.[0]?.toUpperCase() ??
                        "?"}
                    </div>
                    <div>
                      <p className="text-sm font-medium">{member.userName || "—"}</p>
                      <p className="text-xs text-muted-foreground">{member.userEmail}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {member.role === "owner" ? (
                      <span className="flex items-center gap-1 rounded-md bg-amber-500/10 px-2 py-1 text-xs text-amber-400">
                        <RoleIcon className="size-3" />
                        {roleLabels[member.role]}
                      </span>
                    ) : (
                      <>
                        <select
                          value={member.role}
                          onChange={(e) => handleRoleChange(member.userId, e.target.value)}
                          className="cursor-pointer rounded-md border border-border/50 bg-background px-2 py-1 text-xs"
                        >
                          <option value="member">Member</option>
                          <option value="admin">Admin</option>
                        </select>
                        <button
                          onClick={() =>
                            handleRemove(member.userId, member.userName || member.userEmail)
                          }
                          className="cursor-pointer rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                        >
                          <Trash2 className="size-3.5" />
                        </button>
                      </>
                    )}
                  </div>
                </div>
              );
            })}
            {members.length === 0 && (
              <div className="px-4 py-8 text-center text-sm text-muted-foreground">
                No members yet. Invite someone to get started.
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
