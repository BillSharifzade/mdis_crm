import { useState } from 'react';
import { X, Save } from 'lucide-react';
import { SOURCES, PROGRAMS, MANAGER_NAMES, AVATAR_COLORS, randEl, getInitials } from '../data/crmData';
import { useUsers } from '../context/useUsers';
import { usePrograms } from '../context/usePrograms';

export default function LeadModal({ onClose, onSave, showToast }) {
    const { users } = useUsers();
    const { programs } = usePrograms();
    const eligibleManagers = users.filter(u => u.role === 'admin' || u.role === 'admissions');
    const managerOptions = eligibleManagers.length > 0 ? eligibleManagers : MANAGER_NAMES.map((n, i) => ({ id: -1 - i, name: n }));
    const programOptions = programs.length > 0
        ? programs.map(p => ({ id: p.id, name: p.name }))
        : PROGRAMS.map((name, i) => ({ id: -1 - i, name }));

    const [form, setForm] = useState({
        fio: '', phone: '', email: '',
        source: 'Сайт',
        programId: programOptions[0]?.id ?? null,
        assigneeId: managerOptions[0]?.id ?? null,
        socialUrl: '',
        comment: ''
    });

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
            status: 'new',
            managerIdx: Math.max(0, MANAGER_NAMES.indexOf(selectedManager?.name)),
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
                        <div className="form-group full-width">
                            <label>Ссылка на соцсеть (Instagram / VK / Facebook и т.п., кроме Telegram)</label>
                            <input
                                type="url"
                                placeholder="https://instagram.com/username"
                                value={form.socialUrl}
                                onChange={e => handleChange('socialUrl', e.target.value)}
                            />
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
