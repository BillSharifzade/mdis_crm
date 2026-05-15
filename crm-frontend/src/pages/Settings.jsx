import { useState, useEffect, useCallback } from 'react';
import { createPortal } from 'react-dom';
import { api } from '../services/api';
import {
    Plus, Pencil, GripVertical, X, Globe, Send, MessageSquare, Camera, Hash,
    Phone as PhoneIcon, Mail, BarChart3, Trash2, Save, Activity, ArrowUp, ArrowDown,
} from 'lucide-react';
import { getInitials } from '../data/crmData';
import { useStages } from '../context/useStages';

const tabs = [
    { id: 'users', label: 'Пользователи' },
    { id: 'pipeline', label: 'Этапы воронки' },
    { id: 'faculties', label: 'Факультеты' },
    { id: 'sources', label: 'Источники' },
    { id: 'bot', label: 'Telegram-бот' },
    { id: 'integrations', label: 'Интеграции' },
];

const integrationsDefault = [
    { key: 'site', Icon: Globe, color: '#818cf8', name: 'Сайт MDIS', desc: 'Формы заявок и обратного звонка', on: true, bg: 'rgba(99,102,241,0.15)' },
    { key: 'telegram', Icon: Send, color: '#06b6d4', name: 'Telegram', desc: 'Официальный бот приёмной комиссии', on: true, bg: 'rgba(6,182,212,0.15)' },
    { key: 'whatsapp', Icon: MessageSquare, color: '#10b981', name: 'WhatsApp', desc: 'WhatsApp Business API', on: true, bg: 'rgba(16,185,129,0.15)' },
    { key: 'instagram', Icon: Camera, color: '#ec4899', name: 'Instagram', desc: 'Direct-сообщения', on: false, bg: 'rgba(236,72,153,0.15)' },
    { key: 'vk', Icon: Hash, color: '#f59e0b', name: 'ВКонтакте', desc: 'VK Сообщения и лид-формы', on: false, bg: 'rgba(245,158,11,0.15)' },
    { key: 'telephony', Icon: PhoneIcon, color: '#a855f7', name: 'Телефония', desc: 'Облачная АТС, запись звонков', on: true, bg: 'rgba(168,85,247,0.15)' },
    { key: 'email', Icon: Mail, color: '#ef4444', name: 'Email', desc: 'SMTP / IMAP входящие обращения', on: true, bg: 'rgba(239,68,68,0.15)' },
    { key: 'metrika', Icon: BarChart3, color: '#f59e0b', name: 'Яндекс.Метрика', desc: 'UTM-метки и аналитика трафика', on: true, bg: 'rgba(245,158,11,0.15)' },
];

export default function Settings({ allLeads, showToast, role }) {
    const [activeTab, setActiveTab] = useState('users');
    const isAdmin = role === 'admin';

    return (
        <section className="page active">
            <div className="page-header">
                <div>
                    <h1>Настройки</h1>
                    <p className="page-sub">Управление системой и пользователями</p>
                </div>
            </div>
            <div className="settings-layout">
                <div className="settings-tabs">
                    {tabs.map(t => (
                        <button key={t.id} className={`settings-tab ${activeTab === t.id ? 'active' : ''}`} onClick={() => setActiveTab(t.id)}>
                            {t.label}
                        </button>
                    ))}
                </div>
                <div className="settings-content">
                    {activeTab === 'users' && <UsersTab showToast={showToast} isAdmin={isAdmin} />}
                    {activeTab === 'pipeline' && <PipelineTab allLeads={allLeads} showToast={showToast} isAdmin={isAdmin} />}
                    {activeTab === 'faculties' && <FacultiesTab showToast={showToast} isAdmin={isAdmin} />}
                    {activeTab === 'sources' && <SourcesTab showToast={showToast} isAdmin={isAdmin} />}
                    {activeTab === 'bot' && <BotTab showToast={showToast} isAdmin={isAdmin} />}
                    {activeTab === 'integrations' && <IntegrationsTab showToast={showToast} />}
                </div>
            </div>
        </section>
    );
}

// ── USERS + KPI ─────────────────────────────────────────────────────────

