import { useEffect, useState, useMemo } from 'react';
import { ProgramsContext } from './programsContextValue.js';
import { api } from '../services/api.js';

// Show Russian programs first (since the bot/UI is Russian), English seed at the bottom.
const RU_RX = /^[Ѐ-ӿ]/;

export function ProgramsProvider({ children, enabled }) {
    const [programs, setPrograms] = useState([]);

    useEffect(() => {
        if (!api.useApi || !enabled) return;
        let cancelled = false;
        api.getPrograms()
            .then(data => {
                if (cancelled) return;
                const sorted = [...data].sort((a, b) => {
                    const aRu = RU_RX.test(a.name);
                    const bRu = RU_RX.test(b.name);
                    if (aRu !== bRu) return aRu ? -1 : 1;
                    return a.id - b.id;
                });
                setPrograms(sorted);
            })
            .catch(err => console.error('programs load failed', err));
        return () => { cancelled = true; };
    }, [enabled]);

    const value = useMemo(() => {
        const byId = new Map(programs.map(p => [p.id, p]));
        const byName = new Map(programs.map(p => [p.name, p]));
        return { programs, byId, byName };
    }, [programs]);

    return <ProgramsContext.Provider value={value}>{children}</ProgramsContext.Provider>;
}
