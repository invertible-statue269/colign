"use client";

import { createContext, useContext, useEffect, useState, useCallback, useMemo } from "react";
import { createClient } from "@connectrpc/connect";
import { AuthService } from "@/gen/proto/auth/v1/auth_pb";
import { AUTH_CHANGED_EVENT, getAccessToken, getTokenPayload } from "./auth";
import { transport } from "./connect";

interface UserProfile {
  name: string;
  email: string;
  avatarUrl: string;
}

interface ProfileContextValue {
  profile: UserProfile;
  updateProfile: (profile: UserProfile) => void;
}

const meClient = createClient(AuthService, transport);

const defaultProfile: UserProfile = { name: "", email: "", avatarUrl: "" };

const ProfileContext = createContext<ProfileContextValue>({
  profile: defaultProfile,
  updateProfile: () => {},
});

export function ProfileProvider({ children }: { children: React.ReactNode }) {
  const payload = typeof window !== "undefined" ? getTokenPayload() : null;
  const [profile, setProfile] = useState<UserProfile>({
    name: payload?.name || "",
    email: payload?.email || "",
    avatarUrl: "",
  });

  const loadProfile = useCallback(async () => {
    const token = getAccessToken();
    if (!token) {
      setProfile(defaultProfile);
      return;
    }

    try {
      const res = await meClient.me({});
      setProfile({
        name: res.name || "",
        email: res.email || "",
        avatarUrl: res.avatarUrl || "",
      });
    } catch {
      const fallback = getTokenPayload();
      setProfile({
        name: fallback?.name || "",
        email: fallback?.email || "",
        avatarUrl: "",
      });
    }
  }, []);

  useEffect(() => {
    loadProfile();
  }, [loadProfile]);

  useEffect(() => {
    const handleAuthChanged = () => {
      void loadProfile();
    };
    window.addEventListener(AUTH_CHANGED_EVENT, handleAuthChanged);
    return () => window.removeEventListener(AUTH_CHANGED_EVENT, handleAuthChanged);
  }, [loadProfile]);

  const updateProfile = useCallback((updated: UserProfile) => {
    setProfile(updated);
  }, []);

  const value = useMemo(() => ({ profile, updateProfile }), [profile, updateProfile]);

  return <ProfileContext.Provider value={value}>{children}</ProfileContext.Provider>;
}

export function useProfile() {
  return useContext(ProfileContext);
}
