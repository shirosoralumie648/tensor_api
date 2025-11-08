import AppShell from "@/components/workspace/AppShell";
import SidebarNav from "@/components/workspace/SidebarNav";
import RightPanel from "@/components/workspace/RightPanel";
import ChatWrapper from "@/components/home/ChatWrapper";

export default function Workspace() {
  return (
    <AppShell left={<SidebarNav />} right={<RightPanel />}>
      <ChatWrapper />
    </AppShell>
  );
}
