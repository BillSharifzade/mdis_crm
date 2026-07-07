import { useState, useEffect } from 'react';
import { X, Phone, MessageCircle, Mail, Pencil, Plus, Send, Trash2, Save, Link as LinkIcon } from 'lucide-react';
import { SOURCES, PROGRAMS, ENGLISH_SYSTEMS, PAYMENT_STATUSES, getStatusObj, formatHoursAgo, paymentStatusLabel, isMbaProgram } from '../data/crmData';
import { api, statusIdToKey } from '../services/api';
import TelegramChatModal from './TelegramChatModal';
import CallModal from './CallModal';
import { useUsers } from '../context/useUsers';
import { usePrograms } from '../context/usePrograms';
import { useStages } from '../context/useStages';

const iconMap = {
    'phone': Phone,
    'message-circle': MessageCircle,
    'mail': Mail,
};

const REASON_OPTIONS = ['Высокая цена', 'Выбрал другой вуз', 'Не прошёл по баллам', 'Передумал', 'Нет ответа'];

// ISO → значение для <input type="datetime-local"> (локальное время без TZ-суффикса).
function toLocalInput(iso) {
    if (!iso) return '';
    const d = new Date(iso);
    if (isNaN(d.getTime())) return '';
    const pad = n => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export default function DetailSidebar({ lead, onClose, onStatusChange, onUpdate, onDelete, showToast, role = 'admin', onNotif }) {
    const { stages, byKey: stagesByKey } = useStages();
    const stagesList = stages && stages.length > 0 ? stages : [];
    const stObj = stagesByKey.get(lead.status) || getStatusObj(lead.status);
    const { users, byId: usersById } = useUsers();
    const { programs } = usePrograms();
    const assignedUser = usersById.get(lead.assigneeId);
    const managerLabel = assignedUser?.name || 'Не назначен';

    const [interactions, setInteractions] = useState(lead.interactions || []);
    const [statusHistory, setStatusHistory] = useState([]);
    const [newNote, setNewNote] = useState('');
    const [reasonPrompt, setReasonPrompt] = useState(null);
    const [reasonText, setReasonText] = useState('');
    const [nowMs] = useState(() => Date.now());
    const [tgChatOpen, setTgChatOpen] = useState(false);
    const [callOpen, setCallOpen] = useState(false);
    const [editing, setEditing] = useState(false);
    const [reminderClosed, setReminderClosed] = useState(false);

    const canEdit = role !== 'guest';
    const canManage = role === 'admin' || role === 'admissions';

    useEffect(() => {
        if (!api.useApi) return;
        api.getInteractions(lead.id)
            .then(data => { if (data) setInteractions(data); })
            .catch(() => { });
        api.getStatusHistory(lead.id)
            .then(data => { if (data && Array.isArray(data)) setStatusHistory(data); })
            .catch(() => { });
    }, [lead.id]);

    const ensureConsultation = async () => {
        if (lead.status === 'new') {
            await onStatusChange(lead.id, 'consultation');
        }
    };

    const handleAddNote = async () => {
        const text = newNote.trim();
        if (!text) {
            showToast('Введите текст заметки', 'warning');
            return;
        }
        if (!api.useApi) {
            setInteractions(prev => [{
                id: Date.now(),
                icon: 'message-circle',
                bg: 'rgba(245, 158, 11, 0.15)',
                color: '#f59e0b',
                label: 'Заметка',
                note: text,
                hoursAgo: 0,
            }, ...prev]);
            setNewNote('');
            showToast('Заметка сохранена (демо)', 'success');
            return;
        }
        try {
            const newInt = await api.createInteraction(lead.id, 'manual_note', text);
            setInteractions(prev => [newInt, ...prev]);
            setNewNote('');
            onNotif && onNotif({ type: 'interaction', text: `<strong>Заметка</strong> к лиду «${lead.name}»`, leadId: lead.id });
            showToast('Заметка добавлена', 'success');
        } catch (error) {
            console.error(error);
            showToast('Не удалось сохранить заметку', 'error');
        }
    };

    const handleCall = () => {
        if (!canEdit) return;
        setCallOpen(true);
    };

    const handleCallSaved = async () => {
        // Перезагружаем историю — interactions появятся, ensureConsultation сместит статус
        try {
            const fresh = await api.getInteractions(lead.id);
            if (Array.isArray(fresh)) setInteractions(fresh);
        } catch { /* */ }
        await ensureConsultation();
        onNotif && onNotif({ type: 'interaction', text: `<strong>Звонок</strong> сохранён для лида «${lead.name}»`, leadId: lead.id });
    };

    const handleStatusClick = (statusKey) => {
        if (!canEdit) {
            showToast('У роли «guest» нет прав на изменение статуса', 'warning');
            return;
        }
        if (statusKey === lead.status) return;
        if (statusKey === 'lost') {
            setReasonPrompt({ status: statusKey });
            setReasonText('');
            return;
        }
        onStatusChange(lead.id, statusKey);
    };

    const confirmReason = async () => {
        if (!reasonText.trim()) return;
        const ok = await onStatusChange(lead.id, reasonPrompt.status, reasonText.trim());
        if (ok !== false) {
            setReasonPrompt(null);
            setReasonText('');
        }
    };

    const handleDelete = async () => {
        if (!window.confirm(`Удалить лида «${lead.name}»? Действие необратимо.`)) return;
        await onDelete(lead.id);
    };

    const leadDate = lead.date instanceof Date ? lead.date : new Date(lead.date);

    return (
        <>
            <div className="detail-overlay open" onClick={onClose}></div>
            <aside className="detail-sidebar open">
                <div id="detailContent">
                    <div className="detail-header">
                        <span style={{ fontSize: 14, fontWeight: 700, color: 'var(--text-primary)' }}>Карточка лида #{lead.id}</span>
                        <button className="detail-close" onClick={onClose}><X size={14} /></button>
                    </div>
                    <div className="detail-avatar-section">
                        <div className="detail-big-avatar" style={{ background: lead.color }}>{lead.initials}</div>
                        <div style={{ minWidth: 0, flex: 1 }}>
                            <div className="detail-avatar-name">{lead.name}</div>
                            <div className="detail-avatar-program">{lead.program}</div>
                            <div style={{ marginTop: 6 }}>
                                <span className={`status-badge ${stObj.cls}`}>{stObj.label}</span>
                                {lead.refusalReason && (
                                    <span style={{ marginLeft: 8, fontSize: 11, color: '#ef4444' }}>
                                        Причина: {lead.refusalReason}
                                    </span>
                                )}
                            </div>
                        </div>
                        {canEdit && (
                            <div style={{ display: 'flex', flexDirection: 'column', gap: 6, marginLeft: 'auto' }}>
                                <button className="btn btn-outline" onClick={() => setEditing(true)}>
                                    <Pencil size={14} /> Редактировать
                                </button>
                                {canManage && (
                                    <button className="btn btn-outline" onClick={handleDelete} style={{ color: '#ef4444', borderColor: '#ef444440' }}>
                                        <Trash2 size={14} /> Удалить
                                    </button>
                                )}
                            </div>
                        )}
                    </div>

                    <div className="detail-section">
                        <div className="detail-section-title">Контактные данные</div>
                        <div className="detail-fields">
                            <div className="detail-field"><label>Телефон</label><span>{lead.phone || '—'}</span></div>
                            <div className="detail-field"><label>Email</label><span>{lead.email || '—'}</span></div>
                            <div className="detail-field"><label>Источник</label><span>{lead.source}</span></div>
                            <div className="detail-field"><label>Дата заявки</label><span>{leadDate.toLocaleDateString('ru-RU')}</span></div>
                            <div className="detail-field"><label>Программа</label><span>{lead.program}</span></div>
                            <div className="detail-field"><label>Менеджер</label><span>{managerLabel}</span></div>
                            <div className="detail-field"><label>Статус оплаты</label><span>{lead.paymentStatus ? paymentStatusLabel(lead.paymentStatus) : '—'}</span></div>
                            <div className="detail-field"><label>Английский</label><span>{lead.englishLevel || '—'}</span></div>
                            {isMbaProgram(lead.program) && (
                                <>
                                    <div className="detail-field"><label>Компания</label><span>{lead.workCompany || '—'}</span></div>
                                    <div className="detail-field"><label>Должность</label><span>{lead.workPosition || '—'}</span></div>
                                </>
                            )}
                            {lead.reminderAt && !lead.reminderDone && !reminderClosed && (
                                <div className="detail-field" style={{ gridColumn: 'span 2' }}>
                                    <label>Напоминание связаться</label>
                                    <span style={{ color: new Date(lead.reminderAt) <= new Date() ? '#f59e0b' : 'var(--text-primary)' }}>
                                        {new Date(lead.reminderAt).toLocaleString('ru-RU', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' })}
                                        {lead.reminderNote ? ` · ${lead.reminderNote}` : ''}
                                        {' '}
                                        <button
                                            type="button"
                                            onClick={async () => { if (api.useApi) { try { await api.completeReminder(lead.id); } catch { /* */ } } setReminderClosed(true); showToast('Напоминание закрыто', 'success'); }}
                                            style={{ marginLeft: 6, fontSize: 11, padding: '2px 8px', borderRadius: 6, border: '1px solid var(--glass-border)', background: 'var(--glass)', color: 'var(--text-secondary)', cursor: 'pointer' }}
                                        >Выполнено</button>
                                    </span>
                                </div>
                            )}
                            <div className="detail-field" style={{ gridColumn: 'span 2' }}>
                                <label>Соцсеть</label>
                                <span>
                                    {lead.socialUrl
                                        ? <a href={lead.socialUrl} target="_blank" rel="noopener noreferrer" style={{ color: '#a78bfa', display: 'inline-flex', alignItems: 'center', gap: 4 }}>
                                              <LinkIcon size={12} /> {lead.socialUrl}
                                          </a>
                                        : '—'
                                    }
                                </span>
                            </div>
                        </div>
                    </div>

                    <div className="detail-section">
                        <div className="detail-section-title">Этапы воронки</div>
                        <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
                            {stagesList.map(s => (
                                <button
                                    key={s.key}
                                    onClick={() => handleStatusClick(s.key)}
                                    disabled={!canEdit}
                                    className="status-pill"
                                    style={{
                                        padding: '5px 12px', borderRadius: 99, fontSize: 11, fontWeight: 700,
                                        border: `1px solid ${s.color}30`,
                                        background: lead.status === s.key ? `${s.color}20` : 'transparent',
                                        color: lead.status === s.key ? s.color : 'var(--text-muted)',
                                        cursor: canEdit ? 'pointer' : 'not-allowed',
                                        opacity: canEdit ? 1 : 0.6,
                                        fontFamily: 'inherit',
                                    }}
                                >
                                    {s.label}
                                </button>
                            ))}
                        </div>
                    </div>

                    <div className="detail-section">
                        <div className="detail-section-title">История взаимодействий</div>

                        {canEdit && (
                            <div style={{ marginBottom: '16px', display: 'flex', gap: '8px' }}>
                                <input
                                    type="text"
                                    placeholder="Написать заметку..."
                                    value={newNote}
                                    onChange={e => setNewNote(e.target.value)}
                                    style={{ flex: 1, padding: '8px 12px', borderRadius: '8px', border: '1px solid var(--glass-border)', background: 'var(--glass)', color: 'var(--text-primary)', fontSize: '13px' }}
                                    onKeyDown={e => { if (e.key === 'Enter') handleAddNote(); }}
                                />
                                <button
                                    className="btn btn-primary icon-only"
                                    onClick={handleAddNote}
                                    title="Добавить заметку"
                                    style={{ width: 38, height: 38, padding: 0, minWidth: 38, display: 'flex', alignItems: 'center', justifyContent: 'center' }}
                                >
                                    <Plus size={16} />
                                </button>
                            </div>
                        )}

                        <div className="interaction-list">
                            {interactions.length > 0 ? interactions.map((inter, i) => {
                                const IconComp = iconMap[inter.icon] || Phone;
                                return (
                                    <div className="interaction-item" key={inter.id || i}>
                                        <div className="interaction-icon" style={{ background: inter.bg }}>
                                            <IconComp size={13} style={{ color: inter.color }} />
                                        </div>
                                        <div style={{ minWidth: 0, flex: 1 }}>
                                            <span className="interaction-text">{inter.note}</span>
                                            <span className="interaction-time">{inter.label} · {formatHoursAgo(inter.hoursAgo)}</span>
                                        </div>
                                    </div>
                                );
                            }) : (
                                <p style={{ color: 'var(--text-muted)', fontSize: 13, textAlign: 'center', margin: '20px 0' }}>Взаимодействий пока нет</p>
                            )}
                        </div>
                    </div>

                    <div className="detail-section">
                        <div className="detail-section-title">История изменения статусов</div>
                        <div className="interaction-list">
                            {statusHistory.length > 0 ? statusHistory.map((h, i) => {
                                const stKey = statusIdToKey(h.new_status_id);
                                const stObj = getStatusObj(stKey);
                                const createdTime = new Date(h.created_at).getTime();
                                const hoursAgo = Math.max(0, Math.floor((nowMs - createdTime) / (1000 * 60 * 60)));
                                return (
                                    <div className="interaction-item" key={h.id || i}>
                                        <div className="interaction-icon" style={{ background: `${stObj.color}20` }}>
                                            <div style={{ width: 8, height: 8, borderRadius: '50%', background: stObj.color }}></div>
                                        </div>
                                        <div>
                                            <span className="interaction-text">Перевод в статус «{stObj.label}»</span>
                                            <span className="interaction-time">{h.refusal_reason ? `Причина: ${h.refusal_reason} · ` : ''}{formatHoursAgo(hoursAgo)}</span>
                                        </div>
                                    </div>
                                );
                            }) : (
                                <p style={{ color: 'var(--text-muted)', fontSize: 13, textAlign: 'center', margin: '20px 0' }}>История пуста</p>
                            )}
                        </div>
                    </div>

                    <div style={{ padding: '16px 24px', display: 'flex', flexDirection: 'column', gap: 8 }}>
                        <button
                            className="btn btn-primary"
                            style={{
                                background: 'linear-gradient(135deg, #0088cc, #229ED9)',
                                color: '#fff', justifyContent: 'center',
                            }}
                            onClick={() => setTgChatOpen(true)}
                        >
                            <Send size={14} /> Чат в Telegram
                        </button>
                        {canEdit && (
                            <button className="btn btn-outline" onClick={handleCall}>
                                <Phone size={14} /> Позвонить
                            </button>
                        )}
                    </div>

                    {tgChatOpen && (
                        <TelegramChatModal
                            lead={lead}
                            onClose={() => setTgChatOpen(false)}
                            showToast={showToast}
                            role={role}
                            onManagerMessage={ensureConsultation}
                        />
                    )}

                    {callOpen && (
                        <CallModal
                            lead={lead}
                            onClose={() => setCallOpen(false)}
                            onSaved={handleCallSaved}
                            showToast={showToast}
                        />
                    )}
                </div>
            </aside>

            {reasonPrompt && (
                <div className="modal-overlay open" style={{ zIndex: 1100 }} onClick={(e) => { if (e.target === e.currentTarget) setReasonPrompt(null); }}>
                    <div className="modal" style={{ maxWidth: 480 }}>
                        <div className="modal-header">
                            <h2>Причина отказа</h2>
                        </div>
                        <div className="modal-body">
                            <p style={{ color: 'var(--text-secondary)', fontSize: 13, marginBottom: 12 }}>
                                Укажите причину для статуса «Отказ» (обязательное поле по ТЗ, п. 6.5).
                            </p>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 12 }}>
                                {REASON_OPTIONS.map(opt => (
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

            {editing && (
                <EditLeadModal
                    lead={lead}
                    users={users}
                    programs={programs}
                    onClose={() => setEditing(false)}
                    onSave={async (patch) => {
                        const ok = await onUpdate(lead.id, patch);
                        if (ok) setEditing(false);
                    }}
                />
            )}
        </>
    );
}

function EditLeadModal({ lead, users, programs, onClose, onSave }) {
    const eligibleManagers = users.filter(u => u.role === 'admin' || u.role === 'admissions');
    const programOptions = programs && programs.length > 0
        ? programs.map(p => ({ id: p.id, name: p.name }))
        : PROGRAMS.map((name, i) => ({ id: -1 - i, name }));
    const initialProgramId = lead.programId
        || programOptions.find(p => p.name === lead.program)?.id
        || programOptions[0]?.id
        || null;

    // Разбираем сохранённый уровень «SYS SCORE» (например «IELTS 6.0») на части.
    const [engSys0, ...engRest] = (lead.englishLevel || '').trim().split(/\s+/);
    const initialEnglishSystem = ENGLISH_SYSTEMS[engSys0] ? engSys0 : '';
    const initialEnglishScore = initialEnglishSystem ? engRest.join(' ') : '';

    const [form, setForm] = useState({
        name: lead.name || '',
        phone: lead.phone || '',
        email: lead.email || '',
        programId: initialProgramId,
        source: lead.source || SOURCES[0],
        assigneeId: lead.assigneeId || (users[0] && users[0].id) || null,
        socialUrl: lead.socialUrl || '',
        englishSystem: initialEnglishSystem,
        englishScore: initialEnglishScore,
        paymentStatus: lead.paymentStatus || '',
        workCompany: lead.workCompany || '',
        workPosition: lead.workPosition || '',
        reminderAt: toLocalInput(lead.reminderDone ? '' : lead.reminderAt),
        reminderNote: lead.reminderDone ? '' : (lead.reminderNote || ''),
    });
    const [busy, setBusy] = useState(false);
    const editProgramName = programOptions.find(p => p.id === form.programId)?.name || '';
    const showMba = isMbaProgram(editProgramName);
    const hadReminder = !!lead.reminderAt && !lead.reminderDone;

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!form.name.trim()) return;
        if (form.socialUrl && !/^https?:\/\//i.test(form.socialUrl.trim())) {
            alert('Ссылка на соцсеть должна начинаться с http(s)://');
            return;
        }
        const chosenProgram = programOptions.find(p => p.id === form.programId);
        setBusy(true);
        const englishLevel = form.englishSystem && form.englishScore
            ? `${form.englishSystem} ${form.englishScore}`
            : '';
        const patch = {
            ...form,
            socialUrl: form.socialUrl.trim(),
            englishLevel,
            program: chosenProgram?.name || lead.program,
            programId: chosenProgram?.id > 0 ? chosenProgram.id : null,
            paymentStatus: form.paymentStatus,
            workCompany: showMba ? form.workCompany.trim() : '',
            workPosition: showMba ? form.workPosition.trim() : '',
        };
        // Напоминание: снятое поле → очистка, заполненное → (пере)установка.
        if (!form.reminderAt && hadReminder) {
            patch.clearReminder = true;
        } else if (form.reminderAt) {
            patch.reminderAt = form.reminderAt;
            patch.reminderNote = form.reminderNote.trim();
        }
        await onSave(patch);
        setBusy(false);
    };

    return (
        <div className="modal-overlay open" style={{ zIndex: 1100 }} onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}>
            <div className="modal" style={{ maxWidth: 520 }}>
                <div className="modal-header">
                    <h2>Редактировать лида</h2>
                    <button className="modal-close" onClick={onClose}><X size={14} /></button>
                </div>
                <form onSubmit={handleSubmit}>
                    <div className="modal-body">
                        <div className="form-grid">
                            <div className="form-group full-width">
                                <label>ФИО *</label>
                                <input type="text" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} />
                            </div>
                            <div className="form-group">
                                <label>Телефон</label>
                                <input type="tel" value={form.phone} onChange={e => setForm({ ...form, phone: e.target.value })} />
                            </div>
                            <div className="form-group">
                                <label>Email</label>
                                <input type="email" value={form.email} onChange={e => setForm({ ...form, email: e.target.value })} />
                            </div>
                            <div className="form-group">
                                <label>Источник</label>
                                <select value={form.source} onChange={e => setForm({ ...form, source: e.target.value })}>
                                    {SOURCES.map(s => <option key={s}>{s}</option>)}
                                </select>
                            </div>
                            <div className="form-group">
                                <label>Программа</label>
                                <select value={form.programId ?? ''} onChange={e => setForm({ ...form, programId: e.target.value ? parseInt(e.target.value, 10) : null })}>
                                    {programOptions.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                                </select>
                            </div>
                            <div className="form-group full-width">
                                <label>Менеджер</label>
                                <select value={form.assigneeId || ''} onChange={e => setForm({ ...form, assigneeId: e.target.value ? parseInt(e.target.value, 10) : null })}>
                                    <option value="">— не назначен —</option>
                                    {eligibleManagers.map(u => <option key={u.id} value={u.id}>{u.name}</option>)}
                                </select>
                            </div>
                            <div className="form-group">
                                <label>Английский</label>
                                <select
                                    value={form.englishSystem}
                                    onChange={e => setForm({ ...form, englishSystem: e.target.value, englishScore: '' })}
                                >
                                    <option value="">— не указан —</option>
                                    {Object.keys(ENGLISH_SYSTEMS).map(sys => <option key={sys} value={sys}>{sys}</option>)}
                                </select>
                            </div>
                            <div className="form-group">
                                <label>Балл</label>
                                <select
                                    value={form.englishScore}
                                    onChange={e => setForm({ ...form, englishScore: e.target.value })}
                                    disabled={!form.englishSystem}
                                >
                                    <option value="">{form.englishSystem ? '— выберите —' : '— система —'}</option>
                                    {(ENGLISH_SYSTEMS[form.englishSystem] || []).map(v => <option key={v} value={v}>{v}</option>)}
                                </select>
                            </div>
                            <div className="form-group">
                                <label>Статус оплаты</label>
                                <select value={form.paymentStatus} onChange={e => setForm({ ...form, paymentStatus: e.target.value })}>
                                    {PAYMENT_STATUSES.map(p => <option key={p.key} value={p.key}>{p.label}</option>)}
                                </select>
                            </div>
                            {showMba && (
                                <>
                                    <div className="form-group">
                                        <label>Компания</label>
                                        <input type="text" placeholder="Место работы" value={form.workCompany} onChange={e => setForm({ ...form, workCompany: e.target.value })} />
                                    </div>
                                    <div className="form-group">
                                        <label>Должность</label>
                                        <input type="text" placeholder="Должность" value={form.workPosition} onChange={e => setForm({ ...form, workPosition: e.target.value })} />
                                    </div>
                                </>
                            )}
                            <div className="form-group">
                                <label>Напомнить связаться</label>
                                <input type="datetime-local" value={form.reminderAt} onChange={e => setForm({ ...form, reminderAt: e.target.value })} />
                            </div>
                            <div className="form-group">
                                <label>Текст напоминания</label>
                                <input type="text" placeholder="Комментарий к напоминанию" value={form.reminderNote} onChange={e => setForm({ ...form, reminderNote: e.target.value })} disabled={!form.reminderAt} />
                            </div>
                            <div className="form-group full-width">
                                <label>Соцсеть (URL)</label>
                                <input
                                    type="url"
                                    placeholder="https://instagram.com/username"
                                    value={form.socialUrl}
                                    onChange={e => setForm({ ...form, socialUrl: e.target.value })}
                                />
                            </div>
                        </div>
                    </div>
                    <div className="modal-footer">
                        <button type="button" className="btn btn-outline" onClick={onClose} disabled={busy}>Отмена</button>
                        <button type="submit" className="btn btn-primary" disabled={busy}>
                            <Save size={14} /> {busy ? 'Сохранение...' : 'Сохранить'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}
