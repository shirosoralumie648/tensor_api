'use client';

import React, { useState } from 'react';
import { Button, Card, Input, Textarea, Select } from '@/components/ui';
import { Container, VStack, HStack, Navbar, NavbarContent, NavbarBrand, NavbarMenu, NavbarItem, NavbarActions } from '@/components/layout';
import { User, CreditCard, Clock, Settings, LogOut, Copy, Globe } from 'lucide-react';
import Link from 'next/link';
import { useTranslation } from '@/hooks/useTranslation';

type TabType = 'profile' | 'quota' | 'history' | 'settings';

export default function UserPage() {
  const { t, language, setLanguage } = useTranslation();
  const [activeTab, setActiveTab] = useState<TabType>('profile');
  const [copied, setCopied] = useState(false);

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="min-h-screen bg-neutral-50 dark:bg-neutral-900">
      {/* Navbar */}
      <Navbar>
        <NavbarContent>
          <NavbarBrand>
            <Link href="/">
              <div className="text-2xl font-bold bg-gradient-to-r from-primary-500 to-blue-500 bg-clip-text text-transparent cursor-pointer">
                Oblivious
              </div>
            </Link>
          </NavbarBrand>
          <NavbarMenu>
            <NavbarItem href="/chat">Chat</NavbarItem>
            <NavbarItem href="/developer">Developer</NavbarItem>
          </NavbarMenu>
          <NavbarActions>
            <Link href="/login">
              <Button variant="outline" size="sm">
                Logout
              </Button>
            </Link>
          </NavbarActions>
        </NavbarContent>
      </Navbar>

      {/* Main Content */}
      <Container size="2xl" className="py-8">
        <div className="grid md:grid-cols-4 gap-6">
          {/* Sidebar */}
          <div className="md:col-span-1">
            <Card className="p-0">
              <nav className="space-y-1">
                <button
                  onClick={() => setActiveTab('profile')}
                  className={`w-full text-left px-4 py-3 flex items-center gap-3 transition ${
                    activeTab === 'profile'
                      ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 border-l-4 border-primary-500'
                      : 'text-neutral-700 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-neutral-800'
                  }`}
                >
                  <User size={18} />
                  Profile
                </button>
                <button
                  onClick={() => setActiveTab('quota')}
                  className={`w-full text-left px-4 py-3 flex items-center gap-3 transition ${
                    activeTab === 'quota'
                      ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 border-l-4 border-primary-500'
                      : 'text-neutral-700 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-neutral-800'
                  }`}
                >
                  <CreditCard size={18} />
                  Quota
                </button>
                <button
                  onClick={() => setActiveTab('history')}
                  className={`w-full text-left px-4 py-3 flex items-center gap-3 transition ${
                    activeTab === 'history'
                      ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 border-l-4 border-primary-500'
                      : 'text-neutral-700 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-neutral-800'
                  }`}
                >
                  <Clock size={18} />
                  History
                </button>
                <button
                  onClick={() => setActiveTab('settings')}
                  className={`w-full text-left px-4 py-3 flex items-center gap-3 transition ${
                    activeTab === 'settings'
                      ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 border-l-4 border-primary-500'
                      : 'text-neutral-700 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-neutral-800'
                  }`}
                >
                  <Settings size={18} />
                  Settings
                </button>
              </nav>
            </Card>
          </div>

          {/* Content */}
          <div className="md:col-span-3">
            {/* Profile Tab */}
            {activeTab === 'profile' && (
              <VStack spacing="lg">
                <Card>
                  <div className="flex items-start justify-between mb-6">
                    <h2 className="text-2xl font-bold text-neutral-900 dark:text-white">
                      Profile Settings
                    </h2>
                  </div>

                  <VStack spacing="md">
                    <Input
                      label="Full Name"
                      defaultValue="John Doe"
                    />
                    <Input
                      label="Email"
                      type="email"
                      defaultValue="john@example.com"
                      disabled
                    />
                    <Textarea
                      label="Bio"
                      placeholder="Tell us about yourself..."
                      defaultValue="AI enthusiast and developer"
                    />
                    <HStack spacing="md" justify="end">
                      <Button variant="outline">Cancel</Button>
                      <Button>Save Changes</Button>
                    </HStack>
                  </VStack>
                </Card>
              </VStack>
            )}

            {/* Quota Tab */}
            {activeTab === 'quota' && (
              <VStack spacing="lg">
                <Card>
                  <h2 className="text-2xl font-bold text-neutral-900 dark:text-white mb-6">
                    API Quota
                  </h2>

                  <VStack spacing="lg">
                    <div>
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm font-medium text-neutral-700 dark:text-neutral-300">
                          Monthly Usage
                        </span>
                        <span className="text-sm text-neutral-600 dark:text-neutral-400">
                          750 / 1,000 requests
                        </span>
                      </div>
                      <div className="w-full h-2 bg-neutral-200 dark:bg-neutral-700 rounded-full overflow-hidden">
                        <div className="h-full w-3/4 bg-primary-500 transition-all duration-300" />
                      </div>
                    </div>

                    <div>
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm font-medium text-neutral-700 dark:text-neutral-300">
                          Daily Limit
                        </span>
                        <span className="text-sm text-neutral-600 dark:text-neutral-400">
                          45 / 100 requests
                        </span>
                      </div>
                      <div className="w-full h-2 bg-neutral-200 dark:bg-neutral-700 rounded-full overflow-hidden">
                        <div className="h-full w-5/12 bg-green-500 transition-all duration-300" />
                      </div>
                    </div>

                    <div className="p-4 rounded-lg bg-blue-50 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-800">
                      <p className="text-sm text-blue-900 dark:text-blue-100">
                        💡 Upgrade to Pro to get 10x more quota and priority support
                      </p>
                    </div>

                    <Button fullWidth>
                      View Pricing Plans
                    </Button>
                  </VStack>
                </Card>

                {/* API Keys */}
                <Card>
                  <h3 className="text-xl font-bold text-neutral-900 dark:text-white mb-4">
                    API Keys
                  </h3>
                  <VStack spacing="md">
                    <div className="p-4 rounded-lg bg-neutral-100 dark:bg-neutral-700">
                      <div className="flex items-center justify-between">
                        <code className="text-sm font-mono text-neutral-600 dark:text-neutral-300">
                          sk-••••••••••••••••••••
                        </code>
                        <button
                          onClick={() => copyToClipboard('sk-1234567890abcdefghij')}
                          className="p-2 hover:bg-neutral-200 dark:hover:bg-neutral-600 rounded transition"
                        >
                          <Copy size={18} className="text-neutral-600 dark:text-neutral-400" />
                        </button>
                      </div>
                      <p className="text-xs text-neutral-500 mt-2">
                        Created on Nov 20, 2024
                      </p>
                    </div>
                    <Button fullWidth variant="outline">
                      Generate New Key
                    </Button>
                  </VStack>
                </Card>
              </VStack>
            )}

            {/* History Tab */}
            {activeTab === 'history' && (
              <Card>
                <h2 className="text-2xl font-bold text-neutral-900 dark:text-white mb-6">
                  Usage History
                </h2>
                <div className="space-y-4">
                  {[1, 2, 3, 4, 5].map(i => (
                    <div key={i} className="p-4 rounded-lg border border-neutral-200 dark:border-neutral-700 hover:bg-neutral-50 dark:hover:bg-neutral-800 transition">
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="font-medium text-neutral-900 dark:text-white">
                            API Request #{i}
                          </p>
                          <p className="text-sm text-neutral-500">
                            gpt-4 • 150 tokens • Today at 10:{String(i).padStart(2, '0')}:00
                          </p>
                        </div>
                        <div className="text-right">
                          <p className="text-sm font-medium text-neutral-900 dark:text-white">
                            $0.0{i}
                          </p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </Card>
            )}

            {/* Settings Tab */}
            {activeTab === 'settings' && (
              <VStack spacing="lg">
                <Card>
                  <h2 className="text-2xl font-bold text-neutral-900 dark:text-white mb-6">
                    Settings
                  </h2>

                  <VStack spacing="md">
                    {/* Language Selection */}
                    <div className="pb-4 border-b border-neutral-200 dark:border-neutral-700">
                      <div className="flex items-start justify-between gap-4">
                        <div className="flex-1">
                          <p className="font-medium text-neutral-900 dark:text-white mb-1">
                            {t('settings.language')}
                          </p>
                          <p className="text-sm text-neutral-500">
                            {t('settings.languageDesc')}
                          </p>
                        </div>
                        <Select
                          value={language}
                          onChange={(e) => setLanguage(e.target.value as any)}
                          options={[
                            { value: 'en', label: 'English' },
                            { value: 'zh', label: '中文' },
                            { value: 'ja', label: '日本語' },
                            { value: 'es', label: 'Español' },
                            { value: 'fr', label: 'Français' },
                            { value: 'de', label: 'Deutsch' },
                          ]}
                        />
                      </div>
                    </div>

                    <div className="pb-4 border-b border-neutral-200 dark:border-neutral-700">
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="font-medium text-neutral-900 dark:text-white">
                            {t('settings.notifications')}
                          </p>
                          <p className="text-sm text-neutral-500">
                            {t('settings.notificationsDesc')}
                          </p>
                        </div>
                        <input type="checkbox" defaultChecked className="w-4 h-4" />
                      </div>
                    </div>

                    <div className="pb-4 border-b border-neutral-200 dark:border-neutral-700">
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="font-medium text-neutral-900 dark:text-white">
                            {t('settings.marketing')}
                          </p>
                          <p className="text-sm text-neutral-500">
                            {t('settings.marketingDesc')}
                          </p>
                        </div>
                        <input type="checkbox" className="w-4 h-4" />
                      </div>
                    </div>

                    <Button fullWidth variant="danger">
                      {t('settings.deleteAccount')}
                    </Button>
                  </VStack>
                </Card>
              </VStack>
            )}
          </div>
        </div>
      </Container>
    </div>
  );
}

