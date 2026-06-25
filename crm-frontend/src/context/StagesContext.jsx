import { useEffect, useState, useMemo, useCallback } from 'react';
import { StagesContext } from './stagesContextValue.js';
import { api } from '../services/api.js';
import { STATUSES as DEFAULT_STAGES } from '../data/crmData.js';
import { STATUS_ID_TO_KEY } from '../services/api.js';

// Цвета по умолчанию для известных ключей — чтобы UI не выглядел серым,
// если сервер ещё не вернул свои.
const COLOR_BY_KEY = {
    inquiry: '#14b8a6',
    new: '#6366f1',
    consultation: '#06b6d4',
    documents: '#f59e0b',
    exams: '#a855f7',
    payment: '#ec4899',
    enrolled: '#10b981',
    lost: '#ef4444',
};

// Эвристический ключ из имени этапа — нужен, если админ переименовал стандартный.
function deriveKey(name) {
    const n = (name || '').toLowerCase();
    if (n.includes('обращ')) return 'inquiry';
    if (n.includes('нов')) return 'new';
    if (n.includes('консуль')) return 'consultation';
    if (n.includes('документ')) return 'documents';
    if (n.includes('eet') || n.includes('экзам') || n.includes('тест')) return 'exams';
    if (n.includes('оплат') || n.includes('договор')) return 'payment';
    if (n.includes('зачисл')) return 'enrolled';
    if (n.includes('отказ') || n.includes('проигр')) return 'lost';
    return 'st_' + Math.random().toString(36).slice(2, 8);
}

function classForKey(k) {
    return ({
        inquiry: 's-inquiry', new: 's-new', consultation: 's-consult', documents: 's-docs',
        exams: 's-exam', payment: 's-payment', enrolled: 's-enrolled', lost: 's-lost',
    })[k] || 's-new';
}

export function StagesProvider({ children, enabled }) {
    const [stages, setStages] = useState(DEFAULT_STAGES);
    const [reloadKey, setReloadKey] = useState(0);

    const load = useCallback(async () => {
        if (!api.useApi || !enabled) return;
        try {
            const data = await api.getStages();
            if (!Array.isArray(data) || data.length === 0) return;
            // Бэк отдаёт {id, name, order, is_final, is_active}
            const mapped = data
                .filter(s => s.is_active !== false)
                .map(s => {
                    const key = STATUS_ID_TO_KEY[s.id] || deriveKey(s.name);
                    return {
                        id: s.id,
                        key,
                        label: s.name,
                        order: s.order,
                        isFinal: !!s.is_final,
                        color: COLOR_BY_KEY[key] || '#64748b',
                        cls: classForKey(key),
                    };
                })
                .sort((a, b) => a.order - b.order);
            setStages(mapped);
        } catch (err) {
            console.warn('stages load failed', err);
        }
    }, [enabled]);

    useEffect(() => { load(); }, [load, reloadKey]);

    const value = useMemo(() => {
        const byId = new Map(stages.map(s => [s.id, s]));
        const byKey = new Map(stages.map(s => [s.key, s]));
        return { stages, byId, byKey, reload: () => setReloadKey(k => k + 1) };
    }, [stages]);

    return <StagesContext.Provider value={value}>{children}</StagesContext.Provider>;
}
