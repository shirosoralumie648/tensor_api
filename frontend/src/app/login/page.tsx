'use client';

import React, { useState } from 'react';
import { Button, Input, Card } from '@/components/ui';
import { Container, VStack, HStack } from '@/components/layout';
import { Mail, Lock, ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      // TODO: 实现登录逻辑
      console.log('Login:', { email, password });
      // 模拟延迟
      await new Promise(resolve => setTimeout(resolve, 1000));
      router.push('/chat');
    } catch (err) {
      setError('登录失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-neutral-50 to-neutral-100 dark:from-neutral-900 dark:to-neutral-800 flex items-center justify-center py-12 px-4">
      <div className="w-full max-w-md">
        {/* Header */}
        <div className="mb-8 text-center">
          <Link href="/" className="inline-flex items-center gap-2 text-sm text-neutral-600 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-white transition mb-6">
            <ArrowLeft size={16} />
            Back to Home
          </Link>
          <div className="text-3xl font-bold bg-gradient-to-r from-primary-500 to-blue-500 bg-clip-text text-transparent mb-2">
            Oblivious
          </div>
          <p className="text-neutral-600 dark:text-neutral-400">
            Sign in to your account
          </p>
        </div>

        {/* Login Form */}
        <Card className="p-8">
          <form onSubmit={handleLogin} className="space-y-6">
            {error && (
              <div className="p-3 rounded-lg bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800">
                <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
              </div>
            )}

            <Input
              label="Email Address"
              type="email"
              placeholder="you@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              leftIcon={<Mail size={18} />}
              required
            />

            <VStack spacing="sm">
              <Input
                label="Password"
                type="password"
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                leftIcon={<Lock size={18} />}
                required
              />
              <Link href="/forgot-password" className="text-xs text-primary-500 hover:text-primary-600 dark:hover:text-primary-400 transition">
                Forgot password?
              </Link>
            </VStack>

            <Button
              type="submit"
              fullWidth
              loading={loading}
              disabled={!email || !password}
            >
              Sign In
            </Button>
          </form>

          {/* Divider */}
          <div className="relative my-6">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-neutral-200 dark:border-neutral-700" />
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-2 bg-white dark:bg-neutral-800 text-neutral-500">
                Or continue with
              </span>
            </div>
          </div>

          {/* Social Login */}
          <div className="grid grid-cols-2 gap-3">
            <Button variant="outline" fullWidth>
              GitHub
            </Button>
            <Button variant="outline" fullWidth>
              Google
            </Button>
          </div>
        </Card>

        {/* Footer */}
        <p className="text-center text-sm text-neutral-600 dark:text-neutral-400 mt-6">
          Don't have an account?{' '}
          <Link href="/register" className="text-primary-500 hover:text-primary-600 dark:hover:text-primary-400 font-medium transition">
            Sign up
          </Link>
        </p>
      </div>
    </div>
  );
}
