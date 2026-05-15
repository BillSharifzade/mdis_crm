import { useEffect, useState, useCallback } from 'react';
import { Activity, Phone, MessageCircle, FileText, UserPlus, TrendingUp } from 'lucide-react';
import { api } from '../services/api';
import { KpiBar } from './Settings';

export default function MyKpi({ user, showToast }) {
    const [period, setPeriod] = useState(30);
    const [stats, setStats] = useState(null);
    const [loading, setLoading] = useState(true);
    const [periodInitialized, setPeriodInitialized] = useState(false);

    // При первом монтировании пытаемся подхватить period_days из сохранённых целей.
    useEffect(() => {
        if (periodInitialized) return;
        let cancelled = false;
        (async () => {
            try {
                const targets = await api.getMyKpiTargets();
                if (cancelled) return;
                if (Array.isArray(targets) && targets.length > 0) {
                    const minPeriod = targets.reduce(
                        (m, t) => (t.period_days < m ? t.period_days : m),
                        targets[0].period_days || 30,
                    );
                    setPeriod(minPeriod);
                }
            } catch { /* ignore */ }
            finally { if (!cancelled) setPeriodInitialized(true); }
        })();
        return () => { cancelled = true; };
    }, [periodInitialized]);

    const load = useCallback(async () => {
        if (!periodInitialized) return;
        try {
            setLoading(true);
            const s = await api.getMyKpi(period);
            setStats(s);
        } catch (err) {
            showToast && showToast('Ошибка загрузки KPI: ' + err.message, 'error');
        } finally {
            setLoading(false);
        }
    }, [period, periodInitialized, showToast]);

    useEffect(() => { load(); }, [load]);

    const hasGoals = stats && (stats.processed_goal > 0 || stats.created_goal > 0);

    const tiles = stats ? [
        { label: 'Обработано лидов', value: stats.processed, Icon: Activity, color: '#6366f1' },
        { label: 'Создано лидов', value: stats.created, Icon: UserPlus, color: '#10b981' },
        { label: 'Звонков', value: stats.calls_count, Icon: Phone, color: '#a78bfa' },
        { label: 'Заметок', value: stats.notes_count, Icon: FileText, color: '#f59e0b' },
        { label: 'Сообщений', value: stats.messages_count, Icon: MessageCircle, color: '#06b6d4' },
    ] : [];

    return (
        <section className="page active">
            <div className="page-header">
                <div>
                    <h1>Мой KPI</h1>
                    <p className="page-sub">Личная активность{user?.name ? ` · ${user.name}` : ''}</p>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                    <label style={{ fontSize: 12, color: 'var(--text-secondary)' }}>Период:</label>
                    <select
                        value={period}
                        onChange={e => setPeriod(Number(e.target.value))}
                        className="filter-select"
                        style={{ width: 140 }}
                    >
                        <option value={1}>1 день</option>
                        <option value={7}>7 дней</option>
                        <option value={14}>14 дней</option>
                        <option value={30}>30 дней</option>
                        <option value={60}>60 дней</option>
                        <option value={90}>90 дней</option>
                    </select>
                </div>
            </div>

            {loading && <p style={{ color: 'var(--text-muted)' }}>Загрузка…</p>}

            {!loading && stats && (
                <>
                    <div className="kpi-grid" style={{ marginBottom: 20 }}>
                        {tiles.map(t => (
                            <div className="kpi-card" key={t.label}>
                                <div className="kpi-icon" style={{ background: `${t.color}20`, color: t.color }}>
                                    <t.Icon size={20} />
                                </div>
                                <div className="kpi-body">
                                    <span className="kpi-label">{t.label}</span>
                                    <span className="kpi-value">{t.value}</span>
                                </div>
                                <div className="kpi-glow"></div>
                            </div>
                        ))}
                    </div>

                    <div className="card" style={{ padding: 22 }}>
                        <h3 style={{ fontSize: 14, fontWeight: 700, color: 'var(--text-primary)', marginBottom: 14, display: 'flex', alignItems: 'center', gap: 6 }}>
                            <TrendingUp size={16} /> Цели и прогресс
                        </h3>
                        {!hasGoals && (
                            <p style={{ fontSize: 13, color: 'var(--text-muted)' }}>
                                Цели KPI ещё не назначены. Попросите администратора задать цели
                                в разделе «Настройки → Пользователи → KPI».
                            </p>
                        )}
                        {stats.processed_goal > 0 && (
                            <div style={{ marginBottom: 16 }}>
                                <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--text-primary)', marginBottom: 4 }}>
                                    Обработка лидов
                                </div>
                                <KpiBar val={stats.processed} goal={stats.processed_goal} color="#6366f1" />
                            </div>
                        )}
                        {stats.created_goal > 0 && (
                            <div>
                                <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--text-primary)', marginBottom: 4 }}>
                                    Создание лидов
                                </div>
                                <KpiBar val={stats.created} goal={stats.created_goal} color="#10b981" />
                            </div>
                        )}
                    </div>
                </>
            )}
        </section>
    );
}
