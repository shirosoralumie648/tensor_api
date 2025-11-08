import AppShell from "@/components/workspace/AppShell";
import SidebarNav from "@/components/workspace/SidebarNav";
import RightPanel from "@/components/workspace/RightPanel";

export default function Mcp() {
  return (
    <AppShell left={<SidebarNav />} right={<RightPanel />}>
      <div className="rounded-lg border p-6">
        <div className="text-xl font-semibold mb-2">MCP</div>
        <p className="text-sm text-muted-foreground">即将提供：端点注册、能力声明、健康检查与交互调试台。</p>
      </div>
    </AppShell>
  );
}
