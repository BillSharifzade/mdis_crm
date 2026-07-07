import { useEffect, useState, useMemo } from 'react';
import { SourcesContext } from './sourcesContextValue.js';
import { api, sourceLabel } from '../services/api.js';
import { SOURCES as DEFAULT_SOURCES } from '../data/crmData.js';

// Живой список источников обращения из БД (управляется админом в Настройках).
// Используется выпадающими списками в формах лида и фильтрах, чтобы новые
// источники сразу появлялись без правки кода.
export function SourcesProvider({ children, enabled }) {
    const [sources, setSources] = useState([]);

    useEffect(() => {
        if (!api.useApi || !enabled) return;
        let cancelled = false;
        api.getSources()
            .then(data => { if (!cancelled) setSources(Array.isArray(data) ? data : []); })
            .catch(err => console.warn('sources load failed', err));
        return () => { cancelled = true; };
    }, [enabled]);

    const value = useMemo(() => {
        // Сворачиваем к человекочитаемой подписи и дедупим (site/organic → «Сайт»).
        const seen = new Map();
        sources.forEach(s => {
            const label = sourceLabel(s.name);
            if (!seen.has(label)) seen.set(label, { id: s.id, name: s.name, label });
        });
        const options = seen.size > 0
            ? Array.from(seen.values())
            : DEFAULT_SOURCES.map(label => ({ id: null, name: label, label }));
        return { sources, options };
    }, [sources]);

    return <SourcesContext.Provider value={value}>{children}</SourcesContext.Provider>;
}
