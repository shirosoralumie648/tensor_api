import {
  createBrowserRouter,
  RouterProvider,
  useLocation,
  useNavigate,
} from "react-router-dom";
import Home from "./routes/Home.tsx";
import NotFound from "./routes/NotFound.tsx";
import Auth from "./routes/Auth.tsx";
import React, { Suspense, useEffect } from "react";
import { useDeeptrain } from "@/conf/env.ts";
import Landing from "./routes/Landing.tsx";
import OAuthCallback from "./routes/OAuthCallback.tsx";
import Settings from "./routes/Settings.tsx";
import Register from "@/routes/Register.tsx";
import Forgot from "@/routes/Forgot.tsx";
import { lazyFactor } from "@/utils/loader.tsx";
import { useSelector } from "react-redux";
import { selectAdmin, selectAuthenticated, selectInit } from "@/store/auth.ts";

const Generation = lazyFactor(() => import("@/routes/Generation.tsx"));
const Sharing = lazyFactor(() => import("@/routes/Sharing.tsx"));
const Article = lazyFactor(() => import("@/routes/Article.tsx"));

const Admin = lazyFactor(() => import("@/routes/Admin.tsx"));
const Dashboard = lazyFactor(() => import("@/routes/admin/DashBoard.tsx"));
const Market = lazyFactor(() => import("@/routes/admin/Market.tsx"));
const Channel = lazyFactor(() => import("@/routes/admin/Channel.tsx"));
const System = lazyFactor(() => import("@/routes/admin/System.tsx"));
const Charge = lazyFactor(() => import("@/routes/admin/Charge.tsx"));
const Users = lazyFactor(() => import("@/routes/admin/Users.tsx"));
const Broadcast = lazyFactor(() => import("@/routes/admin/Broadcast.tsx"));
const Subscription = lazyFactor(
  () => import("@/routes/admin/Subscription.tsx"),
);
const Logger = lazyFactor(() => import("@/routes/admin/Logger.tsx"));
const Feature = lazyFactor(() => import("@/routes/admin/Feature.tsx"));
const Payment = lazyFactor(() => import("@/routes/admin/Payment.tsx"));
const PaymentOrders = lazyFactor(
  () => import("@/routes/admin/PaymentOrders.tsx"),
);
const Workspace = lazyFactor(() => import("@/routes/Workspace.tsx"));
const WorkspaceTts = lazyFactor(() => import("@/routes/workspace/Tts.tsx"));
const WorkspaceVideo = lazyFactor(() => import("@/routes/workspace/Video.tsx"));
const WorkspaceKb = lazyFactor(() => import("@/routes/workspace/Kb.tsx"));
const WorkspaceMcp = lazyFactor(() => import("@/routes/workspace/Mcp.tsx"));
const WorkspaceAgents = lazyFactor(() => import("@/routes/workspace/Agents.tsx"));
const WorkspacePrompts = lazyFactor(() => import("@/routes/workspace/Prompts.tsx"));

