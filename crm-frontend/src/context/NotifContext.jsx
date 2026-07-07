import { useState, useEffect, useCallback } from 'react';
import { NotifContext } from './notifContextValue.js';

const STORAGE_KEY = 'crm_notifications';
const MAX_NOTIFS = 50;

function loadInitial() {
    try {
        const raw = localStorage.getItem(STORAGE_KEY);
        if (!raw) return [];
        const parsed = JSON.parse(raw);
        if (!Array.isArray(parsed)) return [];
        return parsed.slice(0, MAX_NOTIFS);
    } catch { return []; }
}

function save(notifs) {
    try { localStorage.setItem(STORAGE_KEY, JSON.stringify(notifs.slice(0, MAX_NOTIFS))); }
    catch { /* quota — ignore */ }
}

export function NotifProvider({ children }) {
    const [notifs, setNotifs] = useState(loadInitial);

    useEffect(() => { save(notifs); }, [notifs]);

    const push = useCallback((notif) => {
        const entry = {
            id: Date.now() + Math.random(),
            createdAt: new Date().toISOString(),
            unread: true,
            ...notif,
        };
        setNotifs(prev => {
            // Дедупликация напоминаний: не плодим одинаковые уведомления по
            // одному и тому же лиду (планировщик стреляет один раз, но SSE может
            // переподключиться).
            if (entry.dedupeKey) {
                const filtered = prev.filter(n => n.dedupeKey !== entry.dedupeKey);
                return [entry, ...filtered].slice(0, MAX_NOTIFS);
            }
            return [entry, ...prev].slice(0, MAX_NOTIFS);
        });
    }, []);

    const markAllRead = useCallback(() => {
        setNotifs(prev => prev.map(n => ({ ...n, unread: false })));
    }, []);

    // Удаление одного уведомления (кнопки «Выполнено» / «Закрыть»).
    const remove = useCallback((id) => {
        setNotifs(prev => prev.filter(n => n.id !== id));
    }, []);

    // Убрать все уведомления по лиду (например, когда напоминание закрыто).
    const removeByLead = useCallback((leadId) => {
        setNotifs(prev => prev.filter(n => !(n.leadId === leadId && n.type === 'reminder')));
    }, []);

    const clearAll = useCallback(() => {
        setNotifs([]);
    }, []);

    const unreadCount = notifs.filter(n => n.unread).length;

    return (
        <NotifContext.Provider value={{ notifs, push, markAllRead, remove, removeByLead, clearAll, unreadCount }}>
            {children}
        </NotifContext.Provider>
    );
}
