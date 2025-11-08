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
import { useState } from "react";
import { getMemoryPerformance } from "@/utils/app";
import { Wand2 } from "lucide-react";

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="mb-4">
      <div className="text-xs text-muted-foreground mb-2">{title}</div>
      <div className="space-y-3">{children}</div>
    </div>
  );
}

function Row({ label, control }: { label: string; control: React.ReactNode }) {
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

  return (
    <div className="rounded-lg border p-3 md:p-4 bg-card">
      <Section title="Conversation">
        <Row
          label="Keep context"
          control={<Switch checked={context} onCheckedChange={(v) => dispatch(setContext(v))} />}
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
              onValueChange={(v) => dispatch(setHistory(v[0] ?? 0))}
            />
          }
        />
        <Row
          label={"Send key: " + (sender ? "Enter" : "Ctrl + Enter")}
          control={<Switch checked={sender} onCheckedChange={(v) => dispatch(setSender(v))} />}
        />
        <Row
          label="Web search"
          control={<Switch checked={web} onCheckedChange={() => dispatch(toggleWeb())} />}
        />
      </Section>

      <Separator className="my-3" />

      <Section title="Model">
        <div className="flex items-center gap-2">
          <Select value={currentModel} onValueChange={(v) => selected(v)}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="Select model" />
            </SelectTrigger>
            <SelectContent>
              {models.map((m) => (
                <SelectItem key={m.id} value={m.id}>
                  {m.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button variant="outline" size="icon" onClick={() => dispatch(openMask())}>
            <Wand2 className="h-4 w-4" />
          </Button>
        </div>
      </Section>

      <Separator className="my-3" />

      <Section title="Parameters">
        <Row
          label={`Max tokens (${maxTokens})`}
          control={
            <Input
              className="w-28"
              type="number"
              value={maxTokens}
              onChange={(e) => dispatch(setMaxTokens(Number(e.target.value) || 0))}
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
              onValueChange={(v) => dispatch(setTemperature(v[0] ?? 0))}
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
              onValueChange={(v) => dispatch(setTopP(v[0] ?? 0))}
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
              onValueChange={(v) => dispatch(setTopK(v[0] ?? 0))}
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
              onValueChange={(v) => dispatch(setPresencePenalty(v[0] ?? 0))}
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
              onValueChange={(v) => dispatch(setFrequencyPenalty(v[0] ?? 0))}
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
              onValueChange={(v) => dispatch(setRepetitionPenalty(v[0] ?? 0))}
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
