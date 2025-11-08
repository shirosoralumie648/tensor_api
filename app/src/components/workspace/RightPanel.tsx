import { useDispatch, useSelector } from "react-redux";
import {
  contextSelector,
  historySelector,
  senderSelector,
  maxTokensSelector,
  temperatureSelector,
  topPSelector,
  topKSelector,
  presencePenaltySelector,
  frequencyPenaltySelector,
  repetitionPenaltySelector,
  setContext,
  setHistory,
  setSender,
  setMaxTokens,
  setTemperature,
  setTopP,
  setTopK,
  setPresencePenalty,
  setFrequencyPenalty,
  setRepetitionPenalty,
  dialogSelector,
  showCapabilitiesSelector,
  enableToolsSelector,
  enableJsonSelector,
  parallelToolsSelector,
  setEnableTools,
  setEnableJson,
  setParallelTools,
} from "@/store/settings";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { Slider } from "@/components/ui/slider";
import {
  selectModel,
  selectSupportModels,
  selectWeb,
  toggleWeb,
  useConversationActions,
  openMask,
} from "@/store/chat";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useEffect, useMemo, useState, type ReactNode, type ChangeEvent } from "react";
import { getMemoryPerformance } from "@/utils/app";
import { Wand2 } from "lucide-react";
import type { Model } from "@/api/types.tsx";
import { Badge } from "@/components/ui/badge";
import { getProviderCapabilities, getFallbackProviderCapabilities, type ProviderCapabilitiesMap } from "@/api/providers";

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="mb-4">
      <div className="text-xs text-muted-foreground mb-2">{title}</div>
      <div className="space-y-3">{children}</div>
    </div>
  );
}

function Row({ label, control }: { label: string; control: ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <Label className="text-sm text-muted-foreground">{label}</Label>
      <div className="flex items-center gap-2">{control}</div>
    </div>
  );
}