const router = createBrowserRouter(
  [
    {
      id: "home",
      path: "/",
      element: <Landing />,
      ErrorBoundary: NotFound,
    },
    {
      id: "settings",
      path: "/settings",
      element: (
        <AuthRequired>
          <Suspense>
            <Settings />
          </Suspense>
        </AuthRequired>
      ),
      ErrorBoundary: NotFound,
    },
    {
      id: "app",
      path: "/app",
      element: (
        <AuthRequired>
          <Home />
        </AuthRequired>
      ),
      ErrorBoundary: NotFound,
    },
    {
      id: "workspace",
      path: "/workspace",
      element: (
        <AuthRequired>
          <Suspense>
            <Workspace />
          </Suspense>
        </AuthRequired>
      ),
      ErrorBoundary: NotFound,
    },
    {
      id: "workspace-tts",
      path: "/workspace/tts",
      element: (
        <AuthRequired>
          <Suspense>
            <WorkspaceTts />
          </Suspense>
        </AuthRequired>
      ),
      ErrorBoundary: NotFound,
    },
    {
      id: "workspace-video",
      path: "/workspace/video",
      element: (
        <AuthRequired>
          <Suspense>
            <WorkspaceVideo />
          </Suspense>
        </AuthRequired>
      ),
      ErrorBoundary: NotFound,
    },
    {
      id: "workspace-kb",
      path: "/workspace/kb",
      element: (
        <AuthRequired>
          <Suspense>
            <WorkspaceKb />
          </Suspense>
        </AuthRequired>
      ),
      ErrorBoundary: NotFound,
    },
    {
      id: "workspace-mcp",
      path: "/workspace/mcp",
      element: (
        <AuthRequired>
          <Suspense>
            <WorkspaceMcp />
          </Suspense>
        </AuthRequired>
      ),
      ErrorBoundary: NotFound,
    },
    {
      id: "workspace-agents",
      path: "/workspace/agents",
      element: (
        <AuthRequired>
          <Suspense>
            <WorkspaceAgents />
          </Suspense>
        </AuthRequired>
      ),
      ErrorBoundary: NotFound,
    },
    {
      id: "workspace-prompts",
      path: "/workspace/prompts",
      element: (
        <AuthRequired>
          <Suspense>
            <WorkspacePrompts />
          </Suspense>
        </AuthRequired>
      ),
      ErrorBoundary: NotFound,
    },
    {
      id: "login",
      path: "/login",
      element: (
        <AuthForbidden>
          <Auth />
        </AuthForbidden>
      ),
      ErrorBoundary: NotFound,
    },
    {
      id: "oauth-callback",
      path: "/oauth/callback",
      element: <OAuthCallback />,
      ErrorBoundary: NotFound,
    },
    !useDeeptrain &&
      ({
        id: "register",
        path: "/register",
        element: (
          <AuthForbidden>
            <Register />
          </AuthForbidden>
        ),
        ErrorBoundary: NotFound,
      } as any),
    !useDeeptrain &&
      ({
        id: "forgot",
        path: "/forgot",
        element: (
          <AuthForbidden>
            <Forgot />
          </AuthForbidden>
        ),
        ErrorBoundary: NotFound,
      } as any),
    {
      id: "generation",
      path: "/generate",
      element: (
        <AuthRequired>
          <Suspense>
            <Generation />
          </Suspense>
        </AuthRequired>
      ),
      ErrorBoundary: NotFound,
    },
    {
      id: "share",
      path: "/share/:hash",
      element: (
        <Suspense>
          <Sharing />
        </Suspense>
      ),
      ErrorBoundary: NotFound,
    },
    {
      id: "article",
      path: "/article",
      element: (
        <AuthRequired>
          <Suspense>
            <Article />
          </Suspense>
        </AuthRequired>
      ),
      ErrorBoundary: NotFound,
    },
    {
      id: "admin",
      path: "/admin",
      element: (
        <AdminRequired>
          <Suspense>
            <Admin />
          </Suspense>
        </AdminRequired>
      ),
      children: [
        {
          id: "admin-dashboard",
          path: "",
          element: (
            <Suspense>
              <Dashboard />
            </Suspense>
          ),
        },
        {
          id: "admin-users",
          path: "users",
          element: (
            <Suspense>
              <Users />
            </Suspense>
          ),
        },
        {
          id: "admin-market",
          path: "market",
          element: (
            <Suspense>
              <Market />
            </Suspense>
          ),
        },
        {
          id: "admin-channel",
          path: "channel",
          element: (
            <Suspense>
              <Channel />
            </Suspense>
          ),
        },
        {
          id: "admin-system",
          path: "system",
          element: (
            <Suspense>
              <System />
            </Suspense>
          ),
        },
        {
          id: "admin-charge",
          path: "charge",
          element: (
            <Suspense>
              <Charge />
            </Suspense>
          ),
        },
        {
          id: "admin-broadcast",
          path: "broadcast",
          element: (
            <Suspense>
              <Broadcast />
            </Suspense>
          ),
        },
        {
          id: "admin-subscription",
          path: "subscription",
          element: (
            <Suspense>
              <Subscription />
            </Suspense>
          ),
        },
        {
          id: "admin-logger",
          path: "logger",
          element: (
            <Suspense>
              <Logger />
            </Suspense>
          ),
        },
        {
          id: "admin-payment",
          path: "payment",
          element: (
            <Suspense>
              <Payment />
            </Suspense>
          ),
        },
        {
          id: "admin-payment-orders",
          path: "payment/orders",
          element: (
            <Suspense>
              <PaymentOrders />
            </Suspense>
          ),
        },
        {
          id: "admin-feature",
          path: "feature",
          element: (
            <Suspense>
              <Feature />
            </Suspense>
          ),
        },
      ],
      ErrorBoundary: NotFound,
    },
    {
      id: "not-found",
      path: "*",
      element: <NotFound />,
    },
  ].filter(Boolean),
);

export function AuthRequired({ children }: { children: React.ReactNode }) {
  const init = useSelector(selectInit);
  const authenticated = useSelector(selectAuthenticated);
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    if (init && !authenticated) {
      navigate("/login", { state: { from: location.pathname } });
    }
  }, [init, authenticated]);

  return <>{children}</>;
}

export function AuthForbidden({ children }: { children: React.ReactNode }) {
  const init = useSelector(selectInit);
  const authenticated = useSelector(selectAuthenticated);
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    if (init && authenticated) {
      navigate("/app", { state: { from: location.pathname } });
    }
  }, [init, authenticated]);

  return <>{children}</>;
}

export function AdminRequired({ children }: { children: React.ReactNode }) {
  const init = useSelector(selectInit);
  const admin = useSelector(selectAdmin);
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    if (init && !admin) {
      navigate("/app", { state: { from: location.pathname } });
    }
  }, [init, admin]);

  return <>{children}</>;
}

export function AppRouter() {
  return <RouterProvider router={router} />;
}

export default router;
