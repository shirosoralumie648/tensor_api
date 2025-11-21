/**
 * 翻译 Hook
 * 提供多语言支持
 */

import { useCallback, useMemo, useState, useEffect } from 'react';
import { Language, translations, getPreferredLanguage, setPreferredLanguage } from '@/i18n/config';

export interface UseTranslationReturn {
  t: (key: string) => string;
  language: Language;
  setLanguage: (language: Language) => void;
}

/**
 * 从嵌套对象中获取值
 */
function getNestedValue(obj: any, path: string): string {
  const keys = path.split('.');
  let value = obj;

  for (const key of keys) {
    if (value && typeof value === 'object' && key in value) {
      value = value[key];
    } else {
      return path; // 返回原始 key 作为后备值
    }
  }

  return typeof value === 'string' ? value : path;
}

/**
 * 使用翻译
 */
export function useTranslation(): UseTranslationReturn {
  const [language, setLanguageState] = useState<Language>(() => getPreferredLanguage());

  // 在客户端初始化时读取偏好语言
  useEffect(() => {
    const preferredLang = getPreferredLanguage();
    if (preferredLang !== language) {
      setLanguageState(preferredLang);
    }
  }, []);

  const t = useCallback(
    (key: string): string => {
      const translationObj = translations[language] as Record<string, any>;

      if (!translationObj) {
        return key;
      }

      return getNestedValue(translationObj, key);
    },
    [language]
  );

  const setLanguage = useCallback((lang: Language) => {
    setPreferredLanguage(lang);
    setLanguageState(lang);
  }, []);

  return useMemo(
    () => ({
      t,
      language,
      setLanguage,
    }),
    [t, language, setLanguage]
  );
}

/**
 * 简单翻译函数 (用于组件外)
 */
export function translate(key: string, language?: Language): string {
  const lang = language || getPreferredLanguage();
  const translationObj = translations[lang];

  if (!translationObj) {
    return key;
  }

  return getNestedValue(translationObj, key);
}