function UsersTab({ showToast, isAdmin }) {
    const [usersData, setUsersData] = useState([]);
    const [loading, setLoading] = useState(true);
    const [editorOpen, setEditorOpen] = useState(false);
    const [editorTarget, setEditorTarget] = useState(null);
    const [kpiTarget, setKpiTarget] = useState(null);

    const loadUsers = useCallback(async () => {
        if (!api.useApi) { setLoading(false); return; }
        try {
            setLoading(true);
            const data = await api.getUsers();
            setUsersData(data);
        } catch (error) {
            console.error(error);
            showToast('Ошибка загрузки пользователей', 'error');
        } finally {
            setLoading(false);
        }
    }, [showToast]);

    useEffect(() => { loadUsers(); }, [loadUsers]);

    const handleDelete = async (user) => {
        if (!window.confirm(`Удалить пользователя «${user.name}»? Все его лиды останутся без менеджера.`)) return;
        try {
            await api.deleteUser(user.id);
            showToast('Пользователь удалён', 'success');
            await loadUsers();
        } catch (err) {
            showToast('Ошибка удаления: ' + err.message, 'error');
        }
    };

    const openCreate = () => { setEditorTarget(null); setEditorOpen(true); };
    const openEdit = (u) => { setEditorTarget(u); setEditorOpen(true); };

    const colors = ['#6366f1', '#10b981', '#f59e0b', '#a855f7', '#ec4899'];

    return (
        <div className="users-table-wrap">
            <div className="users-table-header">
                <h3>Пользователи системы</h3>
                {isAdmin && (
                    <button className="btn btn-primary" onClick={openCreate}>
                        <Plus size={14} /> Добавить
                    </button>
                )}
            </div>
            <table className="settings-table">
                <thead><tr><th>Пользователь</th><th>Email</th><th>Роль</th><th>Действия</th></tr></thead>
                <tbody>
                    {loading ? <tr><td colSpan="4" style={{ textAlign: 'center', padding: 20 }}>Загрузка...</td></tr> : usersData.map((u, i) => {
                        const c = colors[i % colors.length];
                        return (
                            <tr key={u.id || i}>
                                <td>
                                    <div className="manager-cell">
                                        <div style={{ width: 32, height: 32, borderRadius: 99, background: c, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 11, fontWeight: 700, color: '#fff' }}>{getInitials(u.name || 'NN')}</div>
                                        <span className="manager-name">{u.name}</span>
                                    </div>
                                </td>
                                <td>{u.email}</td>
                                <td><span style={{ padding: '3px 9px', borderRadius: 99, fontSize: 11, fontWeight: 700, background: `${c}20`, color: c, border: `1px solid ${c}30` }}>{u.role}</span></td>
                                <td>
                                    {isAdmin && (
                                        <div style={{ display: 'flex', gap: 6 }}>
                                            {u.role === 'admissions' && (
                                                <button className="action-btn" onClick={() => setKpiTarget(u)} title="KPI и активность" style={{ color: '#a78bfa' }}>
                                                    <Activity size={13} />
                                                </button>
                                            )}
                                            <button className="action-btn" onClick={() => openEdit(u)} title="Редактировать"><Pencil size={13} /></button>
                                            <button className="action-btn" onClick={() => handleDelete(u)} title="Удалить" style={{ color: '#ef4444' }}><Trash2 size={13} /></button>
                                        </div>
                                    )}
                                </td>
                            </tr>
                        );
                    })}
                    {!loading && usersData.length === 0 && (
                        <tr><td colSpan="4" style={{ textAlign: 'center', padding: 20, color: 'var(--text-muted)' }}>Нет данных {api.useApi ? '' : '(API отключен)'}</td></tr>
                    )}
                </tbody>
            </table>

            {editorOpen && (
                <UserEditorModal
                    target={editorTarget}
                    onClose={() => setEditorOpen(false)}
                    onSaved={() => { setEditorOpen(false); loadUsers(); }}
                    showToast={showToast}
                />
            )}

            {kpiTarget && (
                <KpiModal user={kpiTarget} onClose={() => setKpiTarget(null)} showToast={showToast} />
            )}
        </div>
    );
}

function UserEditorModal({ target, onClose, onSaved, showToast }) {
    const isEdit = !!target;
    const [form, setForm] = useState({
        name: target?.name || '',
        email: target?.email || '',
        password: '',
        role: target?.role || 'admissions',
    });
    const [busy, setBusy] = useState(false);

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!form.name || !form.email) { showToast('Заполните ФИО и email', 'warning'); return; }
        if (!isEdit && !form.password) { showToast('Пароль обязателен при создании', 'warning'); return; }
        setBusy(true);
        try {
            if (isEdit) {
                const payload = { name: form.name, email: form.email, role: form.role };
                if (form.password) payload.password = form.password;
                await api.updateUser(target.id, payload);
                showToast('Изменения сохранены', 'success');
            } else {
                await api.createUser(form);
                showToast('Пользователь создан', 'success');
            }
            onSaved();
        } catch (err) {
            console.error(err);
            showToast((isEdit ? 'Ошибка обновления: ' : 'Ошибка создания: ') + err.message, 'error');
        } finally {
            setBusy(false);
        }
    };

    return createPortal(
        <div className="modal-overlay open" onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}>
            <div className="modal" style={{ maxWidth: 460 }}>
                <div className="modal-header">
                    <h2>{isEdit ? 'Редактировать пользователя' : 'Новый пользователь'}</h2>
                    <button className="modal-close" onClick={onClose}><X size={14} /></button>
                </div>
                <form onSubmit={handleSubmit}>
                    <div className="modal-body">
                        <div className="form-grid">
                            <div className="form-group full-width">
                                <label>ФИО *</label>
                                <input type="text" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} />
                            </div>
                            <div className="form-group full-width">
                                <label>Email *</label>
                                <input type="email" value={form.email} onChange={e => setForm({ ...form, email: e.target.value })} />
                            </div>
                            <div className="form-group full-width">
                                <label>{isEdit ? 'Новый пароль (оставьте пустым, чтобы не менять)' : 'Пароль *'}</label>
                                <input type="password" value={form.password} onChange={e => setForm({ ...form, password: e.target.value })} placeholder={isEdit ? '••••••••' : ''} />
                            </div>
                            <div className="form-group full-width">
                                <label>Роль</label>
                                <select value={form.role} onChange={e => setForm({ ...form, role: e.target.value })}>
                                    <option value="admin">Администратор</option>
                                    <option value="admissions">Менеджер приёма</option>
                                    <option value="guest">Гость (только просмотр)</option>
                                </select>
                            </div>
                        </div>
                    </div>
                    <div className="modal-footer">
                        <button type="button" className="btn btn-outline" onClick={onClose} disabled={busy}>Отмена</button>
                        <button type="submit" className="btn btn-primary" disabled={busy}>
                            {busy ? 'Сохранение...' : (isEdit ? 'Сохранить' : 'Создать')}
                        </button>
                    </div>
                </form>
            </div>
        </div>,
        document.body
    );
}

