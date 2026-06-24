import { useState, useMemo } from 'react';
import { Search, Download, Plus, Eye, Phone, MessageCircle, ChevronLeft, ChevronRight, Upload, GitMerge, Trash2, X } from 'lucide-react';
import { STATUSES, SOURCES, getStatusObj, formatDate } from '../data/crmData';
import { api } from '../services/api';
import TelegramChatModal from '../components/TelegramChatModal';
import CallModal from '../components/CallModal';
import { useUsers } from '../context/useUsers';

const PAGE_SIZE = 12;

export default function Leads({ allLeads, openDetail, openModal, openImport, openMerge, role, showToast, externalSearch = '', onExternalSearchChange, onStatusChange, unreadByLead = {} }) {
    const canEdit = role !== 'guest';
    const canManage = role === 'admin' || role === 'admissions';
    const { users, byId: usersById } = useUsers();
    // Гостей нельзя назначать менеджерами — исключаем из фильтра.
    // Никаких выдуманных «запасных» менеджеров: показываем только реальных.
    const managerOptions = users.filter(u => u.role === 'admin' || u.role === 'admissions');

    const [filterStatus, setFilterStatus] = useState('');
    const [filterSource, setFilterSource] = useState('');
    const [filterManager, setFilterManager] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    const [selected, setSelected] = useState(new Set());
    const [chatLead, setChatLead] = useState(null);
    const [callLead, setCallLead] = useState(null);
    const search = externalSearch || '';

    const filtered = useMemo(() => {
        const q = search.trim().toLowerCase();
        return allLeads.filter(l => {
            if (q) {
                const inName = (l.name || '').toLowerCase().includes(q);
                const inPhone = (l.phone || '').toLowerCase().includes(q);
                const inEmail = (l.email || '').toLowerCase().includes(q);
                if (!inName && !inPhone && !inEmail) return false;
            }
            if (filterStatus && getStatusObj(l.status).label !== filterStatus) return false;
            if (filterSource && l.source !== filterSource) return false;
            if (filterManager) {
                const realName = usersById.get(l.assigneeId)?.name;
                if (realName !== filterManager) return false;
            }
            return true;
        });
    }, [allLeads, search, filterStatus, filterSource, filterManager, usersById]);

    const pages = Math.ceil(filtered.length / PAGE_SIZE) || 1;
    const safePage = Math.min(currentPage, pages);
    const start = (safePage - 1) * PAGE_SIZE;
    const slice = filtered.slice(start, start + PAGE_SIZE);

    const allSelectedOnPage = slice.length > 0 && slice.every(l => selected.has(l.id));
    const someSelectedOnPage = slice.some(l => selected.has(l.id));

    const toggleAll = () => {
        const next = new Set(selected);
        if (allSelectedOnPage) {
            slice.forEach(l => next.delete(l.id));
        } else {
            slice.forEach(l => next.add(l.id));
        }
        setSelected(next);
    };

    const toggleOne = (id) => {
        const next = new Set(selected);
        if (next.has(id)) next.delete(id); else next.add(id);
        setSelected(next);
    };

    const updateSearch = (val) => {
        setCurrentPage(1);
        onExternalSearchChange && onExternalSearchChange(val);
    };

    const clearFilters = () => {
        setFilterStatus('');
        setFilterSource('');
        setFilterManager('');
        setCurrentPage(1);
        onExternalSearchChange && onExternalSearchChange('');
    };

    const handleExport = async (format = 'xlsx') => {
        if (!api.useApi) {
            showToast('Экспорт работает только с подключённым API', 'warning');
            return;
        }
        try {
            await api.exportAnalytics(format);
            showToast(`Скачивается ${format.toUpperCase()}…`, 'success');
        } catch (err) {
            showToast('Ошибка экспорта: ' + err.message, 'error');
        }
    };

    const handleBulkDelete = async () => {
        if (selected.size === 0) return;
        if (!window.confirm(`Удалить ${selected.size} выбранных лидов?`)) return;
        if (!api.useApi) {
            showToast('Удаление работает только с подключённым API', 'warning');
            return;
        }
        try {
            await Promise.all(Array.from(selected).map(id => api.deleteLead(id)));
            showToast(`Удалено: ${selected.size}`, 'success');
            setSelected(new Set());
            // Trigger reload via window event or just navigate
            window.location.reload();
        } catch (err) {
            showToast('Ошибка удаления: ' + err.message, 'error');
        }
    };

    const hasFilters = !!(search || filterStatus || filterSource || filterManager);

    return (
        <section className="page active">
            <div className="page-header">
                <div>
                    <h1>База лидов</h1>
                    <p className="page-sub">{filtered.length} {filtered.length === allLeads.length ? 'абитуриентов всего' : `из ${allLeads.length} (отфильтровано)`}</p>
                </div>
                <div className="header-actions">
                    <button className="btn btn-outline" onClick={() => handleExport('xlsx')}><Download size={14} /> Экспорт</button>
                    {canManage && <button className="btn btn-outline" onClick={openImport}><Upload size={14} /> Импорт Excel</button>}
                    {canManage && <button className="btn btn-outline" onClick={openMerge}><GitMerge size={14} /> Объединить</button>}
                    {canEdit && <button className="btn btn-primary" onClick={openModal}><Plus size={14} /> Новый лид</button>}
                </div>
            </div>

            <div className="filters-row">
                <label className="search-box search-inline" htmlFor="leads-search">
                    <Search size={15} />
                    <input
                        id="leads-search"
                        type="text"
                        placeholder="Поиск по имени, телефону, email..."
                        value={search}
                        onChange={e => updateSearch(e.target.value)}
                    />
                </label>
                <select className="filter-select" value={filterStatus} onChange={e => { setFilterStatus(e.target.value); setCurrentPage(1); }}>
                    <option value="">Все статусы</option>
                    {STATUSES.map(s => <option key={s.key}>{s.label}</option>)}
                </select>
                <select className="filter-select" value={filterSource} onChange={e => { setFilterSource(e.target.value); setCurrentPage(1); }}>
                    <option value="">Все источники</option>
                    {SOURCES.map(s => <option key={s}>{s}</option>)}
                </select>
                <select className="filter-select" value={filterManager} onChange={e => { setFilterManager(e.target.value); setCurrentPage(1); }}>
                    <option value="">Все менеджеры</option>
                    {managerOptions.map(u => <option key={u.id} value={u.name}>{u.name}</option>)}
                </select>
                {hasFilters && (
                    <button className="btn btn-outline" onClick={clearFilters} style={{ height: 38 }}>
                        <X size={13} /> Сбросить
                    </button>
                )}
            </div>

            {selected.size > 0 && (
                <div style={{ padding: '10px 14px', marginBottom: 12, borderRadius: 8, border: '1px solid var(--glass-border)', background: 'var(--glass)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                    <span style={{ fontSize: 13, color: 'var(--text-primary)' }}>Выбрано: <strong>{selected.size}</strong></span>
                    <div style={{ display: 'flex', gap: 8 }}>
                        <button className="btn btn-outline" onClick={() => setSelected(new Set())}><X size={13} /> Снять выделение</button>
                        <button className="btn btn-outline" onClick={() => handleExport('xlsx')}><Download size={13} /> Экспорт</button>
                        {canManage && <button className="btn btn-outline" onClick={handleBulkDelete} style={{ color: '#ef4444', borderColor: '#ef444440' }}><Trash2 size={13} /> Удалить</button>}
                    </div>
                </div>
            )}

            <div className="card table-card">
                <table className="leads-table">
                    <thead>
                        <tr>
                            <th>
                                <input
                                    type="checkbox"
                                    checked={allSelectedOnPage}
                                    ref={el => { if (el) el.indeterminate = !allSelectedOnPage && someSelectedOnPage; }}
                                    onChange={toggleAll}
                                    title={allSelectedOnPage ? 'Снять выделение со страницы' : 'Выбрать всё на странице'}
                                />
                            </th>
                            <th>Лид</th>
                            <th>Программа</th>
                            <th>Источник</th>
                            <th>Статус</th>
                            <th>Менеджер</th>
                            <th>Дата</th>
                            <th>Действия</th>
                        </tr>
                    </thead>
                    <tbody>
                        {slice.length === 0 && (
                            <tr><td colSpan="8" style={{ textAlign: 'center', padding: 40, color: 'var(--text-muted)' }}>Ничего не найдено</td></tr>
                        )}
                        {slice.map(l => {
                            const st = getStatusObj(l.status);
                            const isSel = selected.has(l.id);
                            return (
                                <tr key={l.id} onClick={() => openDetail(l)} className={isSel ? 'row-selected' : ''}>
                                    <td onClick={e => e.stopPropagation()}>
                                        <input type="checkbox" checked={isSel} onChange={() => toggleOne(l.id)} />
                                    </td>
                                    <td>
                                        <div className="lead-cell">
                                            <div className="lead-avatar" style={{ background: l.color, color: '#fff', fontSize: 11, fontWeight: 700 }}>{l.initials}</div>
                                            <div>
                                                <div className="lead-cell-name">{l.name}</div>
                                                <div className="lead-cell-phone">{l.phone}</div>
                                            </div>
                                        </div>
                                    </td>
                                    <td style={{ color: 'var(--text-secondary)' }}>{l.program}</td>
                                    <td><span className="source-chip">{l.source}</span></td>
                                    <td><span className={`status-badge ${st.cls}`}>{st.label}</span></td>
                                    <td>
                                        {(() => {
                                            const u = usersById.get(l.assigneeId);
                                            const name = u?.name || 'Не назначен';
                                            const initials = u?.initials || '—';
                                            const color = u?.color || '#94a3b8';
                                            return (
                                                <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                                                    <div style={{ width: 22, height: 22, borderRadius: 99, background: color, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 9, fontWeight: 700, color: '#fff' }}>{initials}</div>
                                                    <span style={{ fontSize: 12 }}>{name.split(' ')[0]}</span>
                                                </div>
                                            );
                                        })()}
                                    </td>
                                    <td style={{ color: 'var(--text-muted)' }}>{formatDate(l.date)}</td>
                                    <td onClick={e => e.stopPropagation()}>
                                        <div className="action-btns">
                                            <button className="action-btn" onClick={() => openDetail(l)} title="Открыть"><Eye size={13} /></button>
                                            <button className="action-btn" onClick={() => setCallLead(l)} title="Позвонить"><Phone size={13} /></button>
                                            <button className="action-btn" onClick={() => setChatLead(l)} title="Чат в Telegram" style={{ position: 'relative' }}>
                                                <MessageCircle size={13} />
                                                {unreadByLead[l.id] > 0 && (
                                                    <span style={{
                                                        position: 'absolute', top: -3, right: -3,
                                                        minWidth: 14, height: 14, padding: '0 3px',
                                                        borderRadius: 99, background: '#ef4444', color: '#fff',
                                                        fontSize: 9, fontWeight: 700, display: 'flex',
                                                        alignItems: 'center', justifyContent: 'center',
                                                        border: '1.5px solid var(--bg-card)',
                                                    }}>{unreadByLead[l.id]}</span>
                                                )}
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
                {filtered.length > 0 && (
                    <div className="table-pagination">
                        <span className="pagination-info">Показано {start + 1}–{Math.min(start + PAGE_SIZE, filtered.length)} из {filtered.length}</span>
                        <div className="pagination-btns">
                            <button className="page-btn" onClick={() => setCurrentPage(p => Math.max(1, p - 1))}><ChevronLeft size={13} /></button>
                            <div className="page-numbers">
                                {Array.from({ length: Math.min(pages, 7) }, (_, i) => i + 1).map(p => (
                                    <button key={p} className={`page-number-btn ${p === safePage ? 'active' : ''}`} onClick={() => setCurrentPage(p)}>{p}</button>
                                ))}
                            </div>
                            <button className="page-btn" onClick={() => setCurrentPage(p => Math.min(pages, p + 1))}><ChevronRight size={13} /></button>
                        </div>
                    </div>
                )}
            </div>

            {chatLead && (
                <TelegramChatModal
                    lead={chatLead}
                    onClose={() => setChatLead(null)}
                    showToast={showToast}
                    role={role}
                />
            )}

            {callLead && (
                <CallModal
                    lead={callLead}
                    onClose={() => setCallLead(null)}
                    onSaved={() => {
                        if (callLead.status === 'new' && onStatusChange) onStatusChange(callLead.id, 'consultation');
                    }}
                    showToast={showToast}
                />
            )}
        </section>
    );
}
