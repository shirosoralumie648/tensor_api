'use client';

import React, { useState, useRef, useEffect } from 'react';
import { Button, Input, Card, Spinner, FileUpload } from '@/components/ui';
import { HStack, VStack } from '@/components/layout';
import { Send, Plus, Menu, Settings, LogOut, Home, Globe } from 'lucide-react';
import Link from 'next/link';
import { useStreamingChat } from '@/hooks/useStreamingChat';
import { useTranslation } from '@/hooks/useTranslation';

interface Conversation {
  id: string;
  title: string;
  created: Date;
}

export default function ChatPage() {
  const { t, language, setLanguage } = useTranslation();
  const conversationId = 'conv-' + Date.now();
  const { messages, isLoading, error, sendMessage } = useStreamingChat({
    conversationId,
    onConnected: () => console.log(t('chat.connected')),
    onError: (err) => console.error(err.message),
  });

  const [input, setInput] = useState('');
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [showFileUpload, setShowFileUpload] = useState(false);
  const [conversations, setConversations] = useState<Conversation[]>([
    { id: '1', title: 'Welcome', created: new Date() },
  ]);

  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSendMessage = async () => {
    if (!input.trim()) return;
    sendMessage(input);
    setInput('');
  };

  return (
    <div className="h-screen flex bg-white dark:bg-neutral-900">
      {/* Sidebar */}
      <div
        className={`fixed md:relative z-40 w-64 h-screen border-r border-neutral-200 dark:border-neutral-700 bg-neutral-50 dark:bg-neutral-800 flex flex-col transition-transform duration-300 ${
          sidebarOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'
        }`}
      >
        <div className="p-4 border-b border-neutral-200 dark:border-neutral-700">
          <Button fullWidth className="gap-2">
            <Plus size={20} />
            New Chat
          </Button>
        </div>

        <div className="flex-1 overflow-y-auto p-4 space-y-2">
          <div className="text-xs font-semibold text-neutral-500 uppercase px-2 mb-3">
            Conversations
          </div>
          {conversations.map(conv => (
            <button
              key={conv.id}
              className="w-full text-left px-3 py-2 rounded-lg hover:bg-neutral-200 dark:hover:bg-neutral-700 transition text-sm text-neutral-700 dark:text-neutral-300"
            >
              {conv.title}
            </button>
          ))}
        </div>

        <div className="border-t border-neutral-200 dark:border-neutral-700 p-4 space-y-2">
          <button className="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-neutral-200 dark:hover:bg-neutral-700 transition text-sm text-neutral-700 dark:text-neutral-300">
            <Settings size={18} />
            {t('nav.settings')}
          </button>
          <div className="relative group">
            <button className="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-neutral-200 dark:hover:bg-neutral-700 transition text-sm text-neutral-700 dark:text-neutral-300">
              <Globe size={18} />
              {language.toUpperCase()}
            </button>
            <div className="hidden group-hover:block absolute bottom-full left-0 right-0 bg-white dark:bg-neutral-700 border border-neutral-200 dark:border-neutral-600 rounded-lg overflow-hidden z-50">
              {(['en', 'zh', 'ja', 'es', 'fr', 'de'] as const).map(lang => (
                <button
                  key={lang}
                  onClick={() => setLanguage(lang)}
                  className="w-full text-left px-3 py-2 hover:bg-neutral-100 dark:hover:bg-neutral-600 transition text-sm"
                >
                  {lang.toUpperCase()}
                </button>
              ))}
            </div>
          </div>
          <Link href="/">
            <button className="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-neutral-200 dark:hover:bg-neutral-700 transition text-sm text-neutral-700 dark:text-neutral-300">
              <Home size={18} />
              {t('nav.home')}
            </button>
          </Link>
          <button className="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-neutral-200 dark:hover:bg-neutral-700 transition text-sm text-red-600 dark:text-red-400">
            <LogOut size={18} />
            {t('nav.logout')}
          </button>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col">
        {/* Header */}
        <div className="h-16 border-b border-neutral-200 dark:border-neutral-700 px-4 flex items-center justify-between">
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="md:hidden p-2 hover:bg-neutral-100 dark:hover:bg-neutral-800 rounded-lg transition"
          >
            <Menu size={24} />
          </button>
          <div className="text-lg font-semibold text-neutral-900 dark:text-white">
            Chat
          </div>
          <div className="w-8" />
        </div>

        {/* Messages Area */}
        <div className="flex-1 overflow-y-auto p-4 md:p-6 space-y-4">
          {messages.map(message => (
            <div key={message.id} className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'}`}>
              <div
                className={`max-w-xs md:max-w-2xl px-4 py-3 rounded-2xl ${
                  message.role === 'user'
                    ? 'bg-primary-500 text-white rounded-br-none'
                    : 'bg-neutral-100 dark:bg-neutral-800 text-neutral-900 dark:text-white rounded-bl-none'
                }`}
              >
                <p className="text-sm leading-relaxed">{message.content}</p>
                <p className={`text-xs mt-1 ${message.role === 'user' ? 'text-primary-100' : 'text-neutral-500'}`}>
                  {message.timestamp.toLocaleTimeString([], {
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </p>
              </div>
            </div>
          ))}

          {isLoading && (
            <div className="flex justify-start">
              <div className="bg-neutral-100 dark:bg-neutral-800 px-4 py-3 rounded-2xl rounded-bl-none">
                <Spinner size="sm" />
              </div>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        {/* Input Area */}
        <div className="border-t border-neutral-200 dark:border-neutral-700 p-4 md:p-6 bg-white dark:bg-neutral-900">
          <div className="max-w-4xl mx-auto space-y-4">
            {/* File Upload Section */}
            {showFileUpload && (
              <FileUpload
                label={t('chat.attachFile')}
                maxFiles={5}
                maxSize={50 * 1024 * 1024}
                onUpload={(files: any[]) => {
                  console.log('Files uploaded:', files);
                  setShowFileUpload(false);
                }}
              />
            )}

            {/* Input Form */}
            <form
              onSubmit={(e) => {
                e.preventDefault();
                handleSendMessage();
              }}
              className="flex gap-3"
            >
              <Input
                placeholder={t('chat.placeholder')}
                value={input}
                onChange={(e) => setInput(e.target.value)}
                disabled={isLoading}
                containerClassName="flex-1"
              />
              <Button
                type="button"
                onClick={() => setShowFileUpload(!showFileUpload)}
                variant="outline"
                size="md"
                className="px-4"
                title={t('chat.attachFile')}
              >
                📎
              </Button>
              <Button
                type="submit"
                loading={isLoading}
                disabled={!input.trim() || isLoading}
                size="md"
                className="px-6"
              >
                <Send size={20} />
              </Button>
            </form>

            {error && (
              <div className="p-2 rounded-lg bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800">
                <p className="text-xs text-red-600 dark:text-red-400">
                  {t('common.error')}: {error.message}
                </p>
              </div>
            )}

            <p className="text-xs text-neutral-500 mt-2">
              {t('chat.sendHint')}
            </p>
          </div>
        </div>
      </div>

      {/* Overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/50 md:hidden z-30"
          onClick={() => setSidebarOpen(false)}
        />
      )}
    </div>
  );
}