export default function RightPanel() {
  const dispatch = useDispatch();

  const context = useSelector(contextSelector);
  const history = useSelector(historySelector);
  const sender = useSelector(senderSelector);

  const maxTokens = useSelector(maxTokensSelector);
  const temperature = useSelector(temperatureSelector);
  const topP = useSelector(topPSelector);
  const topK = useSelector(topKSelector);
  const presence = useSelector(presencePenaltySelector);
  const frequency = useSelector(frequencyPenaltySelector);
  const repetition = useSelector(repetitionPenaltySelector);

  const web = useSelector(selectWeb);
  const currentModel = useSelector(selectModel);
  const models = useSelector(selectSupportModels);
  const { selected } = useConversationActions();

  const [mem, setMem] = useState<number>(getMemoryPerformance());
  const [capabilities, setCapabilities] = useState<ProviderCapabilitiesMap>({});
  useEffect(() => {
    getProviderCapabilities()
      .then((data) => setCapabilities(data))
      .catch(() => setCapabilities(getFallbackProviderCapabilities()));
  }, []);
  
  // provider-first grouping heuristic
  type ProviderMap = Record<string, ReturnType<typeof models.slice>>;
  const providerOrder = [
    "OpenAI",
    "Anthropic",
    "Google",
    "DeepSeek",
    "Alibaba Qwen",
    "Zhipu GLM",
    "Meta Llama",
    "Mistral",
    "Moonshot",
    "MiniMax",
    "Baidu ERNIE",
    "Tencent Hunyuan",
    "ByteDance Doubao",
    "Ollama",
    "Other",
  ];

  const getProvider = (id: string, tags?: string[]): string => {
    const s = id.toLowerCase();
    const tag = (tags || []).join(" ").toLowerCase();
    if (/^(gpt|o[0-9]|omni)/.test(s) || s.includes("openai") || tag.includes("openai")) return "OpenAI";
    if (s.includes("claude") || tag.includes("anthropic")) return "Anthropic";
    if (s.includes("gemini") || s.includes("palm") || tag.includes("google")) return "Google";
    if (s.includes("deepseek")) return "DeepSeek";
    if (s.startsWith("qwen") || tag.includes("qwen") || tag.includes("alibaba")) return "Alibaba Qwen";
    if (s.includes("glm") || s.includes("chatglm") || tag.includes("zhipu")) return "Zhipu GLM";
    if (s.includes("llama")) return "Meta Llama";
    if (s.includes("mistral")) return "Mistral";
    if (s.includes("moonshot")) return "Moonshot";
    if (s.includes("minimax")) return "MiniMax";
    if (s.includes("ernie") || s.includes("wenxin") || s.includes("qianfan") || tag.includes("baidu")) return "Baidu ERNIE";
    if (s.includes("hunyuan") || tag.includes("tencent")) return "Tencent Hunyuan";
    if (s.includes("doubao") || tag.includes("bytedance")) return "ByteDance Doubao";
    if (s.includes("ollama") || tag.includes("ollama")) return "Ollama";
    return "Other";
  };

  const providerMap = useMemo<ProviderMap>(() => {
    const map: ProviderMap = {} as any;
    models.forEach((m: Model) => {
      const p = getProvider(m.id, m.tag);
      (map[p] = map[p] || []).push(m);
    });
    // sort models by name within each provider
    Object.keys(map).forEach((k) => map[k].sort((a: Model, b: Model) => a.name.localeCompare(b.name)));
    return map;
  }, [models]);

  const providers = useMemo<string[]>(() => {
    const keys = Object.keys(providerMap);
    return keys.sort((a, b) => {
      const ia = providerOrder.indexOf(a);
      const ib = providerOrder.indexOf(b);
      return (ia === -1 ? 999 : ia) - (ib === -1 ? 999 : ib);
    });
  }, [providerMap]);

  const currentProvider = useMemo(() => {
    const current = models.find((m: Model) => m.id === currentModel);
    if (!current) return providers[0];
    return getProvider(current.id, current.tag);
  }, [currentModel, models, providers]);

  const [provider, setProvider] = useState<string>(currentProvider || providers[0]);
  // keep provider in sync when model changes externally
  const providerModels = providerMap[provider] || [];
  useEffect(() => {
    setProvider(currentProvider || providers[0]);
  }, [currentProvider, providers]);

  const featureSet = useMemo(() => {
    const cap = capabilities[provider]?.features;
    return cap || getFallbackProviderCapabilities()[provider]?.features;
  }, [capabilities, provider]);

  const showCapabilities = useSelector(showCapabilitiesSelector);
  const dialogOpen = useSelector(dialogSelector);
  const [capsOpen, setCapsOpen] = useState(false);
  const enableTools = useSelector(enableToolsSelector);
  const enableJson = useSelector(enableJsonSelector);
  const parallelTools = useSelector(parallelToolsSelector);

  return (
    <div className="rounded-lg border p-3 md:p-4 bg-card">
      <Section title="Conversation">
        <Row
          label="Keep context"
          control={<Switch checked={context} onCheckedChange={(v: boolean) => dispatch(setContext(v))} />}
        />
        <Row
          label={`History (${history})`}
          control={
            <Slider
              className="w-40"
              min={0}
              max={32}
              step={1}
              value={[history]}
              onValueChange={(v: number[]) => dispatch(setHistory(v[0] ?? 0))}
            />
          }
        />
        <Row
          label={"Send key: " + (sender ? "Enter" : "Ctrl + Enter")}
          control={<Switch checked={sender} onCheckedChange={(v: boolean) => dispatch(setSender(v))} />}
        />
        <Row
          label="Web search"
          control={<Switch checked={web} onCheckedChange={() => dispatch(toggleWeb())} />}
        />
      </Section>

      <Separator className="my-3" />

      <Section title="Model">
        <div className="flex items-center gap-2">
          <Select value={provider} onValueChange={(v: string) => setProvider(v)}>
            <SelectTrigger className="w-40">
              <SelectValue placeholder="Provider" />
            </SelectTrigger>
            <SelectContent>
              {providers.map((p: string) => (
                <SelectItem key={p} value={p}>{p}</SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Select value={currentModel} onValueChange={(v: string) => selected(v)}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="Select model" />
            </SelectTrigger>
            <SelectContent>
              {providerModels.map((m: Model) => (
                <SelectItem key={m.id} value={m.id}>
                  {m.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Button variant="outline" size="icon" onClick={() => dispatch(openMask())}>
            <Wand2 className="h-4 w-4" />
          </Button>
          <Button variant="outline" size="sm" onClick={() => setCapsOpen((v: boolean) => !v)}>
            Capabilities
          </Button>
        </div>
      </Section>

      {(showCapabilities || capsOpen || dialogOpen) && (
        <>
          <Separator className="my-3" />
          <Section title="Capabilities">
            <div className="flex flex-wrap gap-2">
              <Badge variant={featureSet?.text ? "secondary" : "outline"}>Text</Badge>
              <Badge variant={featureSet?.vision ? "secondary" : "outline"}>Vision</Badge>
              <Badge variant={featureSet?.tools ? "secondary" : "outline"}>Tools</Badge>
              <Badge variant={featureSet?.images ? "secondary" : "outline"}>Images</Badge>
              <Badge variant={featureSet?.audio ? "secondary" : "outline"}>Audio</Badge>
              <Badge variant={featureSet?.video ? "secondary" : "outline"}>Video</Badge>
              <Badge variant={featureSet?.embeddings ? "secondary" : "outline"}>Embeddings</Badge>
            </div>
            <div className="text-xs text-muted-foreground">
              Context: {featureSet?.context ? `${featureSet.context.toLocaleString()} tokens` : "—"}
            </div>
            {(featureSet?.tools || featureSet?.json || featureSet?.parallel_tools) && (
              <div className="space-y-2 pt-2">
                {featureSet?.tools && (
                  <Row
                    label="Tools"
                    control={
                      <Switch
                        checked={enableTools}
                        onCheckedChange={(v: boolean) => dispatch(setEnableTools(v))}
                      />
                    }
                  />
                )}
                {featureSet?.json && (
                  <Row
                    label="JSON mode"
                    control={
                      <Switch
                        checked={enableJson}
                        onCheckedChange={(v: boolean) => dispatch(setEnableJson(v))}
                      />
                    }
                  />
                )}
                {featureSet?.parallel_tools && (
                  <Row
                    label="Parallel tools"
                    control={
                      <Switch
                        checked={parallelTools}
                        onCheckedChange={(v: boolean) => dispatch(setParallelTools(v))}
                      />
                    }
                  />
                )}
              </div>
            )}
          </Section>
          <Separator className="my-3" />
        </>
      )}

      <Section title="Parameters">
        <Row
          label={`Max tokens (${maxTokens})`}
          control={
            <Input
              className="w-28"
              type="number"
              value={maxTokens}
              onChange={(e: ChangeEvent<HTMLInputElement>) => dispatch(setMaxTokens(Number(e.target.value) || 0))}
            />
          }
        />
        <Row
          label={`Temperature (${temperature.toFixed(2)})`}
          control={
            <Slider
              className="w-40"
              min={0}
              max={2}
              step={0.01}
              value={[temperature]}
              onValueChange={(v: number[]) => dispatch(setTemperature(v[0] ?? 0))}
            />
          }
        />
        <Row
          label={`Top P (${topP.toFixed(2)})`}
          control={
            <Slider
              className="w-40"
              min={0}
              max={1}
              step={0.01}
              value={[topP]}
              onValueChange={(v: number[]) => dispatch(setTopP(v[0] ?? 0))}
            />
          }
        />
        <Row
          label={`Top K (${topK})`}
          control={
            <Slider
              className="w-40"
              min={0}
              max={50}
              step={1}
              value={[topK]}
              onValueChange={(v: number[]) => dispatch(setTopK(v[0] ?? 0))}
            />
          }
        />
        <Row
          label={`Presence penalty (${presence.toFixed(2)})`}
          control={
            <Slider
              className="w-40"
              min={-2}
              max={2}
              step={0.01}
              value={[presence]}
              onValueChange={(v: number[]) => dispatch(setPresencePenalty(v[0] ?? 0))}
            />
          }
        />
        <Row
          label={`Frequency penalty (${frequency.toFixed(2)})`}
          control={
            <Slider
              className="w-40"
              min={-2}
              max={2}
              step={0.01}
              value={[frequency]}
              onValueChange={(v: number[]) => dispatch(setFrequencyPenalty(v[0] ?? 0))}
            />
          }
        />
        <Row
          label={`Repetition penalty (${repetition.toFixed(2)})`}
          control={
            <Slider
              className="w-40"
              min={0}
              max={2}
              step={0.01}
              value={[repetition]}
              onValueChange={(v: number[]) => dispatch(setRepetitionPenalty(v[0] ?? 0))}
            />
          }
        />
      </Section>

      <Separator className="my-3" />

      <Section title="System">
        <div className="text-xs text-muted-foreground">Memory: {Number.isNaN(mem) ? "N/A" : `${mem.toFixed(1)} MB`}</div>
        <Button variant="outline" size="sm" onClick={() => setMem(getMemoryPerformance())}>
          Refresh
        </Button>
      </Section>
    </div>
  );
}
