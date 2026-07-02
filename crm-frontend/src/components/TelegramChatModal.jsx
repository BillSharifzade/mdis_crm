import { useState, useEffect, useRef, useCallback } from 'react';
import { X, Send, Bot, User, UserCheck, Loader2 } from 'lucide-react';
import { api } from '../services/api';

const STATE_LABELS = {
    greet: 'Бот приветствует',
    ask_name: 'Бот собирает ФИО',
    ask_program: 'Бот собирает программу',
    ask_english: 'Бот собирает уровень английского',
    ask_phone: 'Бот собирает телефон',
    manager: 'Подключён менеджер',
};

function formatTime(iso) {
    try {
        const d = new Date(iso);
        return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
    } catch {
        return '';
    }
}

function formatDay(iso) {
    try {
        const d = new Date(iso);
        return d.toLocaleDateString('ru-RU', { day: '2-digit', month: 'long' });
    } catch { return ''; }
}

export default function TelegramChatModal({ lead, onClose, showToast, role, onManagerMessage }) {
    const [messages, setMessages] = useState([]);
    const [chatStatus, setChatStatus] = useState(null);
    const [draft, setDraft] = useState('');
    const [sending, setSending] = useState(false);
    const [loading, setLoading] = useState(true);
    const scrollRef = useRef(null);
    const canSend = role !== 'guest';

    const fetchAll = useCallback(async () => {
        try {
            const [msgs, statusRes] = await Promise.all([
                api.getTelegramMessages(lead.id),
                api.getTelegramStatus(lead.id),
            ]);
            setMessages(msgs);
            setChatStatus(statusRes);
            setLoading(false);
        } catch (err) {
            console.error(err);
            setLoading(false);
        }
    }, [lead.id]);

    useEffect(() => {
        fetchAll();
        const interval = setInterval(fetchAll, 5000);
        return () => clearInterval(interval);
    }, [fetchAll]);

    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [messages]);

    const handleSend = async () => {
        const text = draft.trim();
        if (!text) return;
        if (!canSend) return;
        setSending(true);
        try {
            await api.sendTelegramMessage(lead.id, text);
            setDraft('');
            await fetchAll();
            if (onManagerMessage) onManagerMessage();
        } catch (err) {
            console.error(err);
            showToast && showToast('Ошибка отправки: ' + err.message, 'error');
        } finally {
            setSending(false);
        }
    };

    const handleTakeover = async () => {
        try {
            await api.takeoverTelegram(lead.id);
            showToast && showToast('Бот отключён, чат под управлением менеджера', 'success');
            await fetchAll();
        } catch (err) {
            console.error(err);
            showToast && showToast('Не удалось перехватить чат: ' + err.message, 'error');
        }
    };

    const handleKey = (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSend();
        }
    };

    const hasChat = chatStatus && chatStatus.has_chat && chatStatus.chat;
    const chat = hasChat ? chatStatus.chat : null;
    const stateLabel = chat ? (STATE_LABELS[chat.bot_state] || chat.bot_state) : '';
    const managerActive = chat && (chat.bot_state === 'manager' || !chat.bot_active);

    // Группируем сообщения по дням
    const grouped = [];
    let lastDay = '';
    messages.forEach(m => {
        const day = formatDay(m.created_at);
        if (day !== lastDay) {
            grouped.push({ type: 'day', day, key: 'd_' + m.id });
            lastDay = day;
        }
        grouped.push({ type: 'msg', msg: m, key: m.id });
    });

    return (
        <div className="modal-overlay open" onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}>
            <div className="modal" style={{ maxWidth: 640, width: '95vw', height: '85vh', display: 'flex', flexDirection: 'column' }}>
                <div className="modal-header" style={{ paddingRight: 16 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                        <div style={{ width: 40, height: 40, borderRadius: 99, background: 'linear-gradient(135deg, #0088cc, #229ED9)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                            <Bot size={20} style={{ color: '#fff' }} />
                        </div>
                        <div>
                            <h2 style={{ margin: 0, fontSize: 16 }}>{lead.name}</h2>
                            <div style={{ fontSize: 11, color: 'var(--text-muted)', display: 'flex', alignItems: 'center', gap: 6 }}>
                                <span style={{
                                    width: 7, height: 7, borderRadius: 99,
                                    background: managerActive ? '#10b981' : '#06b6d4',
                                }}></span>
                                {hasChat ? stateLabel : 'Чат ещё не открыт'}
                            </div>
                        </div>
                    </div>
                    <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                        {hasChat && !managerActive && canSend && (
                            <button className="btn btn-outline" style={{ fontSize: 12 }} onClick={handleTakeover}>
                                <UserCheck size={13} /> Перехватить
                            </button>
                        )}
                        <button className="modal-close" onClick={onClose}><X size={14} /></button>
                    </div>
                </div>

                {!hasChat && !loading && (
                    <div style={{ padding: 32, textAlign: 'center', color: 'var(--text-muted)', flex: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center', gap: 8 }}>
                        <Bot size={48} style={{ color: 'var(--text-muted)', opacity: 0.5 }} />
                        <div style={{ fontWeight: 700, color: 'var(--text-primary)' }}>Чат Telegram ещё не открыт</div>
                        <div style={{ fontSize: 13 }}>
                            Лид появится здесь, когда он напишет боту первое сообщение.
                        </div>
                        <div style={{ marginTop: 8, fontSize: 12 }}>
                            Поделитесь ссылкой на бота: <a
                                href="https://t.me/mdis_communication_bot"
                                target="_blank" rel="noopener noreferrer"
                                style={{ color: '#06b6d4', fontWeight: 600 }}
                            >https://t.me/mdis_communication_bot</a>
                        </div>
                    </div>
                )}

                {(hasChat || loading) && (
                    <div
                        ref={scrollRef}
                        style={{
                            flex: 1, overflowY: 'auto',
                            padding: '16px 20px',
                            background: 'linear-gradient(180deg, rgba(99,102,241,0.04), transparent)',
                            display: 'flex', flexDirection: 'column', gap: 10,
                        }}
                    >
                        {loading && (
                            <div style={{ textAlign: 'center', color: 'var(--text-muted)', padding: 20 }}>
                                <Loader2 size={20} className="spin" /> Загрузка...
                            </div>
                        )}
                        {!loading && messages.length === 0 && (
                            <div style={{ textAlign: 'center', color: 'var(--text-muted)', padding: 20, fontSize: 13 }}>
                                Сообщений пока нет
                            </div>
                        )}
                        {grouped.map(item => {
                            if (item.type === 'day') {
                                return (
                                    <div key={item.key} style={{ textAlign: 'center', fontSize: 11, color: 'var(--text-muted)', margin: '8px 0' }}>
                                        <span style={{ background: 'var(--glass)', padding: '4px 12px', borderRadius: 99, border: '1px solid var(--glass-border)' }}>
                                            {item.day}
                                        </span>
                                    </div>
                                );
                            }
                            const m = item.msg;
                            const isOut = m.direction === 'outbound';
                            return (
                                <div
                                    key={item.key}
                                    style={{
                                        alignSelf: isOut ? 'flex-end' : 'flex-start',
                                        maxWidth: '78%',
                                        display: 'flex',
                                        flexDirection: 'column',
                                        alignItems: isOut ? 'flex-end' : 'flex-start',
                                        gap: 2,
                                    }}
                                >
                                    <div
                                        style={{
                                            padding: '8px 12px',
                                            borderRadius: 14,
                                            borderBottomRightRadius: isOut ? 4 : 14,
                                            borderBottomLeftRadius: isOut ? 14 : 4,
                                            background: isOut
                                                ? 'linear-gradient(135deg, #6366f1, #8b5cf6)'
                                                : 'var(--glass)',
                                            color: isOut ? '#fff' : 'var(--text-primary)',
                                            border: isOut ? 'none' : '1px solid var(--glass-border)',
                                            fontSize: 13,
                                            lineHeight: 1.4,
                                            whiteSpace: 'pre-wrap',
                                            wordBreak: 'break-word',
                                        }}
                                    >
                                        {m.content}
                                    </div>
                                    <div style={{ fontSize: 10, color: 'var(--text-muted)', display: 'flex', alignItems: 'center', gap: 4 }}>
                                        {isOut ? (m.created_by ? <UserCheck size={10} /> : <Bot size={10} />) : <User size={10} />}
                                        {isOut ? (m.created_by ? 'Менеджер' : 'Бот') : 'Студент'}
                                        · {formatTime(m.created_at)}
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                )}

                {hasChat && (
                    <div style={{ padding: 12, borderTop: '1px solid var(--glass-border)', background: 'var(--bg-card)' }}>
                        {!managerActive && (
                            <div style={{ fontSize: 11, color: '#f59e0b', marginBottom: 6, display: 'flex', alignItems: 'center', gap: 4 }}>
                                <Bot size={11} /> Бот ещё ведёт диалог. Отправка вашего сообщения перехватит чат.
                            </div>
                        )}
                        <div style={{ display: 'flex', gap: 8 }}>
                            <textarea
                                value={draft}
                                onChange={e => setDraft(e.target.value)}
                                onKeyDown={handleKey}
                                placeholder={canSend ? 'Сообщение в Telegram... (Enter — отправить, Shift+Enter — новая строка)' : 'Только для чтения'}
                                rows={2}
                                disabled={!canSend || sending}
                                style={{
                                    flex: 1, padding: '10px 12px', borderRadius: 10,
                                    border: '1px solid var(--glass-border)', background: 'var(--glass)',
                                    color: 'var(--text-primary)', fontSize: 13, fontFamily: 'inherit',
                                    resize: 'none', outline: 'none',
                                }}
                            />
                            <button
                                className="btn btn-primary"
                                onClick={handleSend}
                                disabled={!draft.trim() || sending || !canSend}
                                style={{ padding: '0 16px', alignSelf: 'stretch' }}
                            >
                                {sending ? <Loader2 size={14} className="spin" /> : <Send size={14} />}
                            </button>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
