import { useState, useMemo } from 'react';
import { X, GitMerge, Search } from 'lucide-react';
import { api } from '../services/api';
import { getStatusObj } from '../data/crmData';

export default function MergeLeadsModal({ leads, onClose, onDone, showToast }) {
    const [primaryId, setPrimaryId] = useState(null);
    const [duplicateId, setDuplicateId] = useState(null);
    const [query, setQuery] = useState('');
    const [busy, setBusy] = useState(false);

    const filtered = useMemo(() => {
        const q = query.trim().toLowerCase();
        if (!q) return leads.slice(0, 40);
        return leads.filter(l =>
            (l.name || '').toLowerCase().includes(q) ||
            (l.phone || '').includes(q) ||
            (l.email || '').toLowerCase().includes(q)
        ).slice(0, 40);
    }, [leads, query]);

    const pick = (id) => {
        if (primaryId == null) {
            setPrimaryId(id);
        } else if (id === primaryId) {
            setPrimaryId(null);
        } else if (duplicateId === id) {
            setDuplicateId(null);
        } else {
            setDuplicateId(id);
        }
    };

    const handleMerge = async () => {
        if (!primaryId || !duplicateId) {
            showToast('Выберите основную карточку и дубликат', 'warning');
            return;
        }
        if (!api.useApi) {
            showToast('Слияние доступно только при подключённом API', 'warning');
            return;
        }
        setBusy(true);
        try {
            await api.mergeLeads(primaryId, duplicateId);
            onDone();
        } catch (err) {
            console.error(err);
            showToast('Ошибка слияния: ' + err.message, 'error');
        } finally {
            setBusy(false);
        }
    };

    return (
        <div className="modal-overlay open" onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}>
            <div className="modal" style={{ maxWidth: 640 }}>
                <div className="modal-header">
                    <h2>Объединение дублей</h2>
                    <button className="modal-close" onClick={onClose}><X size={14} /></button>
                </div>
                <div className="modal-body">
                    <p style={{ color: 'var(--text-secondary)', fontSize: 13, marginBottom: 12 }}>
                        Выберите <b>основную</b> карточку (синяя), затем <b>дубликат</b> (красный). Вся история взаимодействий из дубликата будет перенесена в основную карточку, а дубликат — удалён.
                    </p>

                    <div className="search-box search-inline" style={{ marginBottom: 12 }}>
                        <Search size={15} />
                        <input
                            type="text"
                            placeholder="Поиск по имени, телефону, email..."
                            value={query}
                            onChange={e => setQuery(e.target.value)}
                        />
                    </div>

                    <div style={{ maxHeight: 360, overflowY: 'auto', border: '1px solid var(--glass-border)', borderRadius: 8 }}>
                        {filtered.length === 0 && (
                            <div style={{ padding: 16, textAlign: 'center', color: 'var(--text-muted)', fontSize: 13 }}>
                                Ничего не найдено
                            </div>
                        )}
                        {filtered.map(l => {
                            const isPrimary = l.id === primaryId;
                            const isDuplicate = l.id === duplicateId;
                            const st = getStatusObj(l.status);
                            let border = '1px solid var(--glass-border)';
                            let bg = 'transparent';
                            if (isPrimary) { border = '1px solid #3b82f6'; bg = 'rgba(59,130,246,0.1)'; }
                            else if (isDuplicate) { border = '1px solid #ef4444'; bg = 'rgba(239,68,68,0.1)'; }
                            return (
                                <div
                                    key={l.id}
                                    onClick={() => pick(l.id)}
                                    style={{
                                        display: 'flex', alignItems: 'center', gap: 12,
                                        padding: '10px 12px', cursor: 'pointer',
                                        border, background: bg,
                                        borderRadius: 6, margin: 4,
                                    }}
                                >
                                    <div style={{ width: 32, height: 32, borderRadius: 99, background: l.color, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#fff', fontSize: 11, fontWeight: 700 }}>
                                        {l.initials}
                                    </div>
                                    <div style={{ flex: 1, minWidth: 0 }}>
                                        <div style={{ fontWeight: 700, fontSize: 13, color: 'var(--text-primary)' }}>{l.name}</div>
                                        <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>{l.phone} · {l.email || '—'}</div>
                                    </div>
                                    <span className={`status-badge ${st.cls}`}>{st.label}</span>
                                    {isPrimary && <span style={{ fontSize: 11, color: '#3b82f6', fontWeight: 700 }}>Основная</span>}
                                    {isDuplicate && <span style={{ fontSize: 11, color: '#ef4444', fontWeight: 700 }}>Дубликат</span>}
                                </div>
                            );
                        })}
                    </div>
                </div>
                <div className="modal-footer">
                    <button className="btn btn-outline" onClick={onClose} disabled={busy}>Отмена</button>
                    <button className="btn btn-primary" onClick={handleMerge} disabled={!primaryId || !duplicateId || busy}>
                        <GitMerge size={14} /> {busy ? 'Объединение...' : 'Объединить'}
                    </button>
                </div>
            </div>
        </div>
    );
}
