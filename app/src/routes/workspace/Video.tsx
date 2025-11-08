import AppShell from "@/components/workspace/AppShell";
import SidebarNav from "@/components/workspace/SidebarNav";
import RightPanel from "@/components/workspace/RightPanel";

export default function Video() {
  return (
    <AppShell left={<SidebarNav />} right={<RightPanel />}>
      <div className="rounded-lg border p-6">
        <div className="text-xl font-semibold mb-2">视频生成</div>
        <p className="text-sm text-muted-foreground">即将提供：文生视频/图生视频、分辨率/时长控制、提示分镜、不同 Provider 的特性能力聚合。</p>
      </div>
    </AppShell>
  );
}