function KpiModal({ user, onClose, showToast }) {
    const [period, setPeriod] = useState(30);
    const [stats, setStats] = useState(null);
    const [busy, setBusy] = useState(false);
    const [processedGoal, setProcessedGoal] = useState(0);
    const [createdGoal, setCreatedGoal] = useState(0);

    const load = useCallback(async () => {
        try {
            const s = await api.getKpiStats(user.id, period);
            setStats(s);
            setProcessedGoal(s.processed_goal || 0);
            setCreatedGoal(s.created_goal || 0);
        } catch (err) {
            showToast('Ошибка загрузки KPI: ' + err.message, 'error');
        }
    }, [user.id, period, showToast]);

    useEffect(() => { load(); }, [load]);

    const saveAll = async () => {
        setBusy(true);
        try {
            await Promise.all([
                api.upsertKpiTarget(user.id, 'processed', { target_count: Number(processedGoal), period_days: Number(period) }),
                api.upsertKpiTarget(user.id, 'created', { target_count: Number(createdGoal), period_days: Number(period) }),
            ]);
            showToast('Цели KPI сохранены', 'success');
            await load();
        } catch (err) {
            showToast('Ошибка: ' + err.message, 'error');
        } finally {
            setBusy(false);
        }
    };

    return createPortal(
        <div className="modal-overlay open" onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}>
            <div className="modal" style={{ maxWidth: 600 }}>
                <div className="modal-header">
                    <h2>KPI · {user.name}</h2>
                    <button className="modal-close" onClick={onClose}><X size={14} /></button>
                </div>
                <div className="modal-body">
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 16 }}>
                        <label style={{ fontSize: 12, color: 'var(--text-secondary)' }}>Период:</label>
                        <select value={period} onChange={e => setPeriod(Number(e.target.value))} className="filter-select" style={{ width: 130 }}>
                            <option value={1}>1 день</option>
                            <option value={7}>7 дней</option>
                            <option value={14}>14 дней</option>
                            <option value={30}>30 дней</option>
                            <option value={60}>60 дней</option>
                            <option value={90}>90 дней</option>
                        </select>
                    </div>
                    {stats && (
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, marginBottom: 16 }}>
                            <Metric label="Обработано лидов" value={stats.processed} color="#6366f1" />
                            <Metric label="Создано лидов" value={stats.created} color="#10b981" />
                            <Metric label="Звонков" value={stats.calls_count} color="#a78bfa" />
                            <Metric label="Заметок" value={stats.notes_count} color="#f59e0b" />
                            <Metric label="Сообщений" value={stats.messages_count} color="#06b6d4" />
                        </div>
                    )}
                    <hr style={{ border: 0, borderTop: '1px solid var(--glass-border)', margin: '14px 0' }} />
                    <h3 style={{ fontSize: 13, fontWeight: 700, color: 'var(--text-primary)', marginBottom: 10 }}>Целевые KPI</h3>

                    <div className="form-group" style={{ marginBottom: 14 }}>
                        <label>Обработка (любое действие с лидом за {period} {period === 1 ? 'день' : 'дн.'})</label>
                        <input type="number" min="0" value={processedGoal} onChange={e => setProcessedGoal(e.target.value)} style={{ width: '100%' }} />
                        {stats && <KpiBar val={stats.processed} goal={Number(processedGoal)} color="#6366f1" />}
                    </div>

                    <div className="form-group">
                        <label>Создание лидов за {period} {period === 1 ? 'день' : 'дн.'}</label>
                        <input type="number" min="0" value={createdGoal} onChange={e => setCreatedGoal(e.target.value)} style={{ width: '100%' }} />
                        {stats && <KpiBar val={stats.created} goal={Number(createdGoal)} color="#10b981" />}
                    </div>
                </div>
                <div className="modal-footer">
                    <button className="btn btn-outline" onClick={onClose} disabled={busy}>Закрыть</button>
                    <button className="btn btn-primary" onClick={saveAll} disabled={busy}>
                        <Save size={13} /> {busy ? 'Сохранение…' : 'Сохранить'}
                    </button>
                </div>
            </div>
        </div>,
        document.body
    );
}

