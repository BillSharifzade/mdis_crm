import { useState } from 'react';
import { X, Save } from 'lucide-react';
import { SOURCES, PROGRAMS, ENGLISH_SYSTEMS, PAYMENT_STATUSES, AVATAR_COLORS, randEl, getInitials, isMbaProgram } from '../data/crmData';
import { useUsers } from '../context/useUsers';
import { usePrograms } from '../context/usePrograms';

export default function LeadModal({ onClose, onSave, showToast }) {
    const { users } = useUsers();
    const { programs } = usePrograms();
    // Только реальные менеджеры; пусто — значит назначение уйдёт на round-robin.
    const managerOptions = users.filter(u => u.role === 'admin' || u.role === 'admissions');
    const programOptions = programs.length > 0
        ? programs.map(p => ({ id: p.id, name: p.name }))
        : PROGRAMS.map((name, i) => ({ id: -1 - i, name }));

    const [form, setForm] = useState({
        fio: '', phone: '', email: '',
        source: 'Сайт',
        programId: programOptions[0]?.id ?? null,
        assigneeId: null, // по умолчанию — round-robin на бэкенде
        socialUrl: '',
        englishSystem: '',
        englishScore: '',
        paymentStatus: '',
        workCompany: '',
        workPosition: '',
        reminderAt: '',
        reminderNote: '',
        comment: ''
    });

    const selectedProgramName = programOptions.find(p => p.id === form.programId)?.name || '';
    const showMba = isMbaProgram(selectedProgramName);

    const handleChange = (field, value) => {
        setForm(prev => ({ ...prev, [field]: value }));
    };

    const handleSave = () => {
        if (!form.fio || !form.phone) {
            showToast('Заполните обязательные поля', 'warning');
            return;
        }
        if (form.socialUrl && !/^https?:\/\//i.test(form.socialUrl.trim())) {
            showToast('Ссылка на соцсеть должна начинаться с http(s)://', 'warning');
            return;
        }
        const selectedProgram = programOptions.find(p => p.id === form.programId);
        const selectedManager = managerOptions.find(u => u.id === form.assigneeId);
        const englishLevel = form.englishSystem && form.englishScore
            ? `${form.englishSystem} ${form.englishScore}`
            : '';
        onSave({
            name: form.fio,
            initials: getInitials(form.fio),
            phone: form.phone,
            email: form.email,
            source: form.source,
            program: selectedProgram?.name || '',
            programId: selectedProgram?.id > 0 ? selectedProgram.id : null,
            assigneeId: selectedManager?.id > 0 ? selectedManager.id : null,
            socialUrl: form.socialUrl.trim(),
            englishLevel,
            paymentStatus: form.paymentStatus,
            workCompany: showMba ? form.workCompany.trim() : '',
            workPosition: showMba ? form.workPosition.trim() : '',
            reminderAt: form.reminderAt || '',
            reminderNote: form.reminderNote.trim(),
            comment: form.comment.trim(),
            status: 'new',
            date: new Date(),
            color: randEl(AVATAR_COLORS),
            interactions: [],
        });
        onClose();
    };

    return (
        <div className="modal-overlay open" onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}>
            <div className="modal">
                <div className="modal-header">
                    <h2>Новый лид</h2>
                    <button className="modal-close" onClick={onClose}><X size={14} /></button>
                </div>
                <div className="modal-body">
                    <div className="form-grid">
                        <div className="form-group">
                            <label>ФИО *</label>
                            <input type="text" placeholder="Введите полное имя" value={form.fio} onChange={e => handleChange('fio', e.target.value)} />
                        </div>
                        <div className="form-group">
                            <label>Телефон *</label>
                            <input type="tel" placeholder="+992 __ ___ __ __" value={form.phone} onChange={e => handleChange('phone', e.target.value)} />
                        </div>
                        <div className="form-group">
                            <label>Email</label>
                            <input type="email" placeholder="example@mail.com" value={form.email} onChange={e => handleChange('email', e.target.value)} />
                        </div>
                        <div className="form-group">
                            <label>Источник обращения</label>
                            <select value={form.source} onChange={e => handleChange('source', e.target.value)}>
                                {SOURCES.map(s => <option key={s}>{s}</option>)}
                            </select>
                        </div>
                        <div className="form-group">
                            <label>Программа обучения</label>
                            <select value={form.programId ?? ''} onChange={e => handleChange('programId', e.target.value ? parseInt(e.target.value, 10) : null)}>
                                {programOptions.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                            </select>
                        </div>
                        <div className="form-group">
                            <label>Ответственный менеджер</label>
                            <select value={form.assigneeId ?? ''} onChange={e => handleChange('assigneeId', e.target.value ? parseInt(e.target.value, 10) : null)}>
                                <option value="">— round-robin —</option>
                                {managerOptions.map(u => <option key={u.id} value={u.id}>{u.name}</option>)}
                            </select>
                        </div>
                        <div className="form-group">
                            <label>Статус оплаты</label>
                            <select value={form.paymentStatus} onChange={e => handleChange('paymentStatus', e.target.value)}>
                                {PAYMENT_STATUSES.map(p => <option key={p.key} value={p.key}>{p.label}</option>)}
                            </select>
                        </div>
                        {showMba && (
                            <>
                                <div className="form-group">
                                    <label>Компания (место работы)</label>
                                    <input type="text" placeholder="Название компании" value={form.workCompany} onChange={e => handleChange('workCompany', e.target.value)} />
                                </div>
                                <div className="form-group">
                                    <label>Должность</label>
                                    <input type="text" placeholder="Должность" value={form.workPosition} onChange={e => handleChange('workPosition', e.target.value)} />
                                </div>
                            </>
                        )}
                        <div className="form-group full-width">
                            <label>Ссылка на соцсеть (Instagram / VK / Facebook и т.п., кроме Telegram)</label>
                            <input
                                type="url"
                                placeholder="https://instagram.com/username"
                                value={form.socialUrl}
                                onChange={e => handleChange('socialUrl', e.target.value)}
                            />
                        </div>
                        <div className="form-group">
                            <label>Уровень английского</label>
                            <select
                                value={form.englishSystem}
                                onChange={e => setForm(prev => ({ ...prev, englishSystem: e.target.value, englishScore: '' }))}
                            >
                                <option value="">— не указан —</option>
                                {Object.keys(ENGLISH_SYSTEMS).map(sys => <option key={sys} value={sys}>{sys}</option>)}
                            </select>
                        </div>
                        <div className="form-group">
                            <label>Балл</label>
                            <select
                                value={form.englishScore}
                                onChange={e => handleChange('englishScore', e.target.value)}
                                disabled={!form.englishSystem}
                            >
                                <option value="">{form.englishSystem ? '— выберите балл —' : '— сначала система —'}</option>
                                {(ENGLISH_SYSTEMS[form.englishSystem] || []).map(v => <option key={v} value={v}>{v}</option>)}
                            </select>
                        </div>
                        <div className="form-group">
                            <label>Напомнить связаться (дата и время)</label>
                            <input type="datetime-local" value={form.reminderAt} onChange={e => handleChange('reminderAt', e.target.value)} />
                        </div>
                        <div className="form-group">
                            <label>Текст напоминания</label>
                            <input type="text" placeholder="Напр.: перезвонить по следующему набору" value={form.reminderNote} onChange={e => handleChange('reminderNote', e.target.value)} disabled={!form.reminderAt} />
                        </div>
                        <div className="form-group full-width">
                            <label>Комментарий</label>
                            <textarea placeholder="Дополнительные сведения о лиде..." value={form.comment} onChange={e => handleChange('comment', e.target.value)} rows="3"></textarea>
                        </div>
                    </div>
                </div>
                <div className="modal-footer">
                    <button className="btn btn-outline" onClick={onClose}>Отмена</button>
                    <button className="btn btn-primary" onClick={handleSave}><Save size={14} /> Сохранить</button>
                </div>
            </div>
        </div>
    );
}
