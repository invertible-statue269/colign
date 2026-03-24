"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { aiConfigClient } from "@/lib/aiconfig";
import { useI18n } from "@/lib/i18n";
import { showError, showSuccess } from "@/lib/toast";

const MODELS_BY_PROVIDER: Record<string, string[]> = {
  openai: ["gpt-4o", "gpt-4o-mini"],
  anthropic: ["claude-sonnet-4-20250514", "claude-haiku-4-5-20251001"],
};

interface AIConfigSettingsProps {
  projectId: bigint;
}

export function AIConfigSettings({ projectId }: AIConfigSettingsProps) {
  const { t } = useI18n();

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [saved, setSaved] = useState(false);

  const [provider, setProvider] = useState("");
  const [model, setModel] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [apiKeyMasked, setApiKeyMasked] = useState("");
  const [includeProjectContext, setIncludeProjectContext] = useState(false);

  useEffect(() => {
    setLoading(true);
    aiConfigClient
      .getAIConfig({ projectId })
      .then((res) => {
        if (res.config) {
          setProvider(res.config.provider);
          setModel(res.config.model);
          setApiKeyMasked(res.config.apiKeyMasked);
          setIncludeProjectContext(res.config.includeProjectContext);
        }
      })
      .catch((err: unknown) => {
        showError(t("toast.loadFailed"), err);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [projectId]);

  function handleProviderChange(value: string | null) {
    if (!value) return;
    setProvider(value);
    setModel("");
  }

  const availableModels = provider ? (MODELS_BY_PROVIDER[provider] ?? []) : [];

  async function handleTestConnection() {
    if (!provider || !model) return;
    setTesting(true);
    try {
      const res = await aiConfigClient.testConnection({
        provider,
        model,
        apiKey,
      });
      if (res.success) {
        showSuccess(t("aiConfig.testSuccess"));
      } else {
        showError(t("aiConfig.testFailed"), res.error ? new Error(res.error) : undefined);
      }
    } catch (err: unknown) {
      showError(t("aiConfig.testFailed"), err);
    } finally {
      setTesting(false);
    }
  }

  async function handleSave() {
    if (!provider || !model) return;
    setSaving(true);
    try {
      const res = await aiConfigClient.saveAIConfig({
        projectId,
        provider,
        model,
        apiKey,
        includeProjectContext,
      });
      if (res.config) {
        setApiKeyMasked(res.config.apiKeyMasked);
        setApiKey("");
      }
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
      showSuccess(t("toast.saveSuccess"));
    } catch (err: unknown) {
      showError(t("toast.saveFailed"), err);
    } finally {
      setSaving(false);
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
    <Card className="border-border/50">
      <CardHeader>
        <CardTitle>{t("aiConfig.title")}</CardTitle>
        <CardDescription>{t("aiConfig.notConfigured")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-5">
        {/* Provider */}
        <div className="space-y-2">
          <Label>{t("aiConfig.provider")}</Label>
          <Select value={provider} onValueChange={handleProviderChange}>
            <SelectTrigger className="cursor-pointer">
              <SelectValue placeholder={t("aiConfig.provider")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="openai" className="cursor-pointer">
                OpenAI
              </SelectItem>
              <SelectItem value="anthropic" className="cursor-pointer">
                Anthropic
              </SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Model */}
        <div className="space-y-2">
          <Label>{t("aiConfig.model")}</Label>
          <Select value={model} onValueChange={(v) => v && setModel(v)} disabled={!provider}>
            <SelectTrigger className="cursor-pointer">
              <SelectValue placeholder={t("aiConfig.model")} />
            </SelectTrigger>
            <SelectContent>
              {availableModels.map((m) => (
                <SelectItem key={m} value={m} className="cursor-pointer">
                  {m}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* API Key */}
        <div className="space-y-2">
          <Label>{t("aiConfig.apiKey")}</Label>
          <Input
            type="password"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
            placeholder={apiKeyMasked || t("aiConfig.apiKeyPlaceholder")}
          />
        </div>

        {/* Include project context */}
        <div className="flex items-center gap-3">
          <Switch
            checked={includeProjectContext}
            onCheckedChange={setIncludeProjectContext}
          />
          <div>
            <p className="text-sm font-medium">{t("aiConfig.includeContext")}</p>
            <p className="text-xs text-muted-foreground">{t("aiConfig.includeContextHelp")}</p>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-3 pt-2">
          <Button
            variant="outline"
            onClick={handleTestConnection}
            disabled={testing || !provider || !model}
            className="cursor-pointer"
          >
            {testing ? t("common.loading") : t("aiConfig.testConnection")}
          </Button>
          <Button
            onClick={handleSave}
            disabled={saving || !provider || !model}
            className="cursor-pointer"
          >
            {saving ? t("common.saving") : t("aiConfig.save")}
          </Button>
          {saved && <span className="text-sm text-emerald-400">{t("common.saved")}</span>}
        </div>
      </CardContent>
    </Card>
  );
}