// KpiBar — прогресс-бар с поддержкой overpercentage (>100% подсвечивается).
export function KpiBar({ val, goal, color }) {
    if (!goal || goal <= 0) {
        return (
            <div style={{ fontSize: 11, color: 'var(--text-muted)', marginTop: 6 }}>
                Текущее: {val} · цель не задана
            </div>
        );
    }
    const rawPct = Math.round((val / goal) * 100);
    const capped = Math.min(100, rawPct);
    const over = rawPct > 100;
    const fillColor = over ? '#10b981' : color;
    return (
        <>
            <div style={{
                height: 8, background: 'var(--glass)', borderRadius: 99,
                overflow: 'hidden', marginTop: 6, position: 'relative',
            }}>
                <div style={{ height: '100%', width: `${capped}%`, background: fillColor, transition: 'width .4s' }}></div>
                {over && (
                    <div style={{
                        position: 'absolute', inset: 0,
                        background: `repeating-linear-gradient(45deg, transparent 0 4px, rgba(255,255,255,0.18) 4px 8px)`,
                        pointerEvents: 'none',
                    }}></div>
                )}
            </div>
            <div style={{
                fontSize: 11, marginTop: 4,
                color: over ? '#10b981' : 'var(--text-muted)',
                fontWeight: over ? 700 : 400,
            }}>
                {val} / {goal} · {rawPct}% {over && `(перевыполнение +${rawPct - 100}%)`}
            </div>
        </>
    );
}

function Metric({ label, value, color }) {
    return (
        <div style={{ padding: 12, borderRadius: 10, border: '1px solid var(--glass-border)', background: 'var(--glass)' }}>
            <div style={{ fontSize: 11, color: 'var(--text-muted)', marginBottom: 4 }}>{label}</div>
            <div style={{ fontSize: 22, fontWeight: 800, color, lineHeight: 1 }}>{value}</div>
        </div>
    );
}

// ── PIPELINE — full CRUD + reorder ──────────────────────────────────────

