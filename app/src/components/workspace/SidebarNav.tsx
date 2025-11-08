import { cn } from "@/components/ui/lib/utils";
import { Button } from "@/components/ui/button";
import { useLocation, useNavigate } from "react-router-dom";
import { useSelector } from "react-redux";
import { selectAdmin, selectAuthenticated } from "@/store/auth";
import {
  MessageSquare,
  Image as ImageIcon,
  AudioLines,
  Video as VideoIcon,
  BookOpen,
  CircuitBoard,
  Bot,
  Store,
  Settings as SettingsIcon,
  Gauge,
} from "lucide-react";

function NavItem({
  to,
  icon: Icon,
  label,
  className,
}: {
  to: string;
  icon: any;
  label: string;
  className?: string;
}) {
  const nav = useNavigate();
  const { pathname } = useLocation();
  const active = pathname === to;
  return (
    <Button
      variant="ghost"
      className={cn("w-full justify-start gap-2", active && "brand-border rounded-md", className)}
      onClick={() => nav(to)}
    >
      <Icon className="h-4 w-4" />
      <span>{label}</span>
    </Button>
  );
}

export default function SidebarNav() {
  const isAdmin = useSelector(selectAdmin);
  const authed = useSelector(selectAuthenticated);

  return (
    <div className="space-y-1">
      <div className="text-xs text-muted-foreground px-2 mb-1">功能</div>
      <NavItem to="/workspace" icon={MessageSquare} label="对话" />
      <NavItem to="/generate" icon={ImageIcon} label="绘图" />
      <NavItem to="/workspace/tts" icon={AudioLines} label="语音生成" />
      <NavItem to="/workspace/video" icon={VideoIcon} label="视频生成" />
      <NavItem to="/workspace/kb" icon={BookOpen} label="知识库" />
      <NavItem to="/workspace/mcp" icon={CircuitBoard} label="MCP" />
      <NavItem to="/workspace/agents" icon={Bot} label="智能体" />
      <NavItem to="/workspace/prompts" icon={Store} label="Prompt 市场" />
      <div className="pt-2">
        <div className="text-xs text-muted-foreground px-2 mb-1">系统</div>
        <NavItem to={authed ? "/settings" : "/login"} icon={SettingsIcon} label="设置" />
        {isAdmin && <NavItem to="/admin" icon={Gauge} label="管理后台" />}
      </div>
    </div>
  );
}
