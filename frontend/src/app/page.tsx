'use client';

import React from 'react';
import { Button } from '@/components/ui';
import { Container, VStack, HStack } from '@/components/layout';
import { ArrowRight, Zap, MessageSquare, Shield, Cpu, Globe } from 'lucide-react';
import Link from 'next/link';

export default function Home() {
  return (
    <main className="min-h-screen bg-gradient-to-br from-neutral-50 via-white to-neutral-50 dark:from-neutral-900 dark:via-neutral-800 dark:to-neutral-900">
      {/* 导航栏 */}
      <nav className="sticky top-0 z-40 border-b border-neutral-200 dark:border-neutral-700 bg-white/80 dark:bg-neutral-800/80 backdrop-blur-md">
        <Container>
          <div className="h-16 flex items-center justify-between">
            <div className="text-2xl font-bold bg-gradient-to-r from-primary-500 to-blue-500 bg-clip-text text-transparent">
              Oblivious
            </div>
            <div className="hidden md:flex items-center gap-8">
              <a href="#features" className="text-sm text-neutral-600 hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-white transition">
                Features
              </a>
              <a href="#pricing" className="text-sm text-neutral-600 hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-white transition">
                Pricing
              </a>
              <a href="#docs" className="text-sm text-neutral-600 hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-white transition">
                Docs
              </a>
            </div>
            <div className="flex items-center gap-3">
              <Link href="/login">
                <Button variant="ghost" size="sm">
                  Login
                </Button>
              </Link>
              <Link href="/register">
                <Button size="sm">
                  Start Free
                </Button>
              </Link>
            </div>
          </div>
        </Container>
      </nav>

      {/* Hero Section */}
      <section className="py-20 md:py-32">
        <Container>
          <VStack spacing="lg" className="text-center">
            <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full border border-primary-200 bg-primary-50 dark:bg-primary-900/30 dark:border-primary-800">
              <span className="w-2 h-2 rounded-full bg-primary-500 animate-pulse" />
              <span className="text-sm font-medium text-primary-600 dark:text-primary-400">
                Now in Beta
              </span>
            </div>

            <h1 className="text-5xl md:text-7xl font-bold tracking-tight">
              <span className="bg-gradient-to-r from-primary-600 via-blue-600 to-cyan-600 bg-clip-text text-transparent">
                AI Platform
              </span>
              <br />
              <span className="text-neutral-900 dark:text-white">
                Built for Everyone
              </span>
            </h1>

            <p className="text-xl text-neutral-600 dark:text-neutral-400 max-w-2xl">
              Access any AI model from a single unified interface. Fast, reliable, and powerful. 
              Perfect for developers, businesses, and enterprises.
            </p>

            <HStack spacing="md" className="justify-center flex-wrap">
              <Link href="/chat">
                <Button size="lg" className="gap-2">
                  Start Chatting
                  <ArrowRight size={20} />
                </Button>
              </Link>
              <Link href="/register">
                <Button variant="outline" size="lg">
                  View Demo
                </Button>
              </Link>
            </HStack>

            <div className="mt-8 pt-8 border-t border-neutral-200 dark:border-neutral-700">
              <p className="text-sm text-neutral-500 dark:text-neutral-400 mb-4">
                Trusted by thousands of users worldwide
              </p>
              <div className="flex items-center justify-center gap-8">
                <div className="text-center">
                  <div className="text-2xl font-bold text-neutral-900 dark:text-white">1M+</div>
                  <div className="text-xs text-neutral-500">API Calls</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-neutral-900 dark:text-white">50K+</div>
                  <div className="text-xs text-neutral-500">Active Users</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-neutral-900 dark:text-white">99.9%</div>
                  <div className="text-xs text-neutral-500">Uptime</div>
                </div>
              </div>
            </div>
          </VStack>
        </Container>
      </section>

      {/* Features Section */}
      <section id="features" className="py-20 md:py-32 border-t border-neutral-200 dark:border-neutral-700">
        <Container>
          <VStack spacing="xl">
            <div className="text-center">
              <h2 className="text-4xl md:text-5xl font-bold mb-4 text-neutral-900 dark:text-white">
                Powerful Features
              </h2>
              <p className="text-xl text-neutral-600 dark:text-neutral-400 max-w-2xl mx-auto">
                Everything you need to build amazing AI applications
              </p>
            </div>

            <div className="grid md:grid-cols-3 gap-8 mt-12">
              {/* Feature 1 */}
              <div className="p-6 rounded-2xl border border-neutral-200 dark:border-neutral-700 hover:border-primary-200 dark:hover:border-primary-800 transition group">
                <div className="mb-4 inline-flex p-3 rounded-lg bg-primary-100 dark:bg-primary-900/30 group-hover:bg-primary-200 dark:group-hover:bg-primary-800 transition">
                  <Zap className="w-6 h-6 text-primary-600 dark:text-primary-400" />
                </div>
                <h3 className="text-xl font-semibold mb-2 text-neutral-900 dark:text-white">
                  Ultra Fast
                </h3>
                <p className="text-neutral-600 dark:text-neutral-400">
                  Lightning-fast API responses with optimized infrastructure
                </p>
              </div>

              {/* Feature 2 */}
              <div className="p-6 rounded-2xl border border-neutral-200 dark:border-neutral-700 hover:border-primary-200 dark:hover:border-primary-800 transition group">
                <div className="mb-4 inline-flex p-3 rounded-lg bg-primary-100 dark:bg-primary-900/30 group-hover:bg-primary-200 dark:group-hover:bg-primary-800 transition">
                  <MessageSquare className="w-6 h-6 text-primary-600 dark:text-primary-400" />
                </div>
                <h3 className="text-xl font-semibold mb-2 text-neutral-900 dark:text-white">
                  Multiple Models
                </h3>
                <p className="text-neutral-600 dark:text-neutral-400">
                  Access GPT-4, Claude, Gemini and more from one interface
                </p>
              </div>

              {/* Feature 3 */}
              <div className="p-6 rounded-2xl border border-neutral-200 dark:border-neutral-700 hover:border-primary-200 dark:hover:border-primary-800 transition group">
                <div className="mb-4 inline-flex p-3 rounded-lg bg-primary-100 dark:bg-primary-900/30 group-hover:bg-primary-200 dark:group-hover:bg-primary-800 transition">
                  <Shield className="w-6 h-6 text-primary-600 dark:text-primary-400" />
                </div>
                <h3 className="text-xl font-semibold mb-2 text-neutral-900 dark:text-white">
                  Secure & Private
                </h3>
                <p className="text-neutral-600 dark:text-neutral-400">
                  Enterprise-grade security with end-to-end encryption
                </p>
              </div>

              {/* Feature 4 */}
              <div className="p-6 rounded-2xl border border-neutral-200 dark:border-neutral-700 hover:border-primary-200 dark:hover:border-primary-800 transition group">
                <div className="mb-4 inline-flex p-3 rounded-lg bg-primary-100 dark:bg-primary-900/30 group-hover:bg-primary-200 dark:group-hover:bg-primary-800 transition">
                  <Cpu className="w-6 h-6 text-primary-600 dark:text-primary-400" />
                </div>
                <h3 className="text-xl font-semibold mb-2 text-neutral-900 dark:text-white">
                  Powerful API
                </h3>
                <p className="text-neutral-600 dark:text-neutral-400">
                  Simple yet powerful API for developers
                </p>
              </div>

              {/* Feature 5 */}
              <div className="p-6 rounded-2xl border border-neutral-200 dark:border-neutral-700 hover:border-primary-200 dark:hover:border-primary-800 transition group">
                <div className="mb-4 inline-flex p-3 rounded-lg bg-primary-100 dark:bg-primary-900/30 group-hover:bg-primary-200 dark:group-hover:bg-primary-800 transition">
                  <Globe className="w-6 h-6 text-primary-600 dark:text-primary-400" />
                </div>
                <h3 className="text-xl font-semibold mb-2 text-neutral-900 dark:text-white">
                  Global Scale
                </h3>
                <p className="text-neutral-600 dark:text-neutral-400">
                  Deployed worldwide with low latency everywhere
                </p>
              </div>

              {/* Feature 6 */}
              <div className="p-6 rounded-2xl border border-neutral-200 dark:border-neutral-700 hover:border-primary-200 dark:hover:border-primary-800 transition group">
                <div className="mb-4 inline-flex p-3 rounded-lg bg-primary-100 dark:bg-primary-900/30 group-hover:bg-primary-200 dark:group-hover:bg-primary-800 transition">
                  <Zap className="w-6 h-6 text-primary-600 dark:text-primary-400" />
                </div>
                <h3 className="text-xl font-semibold mb-2 text-neutral-900 dark:text-white">
                  Real-time Updates
                </h3>
                <p className="text-neutral-600 dark:text-neutral-400">
                  Live streaming responses and real-time collaboration
                </p>
              </div>
            </div>
          </VStack>
        </Container>
      </section>

      {/* CTA Section */}
      <section className="py-20 md:py-32 border-t border-neutral-200 dark:border-neutral-700">
        <Container>
          <div className="bg-gradient-to-r from-primary-600 to-blue-600 rounded-3xl p-12 md:p-20 text-center text-white">
            <h2 className="text-4xl md:text-5xl font-bold mb-6">
              Ready to Get Started?
            </h2>
            <p className="text-xl mb-8 text-blue-100 max-w-2xl mx-auto">
              Join thousands of developers using Oblivious to power their AI applications
            </p>
            <HStack spacing="md" className="justify-center flex-wrap">
              <Link href="/register">
                <Button size="lg" className="bg-white text-primary-600 hover:bg-neutral-100">
                  Start Free Trial
                </Button>
              </Link>
              <Link href="#docs">
                <Button variant="outline" size="lg" className="border-white text-white hover:bg-white/10">
                  View Documentation
                </Button>
              </Link>
            </HStack>
          </div>
        </Container>
      </section>

      {/* Footer */}
      <footer className="border-t border-neutral-200 dark:border-neutral-700 py-12">
        <Container>
          <div className="grid md:grid-cols-4 gap-8 mb-8">
            <div>
              <h4 className="font-semibold text-neutral-900 dark:text-white mb-4">Product</h4>
              <ul className="space-y-2 text-sm text-neutral-600 dark:text-neutral-400">
                <li><a href="#" className="hover:text-neutral-900 dark:hover:text-white transition">Features</a></li>
                <li><a href="#" className="hover:text-neutral-900 dark:hover:text-white transition">Pricing</a></li>
                <li><a href="#" className="hover:text-neutral-900 dark:hover:text-white transition">API</a></li>
              </ul>
            </div>
            <div>
              <h4 className="font-semibold text-neutral-900 dark:text-white mb-4">Resources</h4>
              <ul className="space-y-2 text-sm text-neutral-600 dark:text-neutral-400">
                <li><a href="#" className="hover:text-neutral-900 dark:hover:text-white transition">Documentation</a></li>
                <li><a href="#" className="hover:text-neutral-900 dark:hover:text-white transition">Blog</a></li>
                <li><a href="#" className="hover:text-neutral-900 dark:hover:text-white transition">Community</a></li>
              </ul>
            </div>
            <div>
              <h4 className="font-semibold text-neutral-900 dark:text-white mb-4">Company</h4>
              <ul className="space-y-2 text-sm text-neutral-600 dark:text-neutral-400">
                <li><a href="#" className="hover:text-neutral-900 dark:hover:text-white transition">About</a></li>
                <li><a href="#" className="hover:text-neutral-900 dark:hover:text-white transition">Contact</a></li>
                <li><a href="#" className="hover:text-neutral-900 dark:hover:text-white transition">Careers</a></li>
              </ul>
            </div>
            <div>
              <h4 className="font-semibold text-neutral-900 dark:text-white mb-4">Legal</h4>
              <ul className="space-y-2 text-sm text-neutral-600 dark:text-neutral-400">
                <li><a href="#" className="hover:text-neutral-900 dark:hover:text-white transition">Privacy</a></li>
                <li><a href="#" className="hover:text-neutral-900 dark:hover:text-white transition">Terms</a></li>
                <li><a href="#" className="hover:text-neutral-900 dark:hover:text-white transition">Security</a></li>
              </ul>
            </div>
          </div>
          <div className="border-t border-neutral-200 dark:border-neutral-700 pt-8 flex items-center justify-between">
            <p className="text-sm text-neutral-600 dark:text-neutral-400">
              © 2024 Oblivious. All rights reserved.
            </p>
            <div className="flex items-center gap-4">
              <a href="#" className="text-neutral-600 hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-white transition">Twitter</a>
              <a href="#" className="text-neutral-600 hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-white transition">GitHub</a>
              <a href="#" className="text-neutral-600 hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-white transition">Discord</a>
            </div>
          </div>
        </Container>
      </footer>
    </main>
  );
}
