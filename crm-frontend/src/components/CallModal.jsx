import { useState } from 'react';
import { createPortal } from 'react-dom';
import { Phone, PhoneCall, PhoneOff, X, Check, Save } from 'lucide-react';
import { api } from '../services/api';

/**
 * Диалог звонка по лиду (T9).
 *
 * Шаг 1: показываем подсвеченный телефон и две кнопки «Ответил / Не ответил».
 * Шаг 2: появляется поле комментария. Для «Ответил» — ещё поле длительности (мин).
 * Сохранение — POST /interactions с outcome + duration_minutes.
 */
export default function CallModal({ lead, onClose, onSaved, showToast }) {
    const [outcome, setOutcome] = useState(''); // '', 'answered', 'no_answer'
    const [comment, setComment] = useState('');
    const [minutes, setMinutes] = useState('');
    const [busy, setBusy] = useState(false);

    const phone = (lead.phone || '').trim();

    const callHref = phone ? `tel:${phone.replace(/[^+\d]/g, '')}` : null;

    const onPickOutcome = (val) => {
        setOutcome(val);
        if (callHref) {
            try { window.open(callHref, '_self'); } catch { /* */ }
        }
    };

    const handleSave = async () => {
        if (!outcome) return;
        if (!comment.trim()) {
            showToast && showToast('Добавьте комментарий', 'warning');
            return;
        }
        if (outcome === 'answered' && (!minutes || Number(minutes) <= 0)) {
            showToast && showToast('Укажите длительность в минутах', 'warning');
            return;
        }
        setBusy(true);
        try {
            await api.logCall({
                leadId: lead.id,
                outcome,
                comment: comment.trim(),
                durationMinutes: outcome === 'answered' ? Number(minutes) : 0,
            });
            showToast && showToast('Звонок сохранён', 'success');
            onSaved && onSaved();
            onClose();
        } catch (err) {
            console.error(err);
            showToast && showToast('Ошибка сохранения: ' + err.message, 'error');
        } finally {
            setBusy(false);
        }
    };

    return createPortal(
        <div className="modal-overlay open" style={{ zIndex: 1100 }} onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}>
            <div className="modal" style={{ maxWidth: 480 }}>
                <div className="modal-header">
                    <h2>Звонок лиду</h2>
                    <button className="modal-close" onClick={onClose}><X size={14} /></button>
                </div>
                <div className="modal-body">
                    <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
                        <div style={{
                            width: 44, height: 44, borderRadius: 99,
                            background: 'linear-gradient(135deg, #6366f1, #a855f7)',
                            display: 'flex', alignItems: 'center', justifyContent: 'center',
                        }}>
                            <Phone size={20} style={{ color: '#fff' }} />
                        </div>
                        <div style={{ flex: 1 }}>
                            <div style={{ fontSize: 13, color: 'var(--text-muted)' }}>{lead.name}</div>
                            <a
                                href={callHref || '#'}
                                onClick={e => { if (!callHref) e.preventDefault(); }}
                                style={{
                                    fontSize: 22, fontWeight: 700, color: '#a78bfa',
                                    letterSpacing: '0.04em', textDecoration: 'none',
                                    background: 'rgba(167,139,250,0.12)',
                                    padding: '4px 10px', borderRadius: 8,
                                    display: 'inline-block',
                                }}
                            >
                                {phone || 'Телефон не указан'}
                            </a>
                        </div>
                    </div>

                    {!outcome && (
                        <div style={{ display: 'flex', gap: 10, marginBottom: 12 }}>
                            <button
                                className="btn btn-primary"
                                style={{ flex: 1, justifyContent: 'center', background: 'linear-gradient(135deg,#10b981,#22c55e)' }}
                                onClick={() => onPickOutcome('answered')}
                            >
                                <PhoneCall size={14} /> Ответил
                            </button>
                            <button
                                className="btn btn-outline"
                                style={{ flex: 1, justifyContent: 'center', color: '#f59e0b', borderColor: '#f59e0b40' }}
                                onClick={() => onPickOutcome('no_answer')}
                            >
                                <PhoneOff size={14} /> Не ответил
                            </button>
                        </div>
                    )}

                    {outcome && (
                        <>
                            <div style={{
                                padding: '8px 12px', borderRadius: 8, marginBottom: 12,
                                background: outcome === 'answered' ? 'rgba(16,185,129,0.1)' : 'rgba(245,158,11,0.1)',
                                color: outcome === 'answered' ? '#10b981' : '#f59e0b',
                                fontSize: 12, fontWeight: 600, display: 'flex', alignItems: 'center', gap: 6,
                            }}>
                                <Check size={12} /> {outcome === 'answered' ? 'Звонок состоялся' : 'Без ответа'}
                                <button
                                    onClick={() => { setOutcome(''); setComment(''); setMinutes(''); }}
                                    style={{
                                        marginLeft: 'auto', background: 'transparent', border: 0,
                                        color: 'inherit', cursor: 'pointer', fontSize: 11, textDecoration: 'underline',
                                    }}
                                >изменить</button>
                            </div>

                            <div className="form-group" style={{ marginBottom: 10 }}>
                                <label>Комментарий *</label>
                                <textarea
                                    rows={3}
                                    placeholder="О чём говорили / почему не ответил..."
                                    value={comment}
                                    onChange={e => setComment(e.target.value)}
                                    style={{
                                        width: '100%', padding: 10, borderRadius: 8,
                                        border: '1px solid var(--glass-border)',
                                        background: 'var(--glass)', color: 'var(--text-primary)',
                                        fontSize: 13, fontFamily: 'inherit', resize: 'vertical',
                                    }}
                                />
                            </div>

                            {outcome === 'answered' && (
                                <div className="form-group">
                                    <label>Длительность, мин *</label>
                                    <input
                                        type="number" min="1" step="1"
                                        placeholder="3"
                                        value={minutes}
                                        onChange={e => setMinutes(e.target.value)}
                                        style={{
                                            width: '100%', padding: 10, borderRadius: 8,
                                            border: '1px solid var(--glass-border)',
                                            background: 'var(--glass)', color: 'var(--text-primary)',
                                            fontSize: 13, fontFamily: 'inherit',
                                        }}
                                    />
                                </div>
                            )}
                        </>
                    )}
                </div>
                <div className="modal-footer">
                    <button className="btn btn-outline" onClick={onClose} disabled={busy}>Отмена</button>
                    <button
                        className="btn btn-primary"
                        onClick={handleSave}
                        disabled={!outcome || busy}
                    >
                        <Save size={14} /> {busy ? 'Сохранение...' : 'Сохранить звонок'}
                    </button>
                </div>
            </div>
        </div>,
        document.body
    );
}