function PipelineTab({ allLeads, showToast, isAdmin }) {
    const { reload } = useStages();
    const [items, setItems] = useState([]);
    const [loading, setLoading] = useState(true);
    const [editing, setEditing] = useState(null); // stage object | { __new: true }
    const [busy, setBusy] = useState(false);
    const [dragIdx, setDragIdx] = useState(null);

    const load = useCallback(async () => {
        try {
            setLoading(true);
            const list = await api.listAllStages();
            setItems(list);
        } catch (err) {
            showToast('Ошибка загрузки этапов: ' + err.message, 'error');
        } finally {
            setLoading(false);
        }
    }, [showToast]);

    useEffect(() => { load(); }, [load]);

    const persistOrder = async (next) => {
        try {
            await api.reorderStages(next.map(s => ({ id: s.id, order: s.order })));
            reload();
            showToast('Порядок этапов сохранён', 'success');
        } catch (err) {
            showToast('Ошибка: ' + err.message, 'error');
            load();
        }
    };

    const move = async (idx, dir) => {
        const target = idx + dir;
        if (target < 0 || target >= items.length) return;
        const next = [...items];
        [next[idx], next[target]] = [next[target], next[idx]];
        next.forEach((s, i) => { s.order = i + 1; });
        setItems(next);
        await persistOrder(next);
    };

    const handleDragStart = (e, idx) => {
        if (!isAdmin) return;
        setDragIdx(idx);
        e.dataTransfer.effectAllowed = 'move';
    };

    const handleDragOver = (e, idx) => {
        if (!isAdmin || dragIdx === null || dragIdx === idx) return;
        e.preventDefault();
        const next = [...items];
        const [moved] = next.splice(dragIdx, 1);
        next.splice(idx, 0, moved);
        next.forEach((s, i) => { s.order = i + 1; });
        setItems(next);
        setDragIdx(idx);
    };

    const handleDragEnd = () => {
        if (dragIdx !== null) {
            persistOrder(items);
        }
        setDragIdx(null);
    };

    const handleDelete = async (s) => {
        if (!window.confirm(`Удалить этап «${s.name}»?`)) return;
        try {
            const res = await api.deleteStage(s.id);
            if (res?.status === 'archived') {
                showToast('Этап используется — заархивирован', 'warning');
            } else {
                showToast('Этап удалён', 'success');
            }
            reload();
            load();
        } catch (err) {
            showToast('Ошибка: ' + err.message, 'error');
        }
    };

    return (
        <div className="users-table-wrap">
            <div className="users-table-header">
                <h3>Этапы воронки</h3>
                {isAdmin && (
                    <button className="btn btn-primary" onClick={() => setEditing({ __new: true, name: '', is_final: false })}>
                        <Plus size={14} /> Добавить этап
                    </button>
                )}
            </div>
            {loading ? <p style={{ padding: 14, color: 'var(--text-muted)' }}>Загрузка…</p> : (
                <div className="pipeline-stage-list" style={{ padding: 16 }}>
                    {isAdmin && (
                        <p style={{ fontSize: 11, color: 'var(--text-muted)', margin: '0 0 10px' }}>
                            Перетаскивайте за <GripVertical size={11} style={{ verticalAlign: 'middle' }}/> или используйте стрелки.
                        </p>
                    )}
                    {items.map((s, i) => (
                        <div
                            className="stage-item"
                            key={s.id}
                            draggable={isAdmin}
                            onDragStart={(e) => handleDragStart(e, i)}
                            onDragOver={(e) => handleDragOver(e, i)}
                            onDragEnd={handleDragEnd}
                            style={{
                                opacity: dragIdx === i ? 0.4 : (s.is_active ? 1 : 0.5),
                                cursor: isAdmin ? 'grab' : 'default',
                            }}
                        >
                            <span className="stage-drag" style={{ cursor: isAdmin ? 'grab' : 'default' }}><GripVertical size={14} /></span>
                            <div className="stage-dot" style={{ background: s.is_final ? '#ef4444' : '#6366f1' }}></div>
                            <span className="stage-name">
                                {s.name}
                                {s.is_final && <span style={{ marginLeft: 6, fontSize: 10, color: '#ef4444' }}>финальный</span>}
                                {!s.is_active && <span style={{ marginLeft: 6, fontSize: 10, color: 'var(--text-muted)' }}>скрыт</span>}
                            </span>
                            <span style={{ fontSize: 12, color: 'var(--text-muted)', marginLeft: 'auto' }}>
                                {allLeads.filter(l => l.statusId === s.id).length} лидов
                            </span>
                            {isAdmin && (
                                <div style={{ display: 'flex', gap: 6, marginLeft: 12 }}>
                                    <button className="action-btn" onClick={() => move(i, -1)} disabled={i === 0} title="Выше"><ArrowUp size={12} /></button>
                                    <button className="action-btn" onClick={() => move(i, 1)} disabled={i === items.length - 1} title="Ниже"><ArrowDown size={12} /></button>
                                    <button className="action-btn" onClick={() => setEditing(s)} title="Редактировать"><Pencil size={12} /></button>
                                    <button className="action-btn" onClick={() => handleDelete(s)} title="Удалить" style={{ color: '#ef4444' }}><Trash2 size={12} /></button>
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            )}

            {editing && (
                <CrudFormModal
                    title={editing.__new ? 'Новый этап' : 'Этап воронки'}
                    fields={[
                        { name: 'name', label: 'Название *', type: 'text' },
                        { name: 'is_final', label: 'Финальный (закрытие)', type: 'checkbox' },
                        ...(editing.__new ? [] : [{ name: 'is_active', label: 'Активен', type: 'checkbox' }]),
                    ]}
                    initial={editing}
                    busy={busy}
                    onCancel={() => setEditing(null)}
                    onSubmit={async (data) => {
                        setBusy(true);
                        try {
                            if (editing.__new) {
                                await api.createStage({ name: data.name, is_final: !!data.is_final });
                                showToast('Этап создан', 'success');
                            } else {
                                await api.updateStage(editing.id, { name: data.name, is_final: !!data.is_final, is_active: data.is_active });
                                showToast('Этап обновлён', 'success');
                            }
                            reload();
                            setEditing(null);
                            load();
                        } catch (err) {
                            showToast('Ошибка: ' + err.message, 'error');
                        } finally {
                            setBusy(false);
                        }
                    }}
                />
            )}
        </div>
    );
}

// ── FACULTIES (programs) ────────────────────────────────────────────────

function FacultiesTab({ showToast, isAdmin }) {
    const [items, setItems] = useState([]);
    const [loading, setLoading] = useState(true);
    const [editing, setEditing] = useState(null);
    const [busy, setBusy] = useState(false);

    const load = useCallback(async () => {
        try {
            setLoading(true);
            setItems(await api.listAllPrograms());
        } catch (err) {
            showToast('Ошибка загрузки: ' + err.message, 'error');
        } finally {
            setLoading(false);
        }
    }, [showToast]);

    useEffect(() => { load(); }, [load]);

    const handleDelete = async (p) => {
        if (!window.confirm(`Удалить «${p.name}»?`)) return;
        try {
            const res = await api.deleteProgram(p.id);
            showToast(res?.status === 'archived' ? 'Программа используется — заархивирована' : 'Программа удалена',
                res?.status === 'archived' ? 'warning' : 'success');
            load();
        } catch (err) {
            showToast('Ошибка: ' + err.message, 'error');
        }
    };

    return (
        <div className="users-table-wrap">
            <div className="users-table-header">
                <h3>Программы / Факультеты</h3>
                {isAdmin && (
                    <button className="btn btn-primary" onClick={() => setEditing({ __new: true, name: '', faculty: '' })}>
                        <Plus size={14} /> Добавить
                    </button>
                )}
            </div>
            {loading ? <p style={{ padding: 14, color: 'var(--text-muted)' }}>Загрузка…</p> : (
                <table className="settings-table">
                    <thead><tr><th>Программа</th><th>Факультет</th><th>Статус</th><th>Действия</th></tr></thead>
                    <tbody>
                        {items.map(p => (
                            <tr key={p.id} style={{ opacity: p.is_active ? 1 : 0.5 }}>
                                <td style={{ fontWeight: 600 }}>{p.name}</td>
                                <td style={{ color: 'var(--text-secondary)' }}>{p.faculty || '—'}</td>
                                <td>
                                    <span style={{
                                        padding: '3px 9px', borderRadius: 99, fontSize: 11, fontWeight: 700,
                                        background: p.is_active ? 'rgba(16,185,129,0.15)' : 'rgba(148,163,184,0.15)',
                                        color: p.is_active ? '#10b981' : 'var(--text-muted)',
                                    }}>{p.is_active ? 'активна' : 'скрыта'}</span>
                                </td>
                                <td>
                                    {isAdmin && (
                                        <div style={{ display: 'flex', gap: 6 }}>
                                            <button className="action-btn" onClick={() => setEditing(p)}><Pencil size={13} /></button>
                                            <button className="action-btn" onClick={() => handleDelete(p)} style={{ color: '#ef4444' }}><Trash2 size={13} /></button>
                                        </div>
                                    )}
                                </td>
                            </tr>
                        ))}
                        {items.length === 0 && <tr><td colSpan="4" style={{ textAlign: 'center', padding: 20, color: 'var(--text-muted)' }}>Пусто</td></tr>}
                    </tbody>
                </table>
            )}

            {editing && (
                <CrudFormModal
                    title={editing.__new ? 'Новая программа' : 'Программа'}
                    fields={[
                        { name: 'name', label: 'Название *', type: 'text' },
                        { name: 'faculty', label: 'Факультет', type: 'text' },
                        ...(editing.__new ? [] : [{ name: 'is_active', label: 'Активна', type: 'checkbox' }]),
                    ]}
                    initial={editing}
                    busy={busy}
                    onCancel={() => setEditing(null)}
                    onSubmit={async (data) => {
                        setBusy(true);
                        try {
                            if (editing.__new) {
                                await api.createProgram({ name: data.name, faculty: data.faculty || '' });
                                showToast('Программа создана', 'success');
                            } else {
                                await api.updateProgram(editing.id, { name: data.name, faculty: data.faculty || '', is_active: data.is_active });
                                showToast('Программа обновлена', 'success');
                            }
                            setEditing(null);
                            load();
                        } catch (err) {
                            showToast('Ошибка: ' + err.message, 'error');
                        } finally {
                            setBusy(false);
                        }
                    }}
                />
            )}
        </div>
    );
}

// ── SOURCES (lead sources) ──────────────────────────────────────────────

function SourcesTab({ showToast, isAdmin }) {
    const [items, setItems] = useState([]);
    const [loading, setLoading] = useState(true);
    const [editing, setEditing] = useState(null);
    const [busy, setBusy] = useState(false);

    const load = useCallback(async () => {
        try {
            setLoading(true);
            setItems(await api.listSources());
        } catch (err) {
            showToast('Ошибка загрузки: ' + err.message, 'error');
        } finally { setLoading(false); }
    }, [showToast]);
    useEffect(() => { load(); }, [load]);

    const handleDelete = async (s) => {
        if (!window.confirm(`Удалить источник «${s.name}»?`)) return;
        try {
            const r = await api.deleteSource(s.id);
            showToast(r?.status === 'archived' ? 'Заархивирован' : 'Удалён', r?.status === 'archived' ? 'warning' : 'success');
            load();
        } catch (err) { showToast('Ошибка: ' + err.message, 'error'); }
    };

    return (
        <div className="users-table-wrap">
            <div className="users-table-header">
                <h3>Источники лидов</h3>
                {isAdmin && (
                    <button className="btn btn-primary" onClick={() => setEditing({ __new: true, name: '' })}>
                        <Plus size={14} /> Добавить
                    </button>
                )}
            </div>
            {loading ? <p style={{ padding: 14, color: 'var(--text-muted)' }}>Загрузка…</p> : (
                <table className="settings-table">
                    <thead><tr><th>Источник</th><th>Статус</th><th>Действия</th></tr></thead>
                    <tbody>
                        {items.map(s => (
                            <tr key={s.id} style={{ opacity: s.is_active ? 1 : 0.5 }}>
                                <td style={{ fontWeight: 600 }}>{s.name}</td>
                                <td>
                                    <span style={{
                                        padding: '3px 9px', borderRadius: 99, fontSize: 11, fontWeight: 700,
                                        background: s.is_active ? 'rgba(16,185,129,0.15)' : 'rgba(148,163,184,0.15)',
                                        color: s.is_active ? '#10b981' : 'var(--text-muted)',
                                    }}>{s.is_active ? 'активен' : 'скрыт'}</span>
                                </td>
                                <td>
                                    {isAdmin && (
                                        <div style={{ display: 'flex', gap: 6 }}>
                                            <button className="action-btn" onClick={() => setEditing(s)}><Pencil size={13} /></button>
                                            <button className="action-btn" onClick={() => handleDelete(s)} style={{ color: '#ef4444' }}><Trash2 size={13} /></button>
                                        </div>
                                    )}
                                </td>
                            </tr>
                        ))}
                        {items.length === 0 && <tr><td colSpan="3" style={{ textAlign: 'center', padding: 20, color: 'var(--text-muted)' }}>Пусто</td></tr>}
                    </tbody>
                </table>
            )}

            {editing && (
                <CrudFormModal
                    title={editing.__new ? 'Новый источник' : 'Источник'}
                    fields={[
                        { name: 'name', label: 'Название *', type: 'text' },
                        ...(editing.__new ? [] : [{ name: 'is_active', label: 'Активен', type: 'checkbox' }]),
                    ]}
                    initial={editing}
                    busy={busy}
                    onCancel={() => setEditing(null)}
                    onSubmit={async (data) => {
                        setBusy(true);
                        try {
                            if (editing.__new) {
                                await api.createSource({ name: data.name });
                                showToast('Источник создан', 'success');
                            } else {
                                await api.updateSource(editing.id, { name: data.name, is_active: data.is_active });
                                showToast('Источник обновлён', 'success');
                            }
                            setEditing(null);
                            load();
                        } catch (err) { showToast('Ошибка: ' + err.message, 'error'); }
                        finally { setBusy(false); }
                    }}
                />
            )}
        </div>
    );
}

// ── BOT settings ────────────────────────────────────────────────────────

const BOT_FIELDS = [
    { key: 'greeting', label: 'Приветствие (поддерживает {name})', textarea: true },
    { key: 'ask_name', label: 'Запрос ФИО', textarea: true },
    { key: 'ask_program', label: 'Запрос программы', textarea: true },
    { key: 'ask_phone', label: 'Запрос телефона', textarea: true },
    { key: 'handoff', label: 'После сбора анкеты', textarea: true },
    { key: 'faq_admission', label: 'FAQ: условия поступления', textarea: true },
    { key: 'faq_event', label: 'FAQ: регистрация на мероприятие', textarea: true },
    { key: 'faq_eet', label: 'FAQ: запись на EET тест', textarea: true },
];

function BotTab({ showToast, isAdmin }) {
    const [data, setData] = useState({});
    const [loading, setLoading] = useState(true);
    const [dirty, setDirty] = useState({});
    const [busy, setBusy] = useState(false);

    const load = useCallback(async () => {
        try {
            setLoading(true);
            const d = await api.getBotSettings();
            setData(d || {});
            setDirty({});
        } catch (err) { showToast('Ошибка: ' + err.message, 'error'); }
        finally { setLoading(false); }
    }, [showToast]);

    useEffect(() => { load(); }, [load]);

    const setField = (k, v) => {
        setData(prev => ({ ...prev, [k]: v }));
        setDirty(prev => ({ ...prev, [k]: true }));
    };

    const saveAll = async () => {
        setBusy(true);
        try {
            for (const k of Object.keys(dirty)) {
                if (dirty[k]) await api.setBotSetting(k, data[k] || '');
            }
            showToast('Тексты бота сохранены', 'success');
            setDirty({});
        } catch (err) { showToast('Ошибка: ' + err.message, 'error'); }
        finally { setBusy(false); }
    };

    if (loading) return <p style={{ padding: 14, color: 'var(--text-muted)' }}>Загрузка…</p>;

    return (
        <div className="users-table-wrap" style={{ padding: 18 }}>
            <div className="users-table-header" style={{ marginBottom: 14 }}>
                <h3>Тексты Telegram-бота</h3>
                {isAdmin && (
                    <button className="btn btn-primary" onClick={saveAll} disabled={busy || Object.values(dirty).every(v => !v)}>
                        <Save size={14} /> {busy ? 'Сохранение…' : 'Сохранить изменения'}
                    </button>
                )}
            </div>
            <p style={{ fontSize: 12, color: 'var(--text-muted)', marginBottom: 16 }}>
                В шаблоне приветствия можно использовать <code>{'{name}'}</code> — будет заменено на имя пользователя из Telegram.
                Изменения подхватятся ботом в течение 30 секунд (кэш).
            </p>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                {BOT_FIELDS.map(f => (
                    <div key={f.key} className="form-group">
                        <label>{f.label}</label>
                        <textarea
                            value={data[f.key] || ''}
                            onChange={e => setField(f.key, e.target.value)}
                            disabled={!isAdmin}
                            rows={4}
                            style={{
                                width: '100%', padding: 10, borderRadius: 8,
                                border: '1px solid var(--glass-border)', background: 'var(--glass)',
                                color: 'var(--text-primary)', fontSize: 13, fontFamily: 'inherit',
                                resize: 'vertical',
                            }}
                        />
                    </div>
                ))}
            </div>
        </div>
    );
}

// ── INTEGRATIONS ────────────────────────────────────────────────────────

const INT_STORAGE_KEY = 'crm_integrations';

function loadIntegrations() {
    try {
        const raw = localStorage.getItem(INT_STORAGE_KEY);
        if (!raw) return integrationsDefault;
        const parsed = JSON.parse(raw);
        return integrationsDefault.map(d => ({ ...d, on: parsed[d.key] !== undefined ? parsed[d.key] : d.on }));
    } catch { return integrationsDefault; }
}

function saveIntegrations(items) {
    const map = {};
    items.forEach(i => { map[i.key] = i.on; });
    localStorage.setItem(INT_STORAGE_KEY, JSON.stringify(map));
}

function IntegrationsTab({ showToast }) {
    const [items, setItems] = useState(() => loadIntegrations());

    const toggle = (idx) => {
        const next = items.map((it, i) => i === idx ? { ...it, on: !it.on } : it);
        setItems(next);
        saveIntegrations(next);
        showToast(`«${next[idx].name}» ${next[idx].on ? 'включена' : 'выключена'}`, 'success');
    };

    return (
        <div className="integration-grid">
            {items.map((integ, i) => {
                const Icon = integ.Icon;
                return (
                    <div className="integration-card" key={integ.key}>
                        <div className="integration-icon" style={{ background: integ.bg }}>
                            <Icon size={20} style={{ color: integ.color }} />
                        </div>
                        <div className="integration-info">
                            <h4>{integ.name}</h4>
                            <p>{integ.desc}</p>
                        </div>
                        <label className="toggle-switch">
                            <input type="checkbox" checked={integ.on} onChange={() => toggle(i)} />
                            <span className="toggle-slider"></span>
                        </label>
                    </div>
                );
            })}
        </div>
    );
}

// ── Reusable CRUD form modal ────────────────────────────────────────────

function CrudFormModal({ title, fields, initial, busy, onCancel, onSubmit }) {
    const [form, setForm] = useState(() => {
        const init = {};
        fields.forEach(f => { init[f.name] = initial[f.name] ?? (f.type === 'checkbox' ? (initial.id ? !!initial[f.name] : true) : ''); });
        if (initial.id) init.is_active = initial.is_active ?? true;
        return init;
    });
    return createPortal(
        <div className="modal-overlay open" onClick={(e) => { if (e.target === e.currentTarget) onCancel(); }}>
            <div className="modal" style={{ maxWidth: 460 }}>
                <div className="modal-header">
                    <h2>{title}</h2>
                    <button className="modal-close" onClick={onCancel}><X size={14} /></button>
                </div>
                <form onSubmit={(e) => { e.preventDefault(); onSubmit(form); }}>
                    <div className="modal-body">
                        <div className="form-grid">
                            {fields.map(f => (
                                <div key={f.name} className="form-group full-width">
                                    {f.type === 'checkbox' ? (
                                        <label style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13 }}>
                                            <input type="checkbox" checked={!!form[f.name]} onChange={e => setForm({ ...form, [f.name]: e.target.checked })} />
                                            {f.label}
                                        </label>
                                    ) : (
                                        <>
                                            <label>{f.label}</label>
                                            <input
                                                type={f.type || 'text'}
                                                value={form[f.name] || ''}
                                                onChange={e => setForm({ ...form, [f.name]: e.target.value })}
                                                required={f.label.includes('*')}
                                            />
                                        </>
                                    )}
                                </div>
                            ))}
                        </div>
                    </div>
                    <div className="modal-footer">
                        <button type="button" className="btn btn-outline" onClick={onCancel} disabled={busy}>Отмена</button>
                        <button type="submit" className="btn btn-primary" disabled={busy}>
                            <Save size={14} /> {busy ? 'Сохранение...' : 'Сохранить'}
                        </button>
                    </div>
                </form>
            </div>
        </div>,
        document.body
    );
}
