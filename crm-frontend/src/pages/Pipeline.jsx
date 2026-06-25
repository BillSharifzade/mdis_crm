import { useState, useMemo } from 'react';
import { Filter, Plus, X } from 'lucide-react';
import { SOURCES } from '../data/crmData';
import { useStages } from '../context/useStages';
import { useUsers } from '../context/useUsers';

export default function Pipeline({ allLeads, openDetail, openModal, onStatusChange, role }) {
    const { stages: stagesFromCtx } = useStages();
    const { users, byId: usersById } = useUsers();
    const stages = stagesFromCtx && stagesFromCtx.length > 0 ? stagesFromCtx : [];
    const [dragId, setDragId] = useState(null);
    const [overKey, setOverKey] = useState(null);
    const [reasonPrompt, setReasonPrompt] = useState(null); // { id, status }
    const [reasonText, setReasonText] = useState('');

    // Фильтры (T11)
    const [filtersOpen, setFiltersOpen] = useState(false);
    const [filterSource, setFilterSource] = useState('');
    const [filterManager, setFilterManager] = useState('');

    const managerOptions = users.filter(u => u.role === 'admin' || u.role === 'admissions');
    const managerNamesList = managerOptions.map(u => u.name);

    const filteredLeads = useMemo(() => {
        return allLeads.filter(l => {
            if (filterSource && l.source !== filterSource) return false;
            if (filterManager) {
                const real = usersById.get(l.assigneeId)?.name;
                if (real !== filterManager) return false;
            }
            return true;
        });
    }, [allLeads, filterSource, filterManager, usersById]);

    const canEdit = role !== 'guest';

    const handleDragStart = (e, leadId) => {
        if (!canEdit) return;
        setDragId(leadId);
        e.dataTransfer.effectAllowed = 'move';
    };

    const handleDragOver = (e, stageKey) => {
        if (!canEdit) return;
        e.preventDefault();
        e.dataTransfer.dropEffect = 'move';
        if (overKey !== stageKey) setOverKey(stageKey);
    };

    const handleDrop = (e, stageKey) => {
        if (!canEdit) return;
        e.preventDefault();
        const id = dragId;
        setDragId(null);
        setOverKey(null);
        if (!id) return;
        const lead = allLeads.find(l => l.id === id);
        if (!lead || lead.status === stageKey) return;

        // Финальный этап с пометкой is_final, у которого ключ lost — требует причину.
        const stage = stages.find(s => s.key === stageKey);
        if (stageKey === 'lost' || (stage && stage.isFinal && stage.key === 'lost')) {
            setReasonPrompt({ id, status: stageKey });
            setReasonText('');
            return;
        }
        onStatusChange && onStatusChange(id, stageKey);
    };

    const confirmReason = async () => {
        if (!reasonText.trim()) return;
        const ok = await onStatusChange(reasonPrompt.id, reasonPrompt.status, reasonText.trim());
        if (ok !== false) {
            setReasonPrompt(null);
            setReasonText('');
        }
    };

    return (
        <section className="page active">
            <div className="page-header">
                <div>
                    <h1>Воронка продаж</h1>
                    <p className="page-sub">
                        Канбан — путь абитуриента
                        {canEdit && <span style={{ marginLeft: 8, color: 'var(--text-muted)', fontSize: 12 }}>· перетащите карточку для смены этапа</span>}
                    </p>
                </div>
                <div className="header-actions">
                    <button className={`btn btn-outline ${filtersOpen ? 'active' : ''}`} onClick={() => setFiltersOpen(o => !o)}>
                        <Filter size={14} /> Фильтр
                        {(filterSource || filterManager) && <span style={{ marginLeft: 4, background: '#6366f1', color: '#fff', borderRadius: 99, padding: '0 6px', fontSize: 10 }}>•</span>}
                    </button>
                    {canEdit && <button className="btn btn-primary" onClick={openModal}><Plus size={14} /> Новый лид</button>}
                </div>
            </div>

            {filtersOpen && (
                <div className="filters-row" style={{ marginBottom: 14 }}>
                    <select className="filter-select" value={filterSource} onChange={e => setFilterSource(e.target.value)}>
                        <option value="">Все источники</option>
                        {SOURCES.map(s => <option key={s}>{s}</option>)}
                    </select>
                    <select className="filter-select" value={filterManager} onChange={e => setFilterManager(e.target.value)}>
                        <option value="">Все менеджеры</option>
                        {managerNamesList.map(n => <option key={n}>{n}</option>)}
                    </select>
                    {(filterSource || filterManager) && (
                        <button className="btn btn-outline" onClick={() => { setFilterSource(''); setFilterManager(''); }}>
                            <X size={13} /> Сбросить
                        </button>
                    )}
                </div>
            )}

            <div className="kanban-board">
                {stages.map(st => {
                    const cards = filteredLeads.filter(l => l.status === st.key);
                    const isOver = overKey === st.key;
                    return (
                        <div
                            className="kanban-col"
                            key={st.key}
                            onDragOver={(e) => handleDragOver(e, st.key)}
                            onDragLeave={() => setOverKey(prev => prev === st.key ? null : prev)}
                            onDrop={(e) => handleDrop(e, st.key)}
                            style={isOver ? { outline: `2px dashed ${st.color}`, outlineOffset: -4, background: `${st.color}10` } : undefined}
                        >
                            <div className="kanban-col-header">
                                <div className="kanban-col-title">
                                    <div className="kanban-col-dot" style={{ background: st.color }}></div>
                                    {st.label}
                                </div>
                                <span className="kanban-col-count">{cards.length}</span>
                            </div>
                            <div className="kanban-cards">
                                {cards.slice(0, 12).map(l => (
                                    <div
                                        className="kanban-card"
                                        key={l.id}
                                        style={{ '--col-color': st.color, opacity: dragId === l.id ? 0.4 : 1, cursor: canEdit ? 'grab' : 'pointer' }}
                                        onClick={() => openDetail(l)}
                                        draggable={canEdit}
                                        onDragStart={(e) => handleDragStart(e, l.id)}
                                        onDragEnd={() => { setDragId(null); setOverKey(null); }}
                                    >
                                        <div className="kanban-card-name">{l.name}</div>
                                        <div className="kanban-card-program">{l.program}</div>
                                        <div className="kanban-card-footer">
                                            <span className="kanban-card-source">{l.source}</span>
                                            {(() => {
                                                const u = usersById.get(l.assigneeId);
                                                return (
                                                    <div className="kanban-card-avatar" style={{ background: u?.color || '#94a3b8' }} title={u?.name || 'Не назначен'}>{u?.initials || '—'}</div>
                                                );
                                            })()}
                                        </div>
                                    </div>
                                ))}
                                {cards.length > 12 && (
                                    <div style={{ textAlign: 'center', fontSize: 12, color: 'var(--text-muted)', padding: 6 }}>+{cards.length - 12} ещё</div>
                                )}
                                {cards.length === 0 && (
                                    <div style={{ textAlign: 'center', fontSize: 12, color: 'var(--text-muted)', padding: 16, border: '1px dashed var(--glass-border)', borderRadius: 8 }}>
                                        Пусто
                                    </div>
                                )}
                            </div>
                        </div>
                    );
                })}
            </div>

            {reasonPrompt && (
                <div className="modal-overlay open" onClick={(e) => { if (e.target === e.currentTarget) setReasonPrompt(null); }}>
                    <div className="modal" style={{ maxWidth: 480 }}>
                        <div className="modal-header">
                            <h2>Причина отказа</h2>
                        </div>
                        <div className="modal-body">
                            <p style={{ color: 'var(--text-secondary)', fontSize: 13, marginBottom: 12 }}>
                                Для перевода в статус «Отказ» необходимо указать причину (требование ТЗ, п. 6.5).
                            </p>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 12 }}>
                                {['Высокая цена', 'Выбрал другой вуз', 'Не прошёл по баллам', 'Передумал', 'Нет ответа'].map(opt => (
                                    <button
                                        key={opt}
                                        type="button"
                                        className="btn btn-outline"
                                        style={{ justifyContent: 'flex-start' }}
                                        onClick={() => setReasonText(opt)}
                                    >
                                        {opt}
                                    </button>
                                ))}
                            </div>
                            <textarea
                                placeholder="Или укажите свою причину..."
                                value={reasonText}
                                onChange={e => setReasonText(e.target.value)}
                                rows="3"
                                style={{ width: '100%', padding: 10, borderRadius: 8, border: '1px solid var(--glass-border)', background: 'var(--glass)', color: 'var(--text-primary)', fontSize: 13, fontFamily: 'inherit' }}
                            />
                        </div>
                        <div className="modal-footer">
                            <button className="btn btn-outline" onClick={() => setReasonPrompt(null)}>Отмена</button>
                            <button className="btn btn-primary" disabled={!reasonText.trim()} onClick={confirmReason}>Подтвердить</button>
                        </div>
                    </div>
                </div>
            )}
        </section>
    );
}
