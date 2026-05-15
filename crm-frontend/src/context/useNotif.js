import { useContext } from 'react';
import { NotifContext } from './notifContextValue.js';

export function useNotif() {
    const ctx = useContext(NotifContext);
    if (!ctx) throw new Error('useNotif must be used inside NotifProvider');
    return ctx;
}
