import AppShell from "@/components/workspace/AppShell";
import SidebarNav from "@/components/workspace/SidebarNav";
import RightPanel from "@/components/workspace/RightPanel";

export default function Agents() {
  return (
    <AppShell left={<SidebarNav />} right={<RightPanel />}>
      <div className="rounded-lg border p-6">
        <div className="text-xl font-semibold mb-2">智能体</div>
        <p className="text-sm text-muted-foreground">即将提供：角色/目标/工具配置，可视化测试与模板化管理。</p>
      </div>
    </AppShell>
  );
}
