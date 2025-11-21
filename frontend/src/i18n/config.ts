/**
 * 国际化配置
 * 支持多语言翻译
 */

export type Language = 'en' | 'zh' | 'ja' | 'es' | 'fr' | 'de';

export interface TranslationKey {
  [key: string]: string | TranslationKey;
}

export type Translations = Record<Language, TranslationKey>;

export const SUPPORTED_LANGUAGES: { code: Language; name: string; nativeName: string }[] = [
  { code: 'en', name: 'English', nativeName: 'English' },
  { code: 'zh', name: 'Chinese', nativeName: '中文' },
  { code: 'ja', name: 'Japanese', nativeName: '日本語' },
  { code: 'es', name: 'Spanish', nativeName: 'Español' },
  { code: 'fr', name: 'French', nativeName: 'Français' },
  { code: 'de', name: 'German', nativeName: 'Deutsch' },
];

export const DEFAULT_LANGUAGE: Language = 'en';

// 翻译文本
export const translations: Translations = {
  en: {
    common: {
      loading: 'Loading...',
      error: 'Error',
      success: 'Success',
      cancel: 'Cancel',
      delete: 'Delete',
      save: 'Save',
      next: 'Next',
      previous: 'Previous',
    },
    nav: {
      home: 'Home',
      chat: 'Chat',
      developer: 'Developer',
      settings: 'Settings',
      logout: 'Logout',
    },
    chat: {
      newConversation: 'New Conversation',
      emptyMessage: 'No messages yet',
      placeholder: 'Type your message here...',
      sendButton: 'Send',
      connected: 'Connected to chat server',
      attachFile: 'Attach file',
      sendHint: 'Press Enter to send or click the send button',
    },
    user: {
      profile: 'Profile',
      quota: 'Quota',
      history: 'History',
      settings: 'Settings',
      fullName: 'Full Name',
      email: 'Email',
      bio: 'Bio',
      saveChanges: 'Save Changes',
    },
    settings: {
      language: 'Language',
      languageDesc: 'Choose your preferred language',
      notifications: 'Email Notifications',
      notificationsDesc: 'Receive email updates about your account',
      marketing: 'Marketing Emails',
      marketingDesc: 'Get updates about new features',
      deleteAccount: 'Delete Account',
    },
  },
  zh: {
    common: {
      loading: '加载中...',
      error: '错误',
      success: '成功',
      cancel: '取消',
      delete: '删除',
      save: '保存',
      next: '下一步',
      previous: '上一步',
    },
    nav: {
      home: '首页',
      chat: '对话',
      developer: '开发者',
      settings: '设置',
      logout: '登出',
    },
    chat: {
      newConversation: '新建对话',
      emptyMessage: '还没有消息',
      placeholder: '输入你的消息...',
      sendButton: '发送',
      connected: '已连接到聊天服务器',
      attachFile: '附加文件',
      sendHint: '按 Enter 发送或点击发送按钮',
    },
    user: {
      profile: '个人资料',
      quota: '配额',
      history: '历史',
      settings: '设置',
      fullName: '姓名',
      email: '邮箱',
      bio: '个人简介',
      saveChanges: '保存更改',
    },
    settings: {
      language: '语言',
      languageDesc: '选择您首选的语言',
      notifications: '邮件通知',
      notificationsDesc: '接收关于您的账户的邮件更新',
      marketing: '营销邮件',
      marketingDesc: '获取新功能的更新',
      deleteAccount: '删除账户',
    },
  },
  ja: {
    common: {
      loading: '読み込み中...',
      error: 'エラー',
      success: '成功',
      cancel: 'キャンセル',
      delete: '削除',
      save: '保存',
      next: '次へ',
      previous: '前へ',
    },
    nav: {
      home: 'ホーム',
      chat: 'チャット',
      developer: '開発者',
      settings: '設定',
      logout: 'ログアウト',
    },
    chat: {
      newConversation: '新しい会話',
      emptyMessage: 'メッセージはありません',
      placeholder: 'メッセージを入力してください...',
      sendButton: '送信',
      connected: 'チャットサーバーに接続しました',
      attachFile: 'ファイルを添付',
      sendHint: 'Enterキーを押すか、送信ボタンをクリックしてください',
    },
    user: {
      profile: 'プロフィール',
      quota: 'クォータ',
      history: '履歴',
      settings: '設定',
      fullName: 'フルネーム',
      email: 'メール',
      bio: 'バイオ',
      saveChanges: '変更を保存',
    },
    settings: {
      language: '言語',
      languageDesc: 'お好みの言語を選択してください',
      notifications: 'メール通知',
      notificationsDesc: 'アカウントに関するメール更新を受け取る',
      marketing: 'マーケティングメール',
      marketingDesc: '新機能の更新を取得',
      deleteAccount: 'アカウントを削除',
    },
  },
  es: {
    common: {
      loading: 'Cargando...',
      error: 'Error',
      success: 'Éxito',
      cancel: 'Cancelar',
      delete: 'Eliminar',
      save: 'Guardar',
      next: 'Siguiente',
      previous: 'Anterior',
    },
    nav: {
      home: 'Inicio',
      chat: 'Chat',
      developer: 'Desarrollador',
      settings: 'Configuración',
      logout: 'Cerrar sesión',
    },
    chat: {
      newConversation: 'Nueva conversación',
      emptyMessage: 'Sin mensajes',
      placeholder: 'Escribe tu mensaje aquí...',
      sendButton: 'Enviar',
      connected: 'Conectado al servidor de chat',
      attachFile: 'Adjuntar archivo',
      sendHint: 'Presiona Enter para enviar o haz clic en el botón enviar',
    },
    user: {
      profile: 'Perfil',
      quota: 'Cuota',
      history: 'Historial',
      settings: 'Configuración',
      fullName: 'Nombre completo',
      email: 'Correo electrónico',
      bio: 'Biografía',
      saveChanges: 'Guardar cambios',
    },
    settings: {
      language: 'Idioma',
      languageDesc: 'Elija su idioma preferido',
      notifications: 'Notificaciones por correo',
      notificationsDesc: 'Reciba actualizaciones por correo sobre su cuenta',
      marketing: 'Correos de marketing',
      marketingDesc: 'Obtenga actualizaciones sobre nuevas funciones',
      deleteAccount: 'Eliminar cuenta',
    },
  },
  fr: {
    common: {
      loading: 'Chargement...',
      error: 'Erreur',
      success: 'Succès',
      cancel: 'Annuler',
      delete: 'Supprimer',
      save: 'Enregistrer',
      next: 'Suivant',
      previous: 'Précédent',
    },
    nav: {
      home: 'Accueil',
      chat: 'Chat',
      developer: 'Développeur',
      settings: 'Paramètres',
      logout: 'Déconnexion',
    },
    chat: {
      newConversation: 'Nouvelle conversation',
      emptyMessage: 'Pas de messages',
      placeholder: 'Tapez votre message ici...',
      sendButton: 'Envoyer',
      connected: 'Connecté au serveur de chat',
      attachFile: 'Joindre un fichier',
      sendHint: 'Appuyez sur Entrée pour envoyer ou cliquez sur le bouton envoyer',
    },
    user: {
      profile: 'Profil',
      quota: 'Quota',
      history: 'Historique',
      settings: 'Paramètres',
      fullName: 'Nom complet',
      email: 'E-mail',
      bio: 'Biographie',
      saveChanges: 'Enregistrer les modifications',
    },
    settings: {
      language: 'Langue',
      languageDesc: 'Choisissez votre langue préférée',
      notifications: 'Notifications par courrier électronique',
      notificationsDesc: 'Recevoir des mises à jour par courrier électronique sur votre compte',
      marketing: 'Courriers de marketing',
      marketingDesc: 'Obtenir des mises à jour sur les nouvelles fonctionnalités',
      deleteAccount: 'Supprimer le compte',
    },
  },
  de: {
    common: {
      loading: 'Wird geladen...',
      error: 'Fehler',
      success: 'Erfolg',
      cancel: 'Abbrechen',
      delete: 'Löschen',
      save: 'Speichern',
      next: 'Weiter',
      previous: 'Zurück',
    },
    nav: {
      home: 'Startseite',
      chat: 'Chat',
      developer: 'Entwickler',
      settings: 'Einstellungen',
      logout: 'Abmelden',
    },
    chat: {
      newConversation: 'Neue Unterhaltung',
      emptyMessage: 'Keine Nachrichten',
      placeholder: 'Geben Sie Ihre Nachricht hier ein...',
      sendButton: 'Senden',
      connected: 'Mit Chat-Server verbunden',
      attachFile: 'Datei anhängen',
      sendHint: 'Drücken Sie Enter zum Senden oder klicken Sie auf die Schaltfläche Senden',
    },
    user: {
      profile: 'Profil',
      quota: 'Kontingent',
      history: 'Verlauf',
      settings: 'Einstellungen',
      fullName: 'Vollständiger Name',
      email: 'E-Mail',
      bio: 'Biographie',
      saveChanges: 'Änderungen speichern',
    },
    settings: {
      language: 'Sprache',
      languageDesc: 'Wählen Sie Ihre bevorzugte Sprache',
      notifications: 'E-Mail-Benachrichtigungen',
      notificationsDesc: 'Erhalten Sie E-Mail-Aktualisierungen zu Ihrem Konto',
      marketing: 'Marketing-E-Mails',
      marketingDesc: 'Erhalten Sie Updates zu neuen Funktionen',
      deleteAccount: 'Konto löschen',
    },
  },
};

/**
 * 获取用户偏好的语言
 */
export function getPreferredLanguage(): Language {
  if (typeof window === 'undefined') {
    return DEFAULT_LANGUAGE;
  }

  // 从本地存储获取
  const saved = localStorage.getItem('preferred-language') as Language | null;
  if (saved && SUPPORTED_LANGUAGES.some(l => l.code === saved)) {
    return saved;
  }

  // 从浏览器语言获取
  const browserLang = navigator.language.split('-')[0];
  const matched = SUPPORTED_LANGUAGES.find(l => l.code === browserLang);
  if (matched) {
    return matched.code;
  }

  return DEFAULT_LANGUAGE;
}

/**
 * 保存用户语言偏好
 */
export function setPreferredLanguage(language: Language): void {
  if (typeof window !== 'undefined') {
    localStorage.setItem('preferred-language', language);
  }
}

