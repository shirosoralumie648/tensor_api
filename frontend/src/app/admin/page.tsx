'use client';

import React, { useState } from 'react';
import { Button, Card, Input, Select, Table, StatCard, Alert } from '@/components/ui';
import { Container, VStack, HStack, Navbar, NavbarContent, NavbarBrand, NavbarMenu, NavbarItem, NavbarActions } from '@/components/layout';
import { Users, BarChart3, Settings, LogOut, Trash2, Edit, Search } from 'lucide-react';
import Link from 'next/link';
import { useTranslation } from '@/hooks/useTranslation';

type AdminTab = 'dashboard' | 'users' | 'channels' | 'settings';

interface User {
  id: string;
  name: string;
  email: string;
  role: 'user' | 'developer' | 'admin';
  status: 'active' | 'inactive' | 'suspended';
  created: Date;
  requests: number;
}

interface Channel {
  id: string;
  name: string;
  type: string;
  status: 'active' | 'inactive';
  models: string[];
  requests: number;
}

interface AdminStats {
  totalUsers: number;
  activeUsers: number;
  totalRequests: number;
  totalRevenue: number;
  systemHealth: number;
}

export default function AdminPage() {
  const { t, language, setLanguage } = useTranslation();
  const [activeTab, setActiveTab] = useState<AdminTab>('dashboard');
  const [searchQuery, setSearchQuery] = useState('');
  const [filterRole, setFilterRole] = useState<'all' | 'user' | 'developer' | 'admin'>('all');
  const [filterStatus, setFilterStatus] = useState<'all' | 'active' | 'inactive' | 'suspended'>('all');

  const [users, setUsers] = useState<User[]>([
    {
      id: '1',
      name: 'John Doe',
      email: 'john@example.com',
      role: 'user',
      status: 'active',
      created: new Date(Date.now() - 30 * 24 * 3600000),
      requests: 1234,
    },
    {
      id: '2',
      name: 'Jane Smith',
      email: 'jane@example.com',
      role: 'developer',
      status: 'active',
      created: new Date(Date.now() - 60 * 24 * 3600000),
      requests: 5678,
    },
    {
      id: '3',
      name: 'Bob Wilson',
      email: 'bob@example.com',
      role: 'user',
      status: 'suspended',
      created: new Date(Date.now() - 90 * 24 * 3600000),
      requests: 234,
    },
  ]);

  const [channels, setChannels] = useState<Channel[]>([
    {
      id: '1',
      name: 'OpenAI',
      type: 'LLM',
      status: 'active',
      models: ['gpt-4', 'gpt-3.5-turbo'],
      requests: 45234,
    },
    {
      id: '2',
      name: 'Anthropic',
      type: 'LLM',
      status: 'active',
      models: ['claude-3-opus', 'claude-3-sonnet'],
      requests: 23421,
    },
    {
      id: '3',
      name: 'Google',
      type: 'LLM',
      status: 'inactive',
      models: ['gemini-pro', 'gemini-pro-vision'],
      requests: 8923,
    },
  ]);

  const stats: AdminStats = {
    totalUsers: 1250,
    activeUsers: 892,
    totalRequests: 125000,
    totalRevenue: 28456.78,
    systemHealth: 99.8,
  };

  const filteredUsers = users.filter(user => {
    const matchesSearch = user.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         user.email.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesRole = filterRole === 'all' || user.role === filterRole;
    const matchesStatus = filterStatus === 'all' || user.status === filterStatus;
    return matchesSearch && matchesRole && matchesStatus;
  });

  const deleteUser = (id: string) => {
    setUsers(users.filter(u => u.id !== id));
  };

  const deleteChannel = (id: string) => {
    setChannels(channels.filter(c => c.id !== id));
  };

  const getRoleColor = (role: string) => {
    switch (role) {
      case 'admin':
        return 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400';
      case 'developer':
        return 'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400';
      default:
        return 'bg-gray-100 dark:bg-gray-900/30 text-gray-700 dark:text-gray-400';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400';
      case 'inactive':
        return 'bg-gray-100 dark:bg-gray-900/30 text-gray-700 dark:text-gray-400';
      case 'suspended':
        return 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400';
      default:
        return 'bg-gray-100 dark:bg-gray-900/30 text-gray-700 dark:text-gray-400';
    }
  };

  return (
    <div className="min-h-screen bg-neutral-50 dark:bg-neutral-900">
      {/* Navbar */}
      <Navbar>
        <NavbarContent>
          <NavbarBrand>
            <Link href="/">
              <div className="text-2xl font-bold bg-gradient-to-r from-primary-500 to-blue-500 bg-clip-text text-transparent cursor-pointer">
                Oblivious Admin
              </div>
            </Link>
          </NavbarBrand>
          <NavbarMenu>
            <NavbarItem href="/chat">{t('nav.chat')}</NavbarItem>
            <NavbarItem href="/developer">Developer</NavbarItem>
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
                  onClick={() => setActiveTab('users')}
                  className={`w-full text-left px-4 py-3 rounded-lg font-medium transition ${
                    activeTab === 'users'
                      ? 'bg-primary-500 text-white'
                      : 'text-neutral-600 dark:text-neutral-400 hover:bg-neutral-100 dark:hover:bg-neutral-800'
                  }`}
                >
                  <Users className="inline mr-2" size={18} />
                  Users
                </button>
                <button
                  onClick={() => setActiveTab('channels')}
                  className={`w-full text-left px-4 py-3 rounded-lg font-medium transition ${
                    activeTab === 'channels'
                      ? 'bg-primary-500 text-white'
                      : 'text-neutral-600 dark:text-neutral-400 hover:bg-neutral-100 dark:hover:bg-neutral-800'
                  }`}
                >
                  <BarChart3 className="inline mr-2" size={18} />
                  Channels
                </button>
                <button
                  onClick={() => setActiveTab('settings')}
                  className={`w-full text-left px-4 py-3 rounded-lg font-medium transition ${
                    activeTab === 'settings'
                      ? 'bg-primary-500 text-white'
                      : 'text-neutral-600 dark:text-neutral-400 hover:bg-neutral-100 dark:hover:bg-neutral-800'
                  }`}
                >
                  <Settings className="inline mr-2" size={18} />
                  Settings
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
                  Admin Dashboard
                </h1>

                {/* Stats Grid */}
                <div className="grid md:grid-cols-2 gap-4">
                  <StatCard
                    label="Total Users"
                    value={stats.totalUsers}
                    icon={<Users size={24} />}
                    trend={{ value: 5, direction: 'up' }}
                    color="primary"
                  />
                  <StatCard
                    label="Active Users"
                    value={stats.activeUsers}
                    icon={<Users size={24} />}
                    trend={{ value: 3, direction: 'up' }}
                    color="success"
                  />
                  <StatCard
                    label="Total Requests"
                    value={`${(stats.totalRequests / 1000).toFixed(0)}K`}
                    icon={<BarChart3 size={24} />}
                    trend={{ value: 12, direction: 'up' }}
                    color="info"
                  />
                  <StatCard
                    label="Revenue (This Month)"
                    value={`$${stats.totalRevenue.toFixed(2)}`}
                    icon={<BarChart3 size={24} />}
                    trend={{ value: 8, direction: 'up' }}
                    color="success"
                  />
                </div>

                {/* System Health */}
                <Card>
                  <h2 className="text-xl font-bold text-neutral-900 dark:text-white mb-4">
                    System Health
                  </h2>
                  <VStack spacing="md">
                    <div className="flex items-center justify-between">
                      <span className="font-medium text-neutral-900 dark:text-white">
                        Uptime
                      </span>
                      <span className="text-2xl font-bold text-green-600 dark:text-green-400">
                        {stats.systemHealth.toFixed(2)}%
                      </span>
                    </div>
                    <div className="w-full h-3 bg-neutral-200 dark:bg-neutral-700 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-gradient-to-r from-green-400 to-green-600"
                        style={{ width: `${stats.systemHealth}%` }}
                      />
                    </div>
                  </VStack>
                </Card>

                {/* Recent Logs */}
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
                              Event #{i}
                            </p>
                            <p className="text-xs text-neutral-500">
                              User action logged • {new Date(Date.now() - i * 3600000).toLocaleTimeString()}
                            </p>
                          </div>
                          <span className="text-xs font-medium px-2 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 rounded">
                            Info
                          </span>
                        </div>
                      </div>
                    ))}
                  </VStack>
                </Card>
              </VStack>
            )}

            {/* Users Tab */}
            {activeTab === 'users' && (
              <VStack spacing="lg">
                <div className="flex items-center justify-between">
                  <h1 className="text-3xl font-bold text-neutral-900 dark:text-white">
                    User Management
                  </h1>
                  <Button>
                    Create User
                  </Button>
                </div>

                {/* Filters */}
                <Card>
                  <div className="grid md:grid-cols-3 gap-4">
                    <Input
                      placeholder="Search by name or email..."
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      containerClassName="w-full"
                    />
                    <Select
                      options={[
                        { value: 'all', label: 'All Roles' },
                        { value: 'user', label: 'User' },
                        { value: 'developer', label: 'Developer' },
                        { value: 'admin', label: 'Admin' },
                      ]}
                      value={filterRole}
                      onChange={(e) => setFilterRole(e.target.value as any)}
                    />
                    <Select
                      options={[
                        { value: 'all', label: 'All Status' },
                        { value: 'active', label: 'Active' },
                        { value: 'inactive', label: 'Inactive' },
                        { value: 'suspended', label: 'Suspended' },
                      ]}
                      value={filterStatus}
                      onChange={(e) => setFilterStatus(e.target.value as any)}
                    />
                  </div>
                </Card>

                {/* Users Table */}
                <Table
                  columns={[
                    { key: 'name', label: 'Name', width: '20%' },
                    { key: 'email', label: 'Email', width: '25%' },
                    {
                      key: 'role',
                      label: 'Role',
                      width: '15%',
                      render: (value) => (
                        <span className={`text-xs font-medium px-2 py-1 rounded ${getRoleColor(value)}`}>
                          {value.toUpperCase()}
                        </span>
                      ),
                    },
                    {
                      key: 'status',
                      label: 'Status',
                      width: '15%',
                      render: (value) => (
                        <span className={`text-xs font-medium px-2 py-1 rounded ${getStatusColor(value)}`}>
                          {value.toUpperCase()}
                        </span>
                      ),
                    },
                    { key: 'requests', label: 'Requests', width: '10%', align: 'right' },
                    {
                      key: 'id',
                      label: 'Actions',
                      width: '15%',
                      align: 'right',
                      render: (value) => (
                        <HStack spacing="sm">
                          <button className="p-1 hover:bg-neutral-100 dark:hover:bg-neutral-700 rounded transition">
                            <Edit size={16} className="text-neutral-600 dark:text-neutral-400" />
                          </button>
                          <button
                            onClick={() => deleteUser(value)}
                            className="p-1 hover:bg-red-100 dark:hover:bg-red-900/30 rounded transition"
                          >
                            <Trash2 size={16} className="text-red-600 dark:text-red-400" />
                          </button>
                        </HStack>
                      ),
                    },
                  ]}
                  data={filteredUsers}
                  striped={true}
                  hoverable={true}
                />
              </VStack>
            )}

            {/* Channels Tab */}
            {activeTab === 'channels' && (
              <VStack spacing="lg">
                <div className="flex items-center justify-between">
                  <h1 className="text-3xl font-bold text-neutral-900 dark:text-white">
                    Channel Management
                  </h1>
                  <Button>
                    Add Channel
                  </Button>
                </div>

                {/* Channels List */}
                <div className="space-y-4">
                  {channels.map(channel => (
                    <Card key={channel.id}>
                      <div className="flex items-center justify-between">
                        <div>
                          <h3 className="text-lg font-bold text-neutral-900 dark:text-white">
                            {channel.name}
                          </h3>
                          <p className="text-sm text-neutral-600 dark:text-neutral-400 mt-1">
                            Type: {channel.type} • Requests: {channel.requests.toLocaleString()}
                          </p>
                          <div className="flex gap-2 mt-2">
                            {channel.models.map(model => (
                              <span
                                key={model}
                                className="text-xs px-2 py-1 bg-neutral-100 dark:bg-neutral-700 text-neutral-700 dark:text-neutral-300 rounded"
                              >
                                {model}
                              </span>
                            ))}
                          </div>
                        </div>
                        <div className="flex items-center gap-2">
                          <span
                            className={`text-xs font-medium px-3 py-1 rounded ${
                              channel.status === 'active'
                                ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400'
                                : 'bg-gray-100 dark:bg-gray-900/30 text-gray-700 dark:text-gray-400'
                            }`}
                          >
                            {channel.status.toUpperCase()}
                          </span>
                          <button className="p-2 hover:bg-neutral-100 dark:hover:bg-neutral-700 rounded transition">
                            <Edit size={18} className="text-neutral-600 dark:text-neutral-400" />
                          </button>
                          <button
                            onClick={() => deleteChannel(channel.id)}
                            className="p-2 hover:bg-red-100 dark:hover:bg-red-900/30 rounded transition"
                          >
                            <Trash2 size={18} className="text-red-600 dark:text-red-400" />
                          </button>
                        </div>
                      </div>
                    </Card>
                  ))}
                </div>
              </VStack>
            )}

            {/* Settings Tab */}
            {activeTab === 'settings' && (
              <VStack spacing="lg">
                <h1 className="text-3xl font-bold text-neutral-900 dark:text-white">
                  System Settings
                </h1>

                <Alert variant="info">
                  Modify system settings carefully. Changes may affect all users.
                </Alert>

                {/* General Settings */}
                <Card>
                  <h2 className="text-xl font-bold text-neutral-900 dark:text-white mb-6">
                    General Settings
                  </h2>
                  <VStack spacing="md">
                    <Input
                      label="System Name"
                      placeholder="Oblivious"
                      defaultValue="Oblivious"
                    />
                    <Input
                      label="Support Email"
                      placeholder="support@example.com"
                      defaultValue="support@oblivious.com"
                    />
                    <Input
                      label="Max Request Rate (req/min)"
                      type="number"
                      placeholder="1000"
                      defaultValue="1000"
                    />
                  </VStack>
                </Card>

                {/* Security Settings */}
                <Card>
                  <h2 className="text-xl font-bold text-neutral-900 dark:text-white mb-6">
                    Security Settings
                  </h2>
                  <VStack spacing="md">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="font-medium text-neutral-900 dark:text-white">
                          Two-Factor Authentication
                        </p>
                        <p className="text-sm text-neutral-600 dark:text-neutral-400 mt-1">
                          Require 2FA for all admin accounts
                        </p>
                      </div>
                      <input type="checkbox" defaultChecked className="w-4 h-4" />
                    </div>
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="font-medium text-neutral-900 dark:text-white">
                          API Rate Limiting
                        </p>
                        <p className="text-sm text-neutral-600 dark:text-neutral-400 mt-1">
                          Enable per-user rate limiting
                        </p>
                      </div>
                      <input type="checkbox" defaultChecked className="w-4 h-4" />
                    </div>
                  </VStack>
                </Card>

                {/* Maintenance */}
                <Card>
                  <h2 className="text-xl font-bold text-neutral-900 dark:text-white mb-6">
                    Maintenance
                  </h2>
                  <VStack spacing="md">
                    <Button fullWidth variant="outline">
                      Clear Cache
                    </Button>
                    <Button fullWidth variant="outline">
                      Export Logs
                    </Button>
                    <Button fullWidth variant="danger">
                      Restart System
                    </Button>
                  </VStack>
                </Card>

                {/* Save */}
                <Button fullWidth>
                  Save Settings
                </Button>
              </VStack>
            )}
          </div>
        </div>
      </Container>
    </div>
  );
}

