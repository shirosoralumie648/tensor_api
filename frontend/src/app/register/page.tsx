'use client';

import React, { useState } from 'react';
import { Button, Input, Card, Checkbox } from '@/components/ui';
import { Container, VStack, HStack } from '@/components/layout';
import { Mail, Lock, User, ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

export default function RegisterPage() {
  const router = useRouter();
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    password: '',
    confirmPassword: '',
  });
  const [agreed, setAgreed] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    // 验证
    if (!agreed) {
      setError('Please agree to the Terms of Service');
      return;
    }

    if (formData.password !== formData.confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (formData.password.length < 8) {
      setError('Password must be at least 8 characters');
      return;
    }

    setLoading(true);

    try {
      // TODO: 实现注册逻辑
      console.log('Register:', formData);
      // 模拟延迟
      await new Promise(resolve => setTimeout(resolve, 1000));
      router.push('/chat');
    } catch (err) {
      setError('Registration failed, please try again');
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
            Create your account
          </p>
        </div>

        {/* Register Form */}
        <Card className="p-8">
          <form onSubmit={handleRegister} className="space-y-6">
            {error && (
              <div className="p-3 rounded-lg bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800">
                <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
              </div>
            )}

            <Input
              label="Full Name"
              type="text"
              name="name"
              placeholder="John Doe"
              value={formData.name}
              onChange={handleChange}
              leftIcon={<User size={18} />}
              required
            />

            <Input
              label="Email Address"
              type="email"
              name="email"
              placeholder="you@example.com"
              value={formData.email}
              onChange={handleChange}
              leftIcon={<Mail size={18} />}
              required
            />

            <Input
              label="Password"
              type="password"
              name="password"
              placeholder="••••••••"
              value={formData.password}
              onChange={handleChange}
              leftIcon={<Lock size={18} />}
              helperText="Must be at least 8 characters"
              required
            />

            <Input
              label="Confirm Password"
              type="password"
              name="confirmPassword"
              placeholder="••••••••"
              value={formData.confirmPassword}
              onChange={handleChange}
              leftIcon={<Lock size={18} />}
              required
            />

            <Checkbox
              label={
                <>
                  I agree to the{' '}
                  <Link href="/terms" className="text-primary-500 hover:underline">
                    Terms of Service
                  </Link>
                  {' '}and{' '}
                  <Link href="/privacy" className="text-primary-500 hover:underline">
                    Privacy Policy
                  </Link>
                </>
              }
              checked={agreed}
              onChange={(e) => setAgreed(e.target.checked)}
            />

            <Button
              type="submit"
              fullWidth
              loading={loading}
              disabled={!formData.name || !formData.email || !formData.password || !formData.confirmPassword}
            >
              Create Account
            </Button>
          </form>

          {/* Divider */}
          <div className="relative my-6">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-neutral-200 dark:border-neutral-700" />
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-2 bg-white dark:bg-neutral-800 text-neutral-500">
                Or sign up with
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
          Already have an account?{' '}
          <Link href="/login" className="text-primary-500 hover:text-primary-600 dark:hover:text-primary-400 font-medium transition">
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}

