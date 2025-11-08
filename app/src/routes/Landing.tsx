import { useEffect, useRef } from "react";
import { Button } from "@/components/ui/button.tsx";
import { Card, CardContent } from "@/components/ui/card.tsx";
import { Badge } from "@/components/ui/badge.tsx";
import { Rocket, Sparkles, Shield, Zap, Layers, MessageSquare } from "lucide-react";
import router from "@/router.tsx";
import { appLogo, appName, docsEndpoint } from "@/conf/env.ts";

function Feature({ icon, title, desc, className }: { icon: any; title: string; desc: string; className?: string }) {
  return (
    <Card className={`rounded-2xl border border-black/10 dark:border-white/10 bg-black/5 dark:bg-white/5 backdrop-blur-xl transform-gpu will-change-transform hover:-translate-y-0.5 hover:shadow-2xl hover:border-black/20 dark:hover:border-white/20 transition-all duration-300 ${className || ""}`}> 
      <CardContent className={`p-6`}> 
        <div className={`flex items-start gap-3`}>
          <div className={`p-2 rounded-md bg-gradient-to-br from-pink-500/20 to-sky-500/20 text-pink-400`}>{icon}</div>
          <div>
            <div className={`font-semibold text-base`}>{title}</div>
            <div className={`text-sm text-muted-foreground mt-1`}>{desc}</div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export default function Landing() {
  useEffect(() => {
    const elements = document.querySelectorAll<HTMLElement>(".reveal");
    const io = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) entry.target.classList.add("reveal-show");
        });
      },
      { threshold: 0.18 },
    );
    elements.forEach((el) => io.observe(el));
    return () => io.disconnect();
  }, []);

  const tiltRef = useRef<HTMLDivElement>(null);
  const handleShowcaseMove = (e: any) => {
    const el = tiltRef.current;
    if (!el) return;
    const rect = el.getBoundingClientRect();
    const x = (e.clientX - rect.left) / rect.width - 0.5;
    const y = (e.clientY - rect.top) / rect.height - 0.5;
    const rx = (-y * 6).toFixed(2);
    const ry = (x * 8).toFixed(2);
    el.style.transform = `perspective(1200px) rotateX(${rx}deg) rotateY(${ry}deg)`;
  };
  const resetShowcaseTilt = () => {
    const el = tiltRef.current;
    if (!el) return;
    el.style.transform = `perspective(1200px) rotateX(0deg) rotateY(0deg)`;
  };

  return (
    <div className={`min-h-[calc(100vh-64px)] w-full flex flex-col relative overflow-hidden`}> 
      <div className={`pointer-events-none absolute inset-0 -z-10 [mask-image:radial-gradient(ellipse_at_center,black,transparent_70%)]`}>
        <div className={`absolute -top-24 -left-24 h-72 w-72 rounded-full bg-gradient-to-tr from-pink-500 via-fuchsia-500 to-sky-500 opacity-30 blur-3xl`}></div>
        <div className={`absolute -bottom-32 -right-24 h-96 w-96 rounded-full bg-gradient-to-tr from-sky-500 via-fuchsia-500 to-pink-500 opacity-30 blur-3xl`}></div>
      </div>
      <section className={`w-full pt-24 pb-16 flex flex-col items-center text-center`}> 
        <img src={appLogo} alt={appName} className={`w-16 h-16 mb-5 drop-shadow-lg reveal`} />
        <h1
          className={`text-4xl sm:text-6xl font-extrabold tracking-tight mb-3 bg-gradient-to-r from-pink-500 via-fuchsia-400 to-sky-500 bg-clip-text text-transparent reveal`}
          style={{ backgroundSize: "200% 200%", animation: "gradientShift 14s ease infinite" }}
        >
          {appName}
        </h1>
        <p className={`max-w-2xl leading-6 px-4 text-slate-600 dark:text-white/70`}>
          面向多模型、多场景的下一代聊天与知识助手平台。集成 30+ 模型提供商，内置管理后台与 API 中转，支持实时对话、图像、多模态、订阅与配额。
        </p>
        <div className={`flex gap-3 mt-6 reveal`}> 
          <Button
            size={`lg`}
            className={`bg-gradient-to-r from-pink-500 via-fuchsia-500 to-sky-500 text-white shadow-lg hover:opacity-90 hover:shadow-pink-500/30 border-0`}
            style={{ backgroundSize: "200% 200%", animation: "gradientShift 10s ease infinite" }}
            onClick={() => router.navigate("/login")}
          >
            立即开始
          </Button>
          <Button
            size={`lg`}
            variant={`outline`}
            className={`border-black/10 dark:border-white/20 text-slate-900 dark:text-white/90 hover:bg-black/5 dark:hover:bg-white/10`}
            onClick={() => (location.href = docsEndpoint)}
          >
            查看文档
          </Button>
        </div>
        <div className={`mt-4 reveal`}> 
          <Badge variant={`secondary`} className={`px-3 py-1 bg-black/5 dark:bg-white/10 text-slate-700 dark:text-white/80 backdrop-blur`}>开源 · Golang + React + Tauri</Badge>
        </div>
      </section>

      {/* Showcase */}
      <section className={`w-full px-4 sm:px-8 max-w-6xl mx-auto pb-12 reveal`}>
        <div className={`p-[1px] rounded-3xl bg-gradient-to-r from-pink-500 via-fuchsia-500 to-sky-500 bg-[length:200%_200%] animate-[gradientShift_12s_linear_infinite] shadow-[0_0_40px_-10px_rgba(236,72,153,.6)]`}>
          <div
            ref={tiltRef}
            onMouseMove={handleShowcaseMove}
            onMouseLeave={resetShowcaseTilt}
            className={`rounded-[22px] relative overflow-hidden backdrop-blur-xl bg-black/5 dark:bg-white/5 border border-black/10 dark:border-white/10 transform-gpu will-change-transform transition-transform duration-300`}
          >
            <div className={`aspect-[16/9] w-full`}> 
              <div className={`absolute -top-20 -left-10 h-64 w-64 rounded-full bg-gradient-to-tr from-pink-400/40 to-sky-400/40 blur-2xl animate-[floaty_14s_ease-in-out_infinite]`} />
              <div className={`absolute -bottom-24 -right-16 h-72 w-72 rounded-full bg-gradient-to-tr from-sky-400/40 to-pink-400/40 blur-2xl animate-[floaty_18s_ease-in-out_infinite]`} />
              <div className={`relative z-10 h-full w-full p-6 sm:p-10`}>
                <div className={`h-4 w-24 rounded-full bg-black/10 dark:bg-white/20 mb-6`} />
                <div className={`space-y-3`}>
                  <div className={`h-3 w-3/5 rounded bg-black/10 dark:bg-white/15`} />
                  <div className={`h-3 w-2/5 rounded bg-black/5 dark:bg-white/10`} />
                  <div className={`h-3 w-4/5 rounded bg-black/5 dark:bg-white/10`} />
                </div>
                <div className={`grid grid-cols-1 sm:grid-cols-3 gap-4 mt-8`}>
                  <div className={`h-24 rounded-xl bg-black/5 dark:bg-white/10 backdrop-blur`} />
                  <div className={`h-24 rounded-xl bg-black/5 dark:bg-white/10 backdrop-blur`} />
                  <div className={`h-24 rounded-xl bg-black/5 dark:bg-white/10 backdrop-blur`} />
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className={`w-full px-4 sm:px-8 max-w-6xl mx-auto pb-16`}> 
        <div className={`grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4`}> 
          <Feature className={`reveal`} icon={<Sparkles className={`w-5 h-5`} />} title={`多模型聚合`} desc={`支持 OpenAI、Claude、DeepSeek、Ollama 等，自动适配消息格式与流式返回。`} />
          <Feature className={`reveal`} icon={<Shield className={`w-5 h-5`} />} title={`权限与风控`} desc={`管理员后台、通道管理、用量限制、风控与审计。`} />
          <Feature className={`reveal`} icon={<Zap className={`w-5 h-5`} />} title={`实时与多模态`} desc={`文本、图像生成与解析，支持 WebSocket 流式对话。`} />
          <Feature className={`reveal`} icon={<Layers className={`w-5 h-5`} />} title={`插件与扩展`} desc={`可扩展的适配器与中间件体系，便于接入第三方服务。`} />
          <Feature className={`reveal`} icon={<MessageSquare className={`w-5 h-5`} />} title={`UI/UX 体验`} desc={`现代化前端，暗黑模式与移动端友好设计。`} />
          <Feature className={`reveal`} icon={<Rocket className={`w-5 h-5`} />} title={`一键部署`} desc={`提供 Docker Compose 与单镜像部署，快速上线。`} />
        </div>
      </section>

      {/* Use Cases */}
      <section className={`w-full px-4 sm:px-8 max-w-6xl mx-auto pb-10`}>
        <div className={`text-center mb-6 reveal`}>
          <h2 className={`text-2xl font-bold bg-gradient-to-r from-pink-500 via-fuchsia-400 to-sky-500 bg-clip-text text-transparent`} style={{ backgroundSize: "200% 200%", animation: "gradientShift 16s ease infinite" }}>典型场景</h2>
          <p className={`text-slate-600 dark:text-white/70 mt-1`}>客服助手 · 编码助手 · 知识检索 · 图像问答</p>
        </div>
        <div className={`grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4`}>
          {["客服助手", "编码助手", "知识检索", "图像问答"].map((t, i) => (
            <div key={i} className={`reveal rounded-2xl p-[1px] bg-gradient-to-r from-pink-500/60 via-fuchsia-500/60 to-sky-500/60`}>
              <div className={`rounded-[18px] bg-black/5 dark:bg-white/5 backdrop-blur-xl p-5 text-center`}>
                <div className={`text-lg font-semibold`}>{t}</div>
                <div className={`text-sm text-slate-600 dark:text-white/70 mt-1`}>一键集成 · 即刻可用</div>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* Steps */}
      <section className={`w-full px-4 sm:px-8 max-w-6xl mx-auto pb-16`}>
        <div className={`text-center mb-6 reveal`}>
          <h2 className={`text-2xl font-bold`}>3 步上手</h2>
        </div>
        <div className={`grid grid-cols-1 sm:grid-cols-3 gap-4`}>
          {[
            { n: 1, t: "部署", d: "Docker 一键部署，开箱即用" },
            { n: 2, t: "配置", d: "在后台接入模型与通道" },
            { n: 3, t: "使用", d: "聊天、API中转、模型调用" },
          ].map((s, i) => (
            <div key={i} className={`reveal rounded-2xl border border-black/10 dark:border-white/10 bg-black/5 dark:bg-white/5 backdrop-blur-xl p-6 flex flex-col items-center`}>
              <div className={`w-10 h-10 rounded-full bg-gradient-to-r from-pink-500 to-sky-500 text-white flex items-center justify-center font-bold mb-3`}>
                {s.n}
              </div>
              <div className={`font-semibold`}>{s.t}</div>
              <div className={`text-sm text-slate-600 dark:text-white/70 mt-1 text-center`}>{s.d}</div>
            </div>
          ))}
        </div>
      </section>

      {/* Footer */}
      <footer className={`w-full px-4 sm:px-8 max-w-6xl mx-auto pb-10 text-center text-slate-500 dark:text-white/60`}> 
        <div className={`text-sm`}>© {new Date().getFullYear()} {appName} · <a className={`underline hover:text-black dark:hover:text-white`} href={docsEndpoint} target={`_blank`}>文档</a></div>
      </footer>
    </div>
  );
}
