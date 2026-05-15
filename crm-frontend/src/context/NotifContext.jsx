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
        setNotifs(prev => [entry, ...prev].slice(0, MAX_NOTIFS));
    }, []);

    const markAllRead = useCallback(() => {
        setNotifs(prev => prev.map(n => ({ ...n, unread: false })));
    }, []);

    const clearAll = useCallback(() => {
        setNotifs([]);
    }, []);

    const unreadCount = notifs.filter(n => n.unread).length;

    return (
        <NotifContext.Provider value={{ notifs, push, markAllRead, clearAll, unreadCount }}>
            {children}
        </NotifContext.Provider>
    );
}
