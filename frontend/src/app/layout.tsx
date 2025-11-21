import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'Oblivious - AI Application Platform',
  description: 'A modern AI application platform powered by advanced language models',
  viewport: {
    width: 'device-width',
    initialScale: 1,
    maximumScale: 1,
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-CN">
      <body className="bg-gray-50">{children}</body>
    </html>
  );
}

