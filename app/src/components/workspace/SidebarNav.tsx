import { cn } from "@/components/ui/lib/utils";
import { Button } from "@/components/ui/button";
import { useNavigate } from "react-router-dom";
import { useSelector } from "react-redux";
import { selectAdmin, selectAuthenticated } from "@/store/auth";
import {
  Home as HomeIcon,
  Settings as SettingsIcon,
  LayoutPanelLeft,
  MessageSquare,
  Zap,
  FileText,
  Share2,
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
  return (
    <Button
      variant="ghost"
      className={cn("w-full justify-start gap-2", className)}
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
      <div className="text-xs text-muted-foreground px-2 mb-1">Workspace</div>
      <NavItem to="/workspace" icon={MessageSquare} label="Chat" />
      <NavItem to="/generate" icon={Zap} label="Generation" />
      <NavItem to="/article" icon={FileText} label="Article" />
      <NavItem to="/share/demo" icon={Share2} label="Sharing" />
      <div className="pt-2">
        <div className="text-xs text-muted-foreground px-2 mb-1">General</div>
        <NavItem to={authed ? "/settings" : "/login"} icon={SettingsIcon} label="Settings" />
        {isAdmin && (
          <NavItem to="/admin" icon={Gauge} label="Admin" />
        )}
      </div>
    </div>
  );
}
