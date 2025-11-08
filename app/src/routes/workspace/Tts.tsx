import AppShell from "@/components/workspace/AppShell";
import SidebarNav from "@/components/workspace/SidebarNav";
import RightPanel from "@/components/workspace/RightPanel";

export default function Tts() {
  return (
    <AppShell left={<SidebarNav />} right={<RightPanel />}>
      <div className="rounded-lg border p-6">
        <div className="text-xl font-semibold mb-2">语音生成（TTS）</div>
        <p className="text-sm text-muted-foreground">即将提供：文本转语音、多发音人、情感控制、格式（mp3/wav）与采样率，支持不同 Provider 的 TTS 能力。</p>
      </div>
    </AppShell>
  );
}
