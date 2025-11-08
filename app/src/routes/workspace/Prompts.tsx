import AppShell from "@/components/workspace/AppShell";
import SidebarNav from "@/components/workspace/SidebarNav";
import RightPanel from "@/components/workspace/RightPanel";

export default function Prompts() {
  return (
    <AppShell left={<SidebarNav />} right={<RightPanel />}>
      <div className="rounded-lg border p-6">
        <div className="text-xl font-semibold mb-2">Prompt 市场</div>
        <p className="text-sm text-muted-foreground">即将提供：模板库/片段库、变量化表单、一键插入与搜索分组。</p>
      </div>
    </AppShell>
  );
}
