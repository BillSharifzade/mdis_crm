import { UserPlus, Phone, Clock, Bell, CheckCircle, ArrowRightCircle, GitMerge, MessageCircle, Mail, FileDown, Trash2, Check, X } from 'lucide-react';
import { useNotif } from '../context/useNotif';
import { api } from '../services/api';

const TYPE_META = {
    'lead_created': { Icon: UserPlus, bg: 'rgba(99,102,241,0.15)', col: '#818cf8' },
    'reminder': { Icon: Clock, bg: 'rgba(245,158,11,0.18)', col: '#f59e0b' },
    'status_change': { Icon: ArrowRightCircle, bg: 'rgba(6,182,212,0.15)', col: '#06b6d4' },
    'lead_merged': { Icon: GitMerge, bg: 'rgba(168,85,247,0.15)', col: '#a855f7' },
    'interaction': { Icon: MessageCircle, bg: 'rgba(16,185,129,0.15)', col: '#10b981' },
    'export': { Icon: FileDown, bg: 'rgba(245,158,11,0.15)', col: '#f59e0b' },
    'enrolled': { Icon: CheckCircle, bg: 'rgba(16,185,129,0.15)', col: '#10b981' },
    'lost': { Icon: Bell, bg: 'rgba(239,68,68,0.15)', col: '#ef4444' },
    'telegram_chat': { Icon: MessageCircle, bg: 'rgba(6,182,212,0.15)', col: '#06b6d4' },
    'user_action': { Icon: Phone, bg: 'rgba(168,85,247,0.15)', col: '#a855f7' },
};

function relativeTime(iso) {
    const diff = Date.now() - new Date(iso).getTime();
    if (diff < 60_000) return 'только что';
    if (diff < 3600_000) return `${Math.floor(diff / 60_000)} мин назад`;
    if (diff < 86_400_000) return `${Math.floor(diff / 3600_000)} ч назад`;
    return `${Math.floor(diff / 86_400_000)} дн назад`;
}

export default function NotifPanel({ isOpen, onClose, onOpenLead, showToast }) {
    const { notifs, markAllRead, remove, clearAll } = useNotif();

    if (!isOpen) return null;

    // Клик по уведомлению открывает карточку лида (#4 ТЗ) и помечает прочитанным.
    const handleOpen = (n) => {
        if (n.leadId && onOpenLead) {
            onOpenLead(n.leadId);
            onClose();
        }
    };

    // «Выполнено» — для напоминаний закрывает его на бэкенде.
    const handleDone = async (e, n) => {
        e.stopPropagation();
        if (n.type === 'reminder' && n.leadId && api.useApi) {
            try { await api.completeReminder(n.leadId); }
            catch { showToast && showToast('Не удалось закрыть напоминание', 'error'); }
        }
        remove(n.id);
    };

    const handleClose = (e, n) => {
        e.stopPropagation();
        remove(n.id);
    };

    return (
        <>
            <div style={{ position: 'fixed', inset: 0, zIndex: 99 }} onClick={onClose}></div>
            <div className={`notif-panel ${isOpen ? 'open' : ''}`}>
                <div className="notif-panel-header">
                    <h3>Уведомления</h3>
                    <div style={{ display: 'flex', gap: 4 }}>
                        <button onClick={markAllRead} title="Отметить всё прочитанным">Прочитать все</button>
                        <button onClick={clearAll} title="Очистить" style={{ color: 'var(--text-muted)' }}>
                            <Trash2 size={12} />
                        </button>
                    </div>
                </div>
                <div className="notif-list">
                    {notifs.length === 0 && (
                        <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-muted)', fontSize: 13 }}>
                            Пока нет уведомлений
                        </div>
                    )}
                    {notifs.map(n => {
                        const meta = TYPE_META[n.type] || TYPE_META.user_action;
                        const Icon = meta.Icon;
                        const clickable = !!n.leadId;
                        return (
                            <div
                                className={`notif-item ${n.unread ? 'notif-unread' : ''}`}
                                key={n.id}
                                onClick={() => handleOpen(n)}
                                style={{ cursor: clickable ? 'pointer' : 'default', alignItems: 'flex-start' }}
                            >
                                <div className="notif-item-icon" style={{ background: meta.bg }}>
                                    <Icon size={14} style={{ color: meta.col }} />
                                </div>
                                <div style={{ flex: 1, minWidth: 0 }}>
                                    <span className="notif-item-text" dangerouslySetInnerHTML={{ __html: n.text }} />
                                    <span className="notif-item-time">{relativeTime(n.createdAt)}</span>
                                    <div style={{ display: 'flex', gap: 6, marginTop: 6 }}>
                                        {n.type === 'reminder' && (
                                            <button
                                                onClick={(e) => handleDone(e, n)}
                                                title="Отметить выполненным"
                                                style={{ display: 'inline-flex', alignItems: 'center', gap: 3, fontSize: 11, padding: '3px 8px', borderRadius: 6, border: '1px solid rgba(16,185,129,0.4)', background: 'rgba(16,185,129,0.12)', color: '#10b981', cursor: 'pointer' }}
                                            >
                                                <Check size={11} /> Выполнено
                                            </button>
                                        )}
                                        <button
                                            onClick={(e) => handleClose(e, n)}
                                            title="Закрыть уведомление"
                                            style={{ display: 'inline-flex', alignItems: 'center', gap: 3, fontSize: 11, padding: '3px 8px', borderRadius: 6, border: '1px solid var(--glass-border)', background: 'var(--glass)', color: 'var(--text-secondary)', cursor: 'pointer' }}
                                        >
                                            <X size={11} /> Закрыть
                                        </button>
                                    </div>
                                </div>
                                {n.unread && <div className="notif-unread-dot"></div>}
                            </div>
                        );
                    })}
                </div>
            </div>
        </>
    );
}
