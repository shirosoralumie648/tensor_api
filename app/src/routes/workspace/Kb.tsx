import AppShell from "@/components/workspace/AppShell";
import SidebarNav from "@/components/workspace/SidebarNav";
import RightPanel from "@/components/workspace/RightPanel";

export default function Kb() {
  return (
    <AppShell left={<SidebarNav />} right={<RightPanel />}>
      <div className="rounded-lg border p-6">
        <div className="text-xl font-semibold mb-2">知识库</div>
        <p className="text-sm text-muted-foreground">即将提供：语料上传、分块/向量化、索引、权限控制与检索增强。</p>
      </div>
    </AppShell>
  );
}
