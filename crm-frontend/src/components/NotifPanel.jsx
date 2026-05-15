import { UserPlus, Phone, Clock, Bell, CheckCircle, ArrowRightCircle, GitMerge, MessageCircle, Mail, FileDown, Trash2 } from 'lucide-react';
import { useNotif } from '../context/useNotif';

const TYPE_META = {
    'lead_created': { Icon: UserPlus, bg: 'rgba(99,102,241,0.15)', col: '#818cf8' },
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

export default function NotifPanel({ isOpen, onClose }) {
    const { notifs, markAllRead, clearAll } = useNotif();

    if (!isOpen) return null;

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
                        return (
                            <div className={`notif-item ${n.unread ? 'notif-unread' : ''}`} key={n.id}>
                                <div className="notif-item-icon" style={{ background: meta.bg }}>
                                    <Icon size={14} style={{ color: meta.col }} />
                                </div>
                                <div style={{ flex: 1, minWidth: 0 }}>
                                    <span className="notif-item-text" dangerouslySetInnerHTML={{ __html: n.text }} />
                                    <span className="notif-item-time">{relativeTime(n.createdAt)}</span>
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
