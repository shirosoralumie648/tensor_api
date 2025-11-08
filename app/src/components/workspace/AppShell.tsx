import NavBar from "@/components/app/NavBar.tsx";
import type { ReactNode } from "react";

type AppShellProps = {
  left?: ReactNode;
  right?: ReactNode;
  children?: ReactNode;
};

export default function AppShell({ left, right, children }: AppShellProps) {
  return (
    <div className={`min-h-screen flex flex-col bg-background text-foreground`}>
      <NavBar />
      <div className={`flex-1 w-full mx-auto max-w-[1400px] px-3 md:px-6 py-3 md:py-4` }>
        <div className={`grid grid-cols-12 gap-3 md:gap-4`}>
          <aside className={`hidden lg:block col-span-2`}>{left}</aside>
          <main className={`col-span-12 lg:col-span-7 xl:col-span-8`}>{children}</main>
          <aside className={`hidden xl:block col-span-3`}>{right}</aside>
        </div>
      </div>
    </div>
  );
}
