'use client';

import React, { useState } from 'react';
import { Button, Card, Input, Select, Alert } from '@/components/ui';
import { Container, VStack, HStack, Navbar, NavbarContent, NavbarBrand, NavbarMenu, NavbarItem, NavbarActions } from '@/components/layout';
import { Copy, Key, BarChart3, FileText, Play, CreditCard, Settings, LogOut } from 'lucide-react';
import Link from 'next/link';
import { useTranslation } from '@/hooks/useTranslation';

type DeveloperTab = 'dashboard' | 'keys' | 'docs' | 'playground' | 'billing';

interface APIKey {
  id: string;
  name: string;
  key: string;
  lastUsed?: Date;
  created: Date;
  requests: number;
}

interface DashboardStats {
  totalRequests: number;
  totalTokens: number;
  monthlyUsage: number;
  successRate: number;
}

export default function DeveloperPage() {
  const { t, language, setLanguage } = useTranslation();
  const [activeTab, setActiveTab] = useState<DeveloperTab>('dashboard');
  const [copied, setCopied] = useState(false);
  const [apiKeys, setApiKeys] = useState<APIKey[]>([
    {
      id: '1',
      name: 'Production',
      key: 'sk-proj-xxxxxxxxxxxxxxxxxxxx',
      lastUsed: new Date(Date.now() - 3600000),
      created: new Date(Date.now() - 30 * 24 * 3600000),
      requests: 15234,
    },
    {
      id: '2',
      name: 'Development',
      key: 'sk-dev-yyyyyyyyyyyyyyyyyyyy',
      created: new Date(Date.now() - 7 * 24 * 3600000),
      requests: 3421,
    },
  ]);

  const [showCreateKeyModal, setShowCreateKeyModal] = useState(false);
  const [newKeyName, setNewKeyName] = useState('');

  const stats: DashboardStats = {
    totalRequests: 18655,
    totalTokens: 2847392,
    monthlyUsage: 856.32,
    successRate: 99.87,
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const createNewKey = () => {
    const newKey: APIKey = {
      id: String(apiKeys.length + 1),
      name: newKeyName || 'New Key',
      key: `sk-${Math.random().toString(36).substr(2, 20)}`,
      created: new Date(),
      requests: 0,
    };
    setApiKeys([...apiKeys, newKey]);
    setNewKeyName('');
    setShowCreateKeyModal(false);
  };

  const deleteKey = (id: string) => {
    setApiKeys(apiKeys.filter(key => key.id !== id));
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
            <NavbarItem href="/chat">{t('nav.chat')}</NavbarItem>
            <NavbarItem href="/user">{t('nav.settings')}</NavbarItem>
          </NavbarMenu>
          <NavbarActions>
            <Link href="/login">
              <Button variant="outline" size="sm">
                {t('nav.logout')}
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
            <Card className="sticky top-8">
              <VStack spacing="sm">
                <button
                  onClick={() => setActiveTab('dashboard')}
                  className={`w-full text-left px-4 py-3 rounded-lg font-medium transition ${
                    activeTab === 'dashboard'
                      ? 'bg-primary-500 text-white'
                      : 'text-neutral-600 dark:text-neutral-400 hover:bg-neutral-100 dark:hover:bg-neutral-800'
                  }`}
                >
                  <BarChart3 className="inline mr-2" size={18} />
                  Dashboard
                </button>
                <button
                  onClick={() => setActiveTab('keys')}
                  className={`w-full text-left px-4 py-3 rounded-lg font-medium transition ${
                    activeTab === 'keys'
                      ? 'bg-primary-500 text-white'
                      : 'text-neutral-600 dark:text-neutral-400 hover:bg-neutral-100 dark:hover:bg-neutral-800'
                  }`}
                >
                  <Key className="inline mr-2" size={18} />
                  API Keys
                </button>
                <button
                  onClick={() => setActiveTab('docs')}
                  className={`w-full text-left px-4 py-3 rounded-lg font-medium transition ${
                    activeTab === 'docs'
                      ? 'bg-primary-500 text-white'
                      : 'text-neutral-600 dark:text-neutral-400 hover:bg-neutral-100 dark:hover:bg-neutral-800'
                  }`}
                >
                  <FileText className="inline mr-2" size={18} />
                  API Docs
                </button>
                <button
                  onClick={() => setActiveTab('playground')}
                  className={`w-full text-left px-4 py-3 rounded-lg font-medium transition ${
                    activeTab === 'playground'
                      ? 'bg-primary-500 text-white'
                      : 'text-neutral-600 dark:text-neutral-400 hover:bg-neutral-100 dark:hover:bg-neutral-800'
                  }`}
                >
                  <Play className="inline mr-2" size={18} />
                  Playground
                </button>
                <button
                  onClick={() => setActiveTab('billing')}
                  className={`w-full text-left px-4 py-3 rounded-lg font-medium transition ${
                    activeTab === 'billing'
                      ? 'bg-primary-500 text-white'
                      : 'text-neutral-600 dark:text-neutral-400 hover:bg-neutral-100 dark:hover:bg-neutral-800'
                  }`}
                >
                  <CreditCard className="inline mr-2" size={18} />
                  Billing
                </button>
              </VStack>
            </Card>
          </div>

          {/* Main Content Area */}
          <div className="md:col-span-3">
            {/* Dashboard Tab */}
            {activeTab === 'dashboard' && (
              <VStack spacing="lg">
                <h1 className="text-3xl font-bold text-neutral-900 dark:text-white">
                  API Dashboard
                </h1>

                {/* Stats Grid */}
                <div className="grid md:grid-cols-2 gap-4">
                  {/* Total Requests */}
                  <Card>
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-sm text-neutral-600 dark:text-neutral-400">
                          Total Requests
                        </p>
                        <p className="text-3xl font-bold text-neutral-900 dark:text-white mt-2">
                          {stats.totalRequests.toLocaleString()}
                        </p>
                      </div>
                      <BarChart3 size={40} className="text-primary-500 opacity-50" />
                    </div>
                  </Card>

                  {/* Total Tokens */}
                  <Card>
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-sm text-neutral-600 dark:text-neutral-400">
                          Total Tokens
                        </p>
                        <p className="text-3xl font-bold text-neutral-900 dark:text-white mt-2">
                          {(stats.totalTokens / 1000000).toFixed(2)}M
                        </p>
                      </div>
                      <FileText size={40} className="text-blue-500 opacity-50" />
                    </div>
                  </Card>

                  {/* Monthly Usage */}
                  <Card>
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-sm text-neutral-600 dark:text-neutral-400">
                          Monthly Usage
                        </p>
                        <p className="text-3xl font-bold text-neutral-900 dark:text-white mt-2">
                          ${stats.monthlyUsage.toFixed(2)}
                        </p>
                      </div>
                      <CreditCard size={40} className="text-green-500 opacity-50" />
                    </div>
                  </Card>

                  {/* Success Rate */}
                  <Card>
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-sm text-neutral-600 dark:text-neutral-400">
                          Success Rate
                        </p>
                        <p className="text-3xl font-bold text-neutral-900 dark:text-white mt-2">
                          {stats.successRate.toFixed(2)}%
                        </p>
                      </div>
                      <div className="text-yellow-500 opacity-50" style={{ fontSize: '32px' }}>
                        ✓
                      </div>
                    </div>
                  </Card>
                </div>

                {/* Recent Activity */}
                <Card>
                  <h2 className="text-xl font-bold text-neutral-900 dark:text-white mb-4">
                    Recent Activity
                  </h2>
                  <VStack spacing="md">
                    {[1, 2, 3, 4, 5].map(i => (
                      <div key={i} className="pb-3 border-b border-neutral-200 dark:border-neutral-700 last:border-b-0">
                        <div className="flex items-center justify-between">
                          <div>
                            <p className="text-sm font-medium text-neutral-900 dark:text-white">
                              API Request #{i}
                            </p>
                            <p className="text-xs text-neutral-500">
                              gpt-4 • {Math.floor(Math.random() * 2000)} tokens • Today at {String(10 + i).padStart(2, '0')}:00:00
                            </p>
                          </div>
                          <span className="text-xs font-medium px-2 py-1 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 rounded">
                            Success
                          </span>
                        </div>
                      </div>
                    ))}
                  </VStack>
                </Card>
              </VStack>
            )}

            {/* API Keys Tab */}
            {activeTab === 'keys' && (
              <VStack spacing="lg">
                <div className="flex items-center justify-between">
                  <h1 className="text-3xl font-bold text-neutral-900 dark:text-white">
                    API Keys
                  </h1>
                  <Button onClick={() => setShowCreateKeyModal(true)}>
                    <Key size={18} className="mr-2" />
                    Create Key
                  </Button>
                </div>

                {/* Create Key Modal */}
                {showCreateKeyModal && (
                  <Card className="border-2 border-primary-500">
                    <h2 className="text-lg font-bold text-neutral-900 dark:text-white mb-4">
                      Create New API Key
                    </h2>
                    <VStack spacing="md">
                      <Input
                        label="Key Name"
                        placeholder="e.g., Production, Development"
                        value={newKeyName}
                        onChange={(e) => setNewKeyName(e.target.value)}
                      />
                      <HStack spacing="md">
                        <Button onClick={createNewKey} fullWidth>
                          Create
                        </Button>
                        <Button
                          onClick={() => setShowCreateKeyModal(false)}
                          variant="outline"
                          fullWidth
                        >
                          Cancel
                        </Button>
                      </HStack>
                    </VStack>
                  </Card>
                )}

                {/* API Keys List */}
                <VStack spacing="md">
                  {apiKeys.map(key => (
                    <Card key={key.id}>
                      <VStack spacing="md">
                        <div className="flex items-center justify-between">
                          <div>
                            <h3 className="text-lg font-bold text-neutral-900 dark:text-white">
                              {key.name}
                            </h3>
                            <p className="text-sm text-neutral-500 mt-1">
                              Created {key.created.toLocaleDateString()}
                            </p>
                          </div>
                          <button
                            onClick={() => deleteKey(key.id)}
                            className="px-3 py-2 text-sm font-medium text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 rounded transition"
                          >
                            Delete
                          </button>
                        </div>

                        {/* Key Display */}
                        <div className="bg-neutral-100 dark:bg-neutral-800 rounded-lg p-3 flex items-center justify-between">
                          <code className="text-sm text-neutral-700 dark:text-neutral-300 font-mono">
                            {key.key}
                          </code>
                          <button
                            onClick={() => copyToClipboard(key.key)}
                            className="p-2 hover:bg-neutral-200 dark:hover:bg-neutral-700 rounded transition"
                          >
                            <Copy size={18} className="text-neutral-600 dark:text-neutral-400" />
                          </button>
                        </div>

                        {/* Key Stats */}
                        <HStack spacing="md">
                          <div className="flex-1">
                            <p className="text-xs text-neutral-600 dark:text-neutral-400">
                              Total Requests
                            </p>
                            <p className="text-lg font-bold text-neutral-900 dark:text-white">
                              {key.requests.toLocaleString()}
                            </p>
                          </div>
                          <div className="flex-1">
                            <p className="text-xs text-neutral-600 dark:text-neutral-400">
                              Last Used
                            </p>
                            <p className="text-lg font-bold text-neutral-900 dark:text-white">
                              {key.lastUsed ? key.lastUsed.toLocaleDateString() : 'Never'}
                            </p>
                          </div>
                        </HStack>
                      </VStack>
                    </Card>
                  ))}
                </VStack>
              </VStack>
            )}

            {/* API Docs Tab */}
            {activeTab === 'docs' && (
              <VStack spacing="lg">
                <h1 className="text-3xl font-bold text-neutral-900 dark:text-white">
                  API Documentation
                </h1>

                <Alert variant="info">
                  API documentation is automatically generated from your backend endpoints.
                </Alert>

                <Card>
                  <h2 className="text-xl font-bold text-neutral-900 dark:text-white mb-4">
                    Chat Endpoints
                  </h2>
                  <VStack spacing="md">
                    {['POST /api/chat/completions', 'GET /api/chat/conversations', 'POST /api/chat/messages'].map(endpoint => (
                      <div key={endpoint} className="pb-4 border-b border-neutral-200 dark:border-neutral-700 last:border-b-0">
                        <code className="text-sm font-mono text-primary-600 dark:text-primary-400">
                          {endpoint}
                        </code>
                        <p className="text-sm text-neutral-600 dark:text-neutral-400 mt-2">
                          Endpoint description and usage examples...
                        </p>
                      </div>
                    ))}
                  </VStack>
                </Card>

                <Button fullWidth>
                  View Full Documentation
                </Button>
              </VStack>
            )}

            {/* Playground Tab */}
            {activeTab === 'playground' && (
              <VStack spacing="lg">
                <h1 className="text-3xl font-bold text-neutral-900 dark:text-white">
                  API Playground
                </h1>

                <div className="grid md:grid-cols-2 gap-6">
                  {/* Request */}
                  <Card>
                    <h2 className="text-lg font-bold text-neutral-900 dark:text-white mb-4">
                      Request
                    </h2>
                    <VStack spacing="md">
                      <div>
                        <label className="text-sm font-medium text-neutral-700 dark:text-neutral-300 mb-2 block">
                          Method
                        </label>
                        <Select
                          options={[
                            { value: 'post', label: 'POST' },
                            { value: 'get', label: 'GET' },
                            { value: 'put', label: 'PUT' },
                          ]}
                          defaultValue="post"
                        />
                      </div>
                      <div>
                        <label className="text-sm font-medium text-neutral-700 dark:text-neutral-300 mb-2 block">
                          Endpoint
                        </label>
                        <Input
                          placeholder="/api/chat/completions"
                          defaultValue="/api/chat/completions"
                        />
                      </div>
                      <div>
                        <label className="text-sm font-medium text-neutral-700 dark:text-neutral-300 mb-2 block">
                          Request Body
                        </label>
                        <textarea
                          className="w-full h-32 px-3 py-2 border border-neutral-300 dark:border-neutral-600 rounded-lg bg-white dark:bg-neutral-800 text-neutral-900 dark:text-white"
                          defaultValue={JSON.stringify(
                            { model: 'gpt-4', messages: [{ role: 'user', content: 'Hello' }] },
                            null,
                            2
                          )}
                        />
                      </div>
                      <Button fullWidth>
                        <Play size={18} className="mr-2" />
                        Send Request
                      </Button>
                    </VStack>
                  </Card>

                  {/* Response */}
                  <Card>
                    <h2 className="text-lg font-bold text-neutral-900 dark:text-white mb-4">
                      Response
                    </h2>
                    <VStack spacing="md">
                      <div>
                        <label className="text-sm font-medium text-neutral-700 dark:text-neutral-300 mb-2 block">
                          Status: 200 OK
                        </label>
                      </div>
                      <textarea
                        className="w-full h-32 px-3 py-2 border border-neutral-300 dark:border-neutral-600 rounded-lg bg-neutral-50 dark:bg-neutral-800 text-neutral-900 dark:text-white font-mono text-sm"
                        readOnly
                        defaultValue={JSON.stringify(
                          {
                            choices: [
                              {
                                message: { role: 'assistant', content: 'Hello! How can I help you?' },
                              },
                            ],
                          },
                          null,
                          2
                        )}
                      />
                    </VStack>
                  </Card>
                </div>
              </VStack>
            )}

            {/* Billing Tab */}
            {activeTab === 'billing' && (
              <VStack spacing="lg">
                <h1 className="text-3xl font-bold text-neutral-900 dark:text-white">
                  Billing & Usage
                </h1>

                <div className="grid md:grid-cols-2 gap-4">
                  <Card>
                    <h3 className="text-lg font-bold text-neutral-900 dark:text-white mb-2">
                      Current Month
                    </h3>
                    <p className="text-3xl font-bold text-primary-600 dark:text-primary-400">
                      ${stats.monthlyUsage.toFixed(2)}
                    </p>
                    <p className="text-sm text-neutral-500 mt-2">
                      of $1,000.00 monthly limit
                    </p>
                  </Card>

                  <Card>
                    <h3 className="text-lg font-bold text-neutral-900 dark:text-white mb-2">
                      Available Credit
                    </h3>
                    <p className="text-3xl font-bold text-green-600 dark:text-green-400">
                      ${(1000 - stats.monthlyUsage).toFixed(2)}
                    </p>
                    <p className="text-sm text-neutral-500 mt-2">
                      remaining this month
                    </p>
                  </Card>
                </div>

                <Card>
                  <h2 className="text-lg font-bold text-neutral-900 dark:text-white mb-4">
                    Usage Breakdown
                  </h2>
                  <VStack spacing="md">
                    {['gpt-4', 'gpt-3.5-turbo', 'claude-3', 'gemini-pro'].map((model, i) => (
                      <div key={model} className="pb-3 border-b border-neutral-200 dark:border-neutral-700 last:border-b-0">
                        <div className="flex items-center justify-between mb-2">
                          <span className="font-medium text-neutral-900 dark:text-white">
                            {model}
                          </span>
                          <span className="text-sm font-medium text-neutral-600 dark:text-neutral-400">
                            ${(i * 150 + 200).toFixed(2)}
                          </span>
                        </div>
                        <div className="w-full h-2 bg-neutral-200 dark:bg-neutral-700 rounded-full overflow-hidden">
                          <div
                            className="h-full bg-gradient-to-r from-primary-400 to-primary-600"
                            style={{ width: `${(i * 25 + 40) % 90}%` }}
                          />
                        </div>
                      </div>
                    ))}
                  </VStack>
                </Card>

                <Button fullWidth>
                  Upgrade Plan
                </Button>
              </VStack>
            )}
          </div>
        </div>
      </Container>
    </div>
  );
}

