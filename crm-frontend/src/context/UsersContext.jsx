import { useEffect, useState, useMemo } from 'react';
import { UsersContext } from './usersContextValue.js';
import { api } from '../services/api.js';

const COLORS = ['#6366f1', '#10b981', '#f59e0b', '#a855f7', '#ec4899', '#06b6d4', '#ef4444'];

function decorate(user, idx) {
    const initials = (user.name || 'NN').split(/\s+/).slice(0, 2).map(w => w[0]).join('').toUpperCase();
    return { ...user, initials, color: COLORS[idx % COLORS.length] };
}

export function UsersProvider({ children, enabled }) {
    const [users, setUsers] = useState([]);
    const [reloadKey, setReloadKey] = useState(0);

    useEffect(() => {
        if (!api.useApi || !enabled) return;
        let cancelled = false;
        api.getManagers()
            .then(data => {
                if (cancelled) return;
                const arr = Array.isArray(data) ? data : [];
                setUsers(arr.map(decorate));
            })
            .catch(err => console.error('users load failed', err));
        return () => { cancelled = true; };
    }, [enabled, reloadKey]);

    const value = useMemo(() => {
        const byId = new Map(users.map(u => [u.id, u]));
        return { users, byId, reload: () => setReloadKey(k => k + 1) };
    }, [users]);

    return <UsersContext.Provider value={value}>{children}</UsersContext.Provider>;
}
