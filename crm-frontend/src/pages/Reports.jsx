import { useEffect, useMemo, useState } from 'react';
import { BarChart2, TrendingUp, Users, Clock, XCircle, Radio, Download, FileText, Calendar, FileSpreadsheet } from 'lucide-react';
import { api } from '../services/api';
import { getStatusObj, formatDate } from '../data/crmData';
import { useStages } from '../context/useStages';

const iconMap = {
    'bar-chart-2': BarChart2,
    'trending-up': TrendingUp,
    'users': Users,
    'clock': Clock,
    'x-circle': XCircle,
    'radio': Radio,
};

const reports = [
    { icon: 'bar-chart-2', bg: 'rgba(99,102,241,0.15)', col: '#818cf8', title: 'Отчёт по заявкам', desc: 'Сводная таблица лидов с источниками и статусами за выбранный период.' },
    { icon: 'trending-up', bg: 'rgba(16,185,129,0.15)', col: '#10b981', title: 'Конверсия воронки', desc: 'Выбытие и конверсия по каждому этапу, сравнение с предыдущим периодом.' },
    { icon: 'users', bg: 'rgba(236,72,153,0.15)', col: '#ec4899', title: 'KPI менеджеров', desc: 'Звонки, закрытые сделки и привлечённая выручка по каждому сотруднику.' },
    { icon: 'clock', bg: 'rgba(245,158,11,0.15)', col: '#f59e0b', title: 'Скорость обработки', desc: 'Среднее время реакции и перехода между этапами воронки.' },
    { icon: 'x-circle', bg: 'rgba(239,68,68,0.15)', col: '#ef4444', title: 'Анализ отказов', desc: 'Причины потери лидов, ABC-анализ факторов отказа.' },
    { icon: 'radio', bg: 'rgba(6,182,212,0.15)', col: '#06b6d4', title: 'Источники трафика', desc: 'Эффективность каналов привлечения — UTM, органика, реклама, мессенджеры.' },
];

function todayISO() { return new Date().toISOString().slice(0, 10); }
function daysAgoISO(d) {
    const dt = new Date(); dt.setDate(dt.getDate() - d);
    return dt.toISOString().slice(0, 10);
}

export default function Reports({ showToast }) {
    const { stages: stagesAll } = useStages();
    const stages = stagesAll && stagesAll.length > 0 ? stagesAll : [];

    const [from, setFrom] = useState(daysAgoISO(30));
    const [to, setTo] = useState(todayISO());
    const [busy, setBusy] = useState(false);
    const [leads, setLeads] = useState([]);
    const [loadingLeads, setLoadingLeads] = useState(false);
    const [managersKpi, setManagersKpi] = useState([]);

    useEffect(() => {
        if (!api.useApi) return;
        setLoadingLeads(true);
        api.getLeads(500, 0)
            .then(setLeads)
            .catch(err => console.error(err))
            .finally(() => setLoadingLeads(false));
        api.getManagerKpi().then(setManagersKpi).catch(() => setManagersKpi([]));
    }, []);

    const inRange = useMemo(() => {
        const fromD = new Date(from + 'T00:00:00');
        const toD = new Date(to + 'T23:59:59');
        return leads.filter(l => {
            const d = l.date instanceof Date ? l.date : new Date(l.date);
            return d >= fromD && d <= toD;
        });
    }, [leads, from, to]);

    const byStatus = useMemo(() => {
        const map = Object.fromEntries(stages.map(s => [s.key, 0]));
        inRange.forEach(l => { if (map[l.status] !== undefined) map[l.status]++; });
        return map;
    }, [inRange, stages]);

    const bySource = useMemo(() => {
        const map = {};
        inRange.forEach(l => { map[l.source] = (map[l.source] || 0) + 1; });
        return map;
    }, [inRange]);

    const totalEnrolled = byStatus.enrolled || 0;
    const totalLost = byStatus.lost || 0;
    const conversion = inRange.length ? ((totalEnrolled / inRange.length) * 100).toFixed(1) : '0.0';

    const download = async (format, reportTitle) => {
        if (!api.useApi) {
            showToast('Экспорт работает только при подключённом API', 'warning');
            return;
        }
        setBusy(true);
        try {
            await api.exportAnalytics(format, { from, to });
            showToast(`Скачивается «${reportTitle || 'отчёт'}» (${format.toUpperCase()})`, 'success');
        } catch (err) {
            console.error(err);
            showToast('Ошибка экспорта: ' + err.message, 'error');
        } finally {
            setBusy(false);
        }
    };

    const presets = [
        { label: '7 дней', days: 7 },
        { label: '30 дней', days: 30 },
        { label: '90 дней', days: 90 },
        { label: 'Год', days: 365 },
    ];

    const topSources = Object.entries(bySource).sort((a, b) => b[1] - a[1]).slice(0, 5);
    const recentLeads = useMemo(() => {
        return [...inRange].sort((a, b) => {
            const da = a.date instanceof Date ? a.date : new Date(a.date);
            const db = b.date instanceof Date ? b.date : new Date(b.date);
            return db - da;
        }).slice(0, 15);
    }, [inRange]);

    return (
        <section className="page active">
            <div className="page-header">
                <div>
                    <h1>Отчёты</h1>
                    <p className="page-sub">Экспортируемые документы и сводки</p>
                </div>
                <div className="header-actions">
                    <button className="btn btn-outline" disabled={busy} onClick={() => download('pdf', 'Отчёт PDF')}><FileText size={14} /> Скачать PDF</button>
                    <button className="btn btn-primary" disabled={busy} onClick={() => download('xlsx', 'Отчёт Excel')}><Download size={14} /> Скачать Excel</button>
                </div>
            </div>

            <div className="card" style={{ padding: 16, marginBottom: 20, display: 'flex', alignItems: 'center', gap: 14, flexWrap: 'wrap' }}>
                <Calendar size={16} style={{ color: 'var(--text-secondary)' }} />
                <label style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13, color: 'var(--text-secondary)' }}>
                    с
                    <input type="date" value={from} max={to} onChange={e => setFrom(e.target.value)}
                        style={{ padding: '6px 10px', borderRadius: 6, border: '1px solid var(--glass-border)', background: 'var(--glass)', color: 'var(--text-primary)', fontFamily: 'inherit', fontSize: 13 }} />
                </label>
                <label style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13, color: 'var(--text-secondary)' }}>
                    по
                    <input type="date" value={to} min={from} max={todayISO()} onChange={e => setTo(e.target.value)}
                        style={{ padding: '6px 10px', borderRadius: 6, border: '1px solid var(--glass-border)', background: 'var(--glass)', color: 'var(--text-primary)', fontFamily: 'inherit', fontSize: 13 }} />
                </label>
                <div style={{ display: 'flex', gap: 6, marginLeft: 'auto', flexWrap: 'wrap' }}>
                    {presets.map(p => (
                        <button key={p.days} className="btn btn-outline" style={{ padding: '6px 12px', fontSize: 12 }}
                            onClick={() => { setFrom(daysAgoISO(p.days)); setTo(todayISO()); }}>
                            {p.label}
                        </button>
                    ))}
                </div>
            </div>

            {/* Summary KPIs */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: 14, marginBottom: 20 }}>
                <KpiCard label="Всего лидов за период" value={inRange.length} color="#818cf8" loading={loadingLeads} />
                <KpiCard label="Зачислено" value={totalEnrolled} color="#10b981" loading={loadingLeads} />
                <KpiCard label="Проиграно" value={totalLost} color="#ef4444" loading={loadingLeads} />
                <KpiCard label="Конверсия в зачисление" value={`${conversion}%`} color="#a855f7" loading={loadingLeads} />
                <KpiCard label="Активно в работе" value={inRange.length - totalEnrolled - totalLost} color="#06b6d4" loading={loadingLeads} />
            </div>

            {/* Status & sources breakdown */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginBottom: 20 }}>
                <div className="card" style={{ padding: 20 }}>
                    <h3 style={{ fontSize: 14, fontWeight: 700, marginBottom: 14, color: 'var(--text-primary)' }}>Распределение по этапам</h3>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                        {stages.map(s => {
                            const count = byStatus[s.key] || 0;
                            const pct = inRange.length ? Math.round((count / inRange.length) * 100) : 0;
                            return (
                                <div key={s.key}>
                                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, marginBottom: 4 }}>
                                        <span style={{ color: 'var(--text-secondary)' }}>{s.label}</span>
                                        <span style={{ color: 'var(--text-primary)', fontWeight: 600 }}>{count} <span style={{ color: 'var(--text-muted)', fontWeight: 400 }}>({pct}%)</span></span>
                                    </div>
                                    <div style={{ height: 6, background: 'var(--glass)', borderRadius: 99, overflow: 'hidden' }}>
                                        <div style={{ height: '100%', width: `${pct}%`, background: s.color, transition: 'width .4s ease' }}></div>
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                </div>

                <div className="card" style={{ padding: 20 }}>
                    <h3 style={{ fontSize: 14, fontWeight: 700, marginBottom: 14, color: 'var(--text-primary)' }}>Топ источников трафика</h3>
                    {topSources.length === 0 && (
                        <p style={{ color: 'var(--text-muted)', fontSize: 13 }}>Нет данных за выбранный период</p>
                    )}
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                        {topSources.map(([src, count], i) => {
                            const pct = inRange.length ? Math.round((count / inRange.length) * 100) : 0;
                            const colors = ['#818cf8', '#10b981', '#f59e0b', '#ec4899', '#06b6d4'];
                            const c = colors[i % colors.length];
                            return (
                                <div key={src} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                                    <span style={{ width: 10, height: 10, borderRadius: 99, background: c }}></span>
                                    <span style={{ flex: 1, fontSize: 13, color: 'var(--text-primary)' }}>{src}</span>
                                    <span style={{ fontSize: 12, color: 'var(--text-muted)' }}>{pct}%</span>
                                    <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--text-primary)', minWidth: 30, textAlign: 'right' }}>{count}</span>
                                </div>
                            );
                        })}
                    </div>
                </div>
            </div>

            {/* KPI Менеджеров приёма */}
            <div className="card" style={{ padding: 0, marginBottom: 20 }}>
                <div style={{ padding: '14px 18px', borderBottom: '1px solid var(--glass-border)' }}>
                    <h3 style={{ fontSize: 14, fontWeight: 700, color: 'var(--text-primary)' }}>Эффективность менеджеров приёма</h3>
                </div>
                {managersKpi.length === 0 ? (
                    <p style={{ padding: 18, color: 'var(--text-muted)', fontSize: 13 }}>Нет данных по менеджерам</p>
                ) : (
                    <table className="managers-table">
                        <thead>
                            <tr><th>Менеджер</th><th>Лидов</th><th>Звонков</th><th>Зачислений</th><th>Конверсия</th></tr>
                        </thead>
                        <tbody>
                            {managersKpi.map(m => {
                                const conv = m.leads_count > 0 ? Math.round((m.closed_deals / m.leads_count) * 100) : 0;
                                return (
                                    <tr key={m.manager_id}>
                                        <td><span className="manager-name">{m.manager_name}</span></td>
                                        <td>{m.leads_count}</td>
                                        <td>{m.calls_count}</td>
                                        <td>{m.closed_deals}</td>
                                        <td><span style={{ color: conv > 27 ? '#10b981' : conv > 22 ? '#f59e0b' : 'var(--text-secondary)', fontWeight: 700 }}>{conv}%</span></td>
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
                )}
            </div>

            {/* Recent leads */}
            <div className="card" style={{ padding: 0, marginBottom: 20 }}>
                <div style={{ padding: '14px 18px', borderBottom: '1px solid var(--glass-border)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                    <h3 style={{ fontSize: 14, fontWeight: 700, color: 'var(--text-primary)' }}>Последние лиды за период</h3>
                    <button className="btn btn-outline" onClick={() => download('xlsx', 'Лиды за период')} disabled={busy}>
                        <FileSpreadsheet size={13} /> Экспорт CSV/XLSX
                    </button>
                </div>
                <table className="leads-table">
                    <thead>
                        <tr>
                            <th>Лид</th>
                            <th>Программа</th>
                            <th>Источник</th>
                            <th>Статус</th>
                            <th>Дата</th>
                        </tr>
                    </thead>
                    <tbody>
                        {recentLeads.length === 0 && (
                            <tr><td colSpan="5" style={{ textAlign: 'center', padding: 30, color: 'var(--text-muted)' }}>
                                {loadingLeads ? 'Загрузка…' : 'Нет лидов в выбранном периоде'}
                            </td></tr>
                        )}
                        {recentLeads.map(l => {
                            const st = getStatusObj(l.status);
                            return (
                                <tr key={l.id}>
                                    <td>
                                        <div className="lead-cell">
                                            <div className="lead-avatar" style={{ background: l.color, color: '#fff', fontSize: 11, fontWeight: 700 }}>{l.initials}</div>
                                            <div>
                                                <div className="lead-cell-name">{l.name}</div>
                                                <div className="lead-cell-phone">{l.phone || '—'}</div>
                                            </div>
                                        </div>
                                    </td>
                                    <td style={{ color: 'var(--text-secondary)' }}>{l.program}</td>
                                    <td><span className="source-chip">{l.source}</span></td>
                                    <td><span className={`status-badge ${st.cls}`}>{st.label}</span></td>
                                    <td style={{ color: 'var(--text-muted)' }}>{formatDate(l.date)}</td>
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
            </div>

            <h2 style={{ fontSize: 16, fontWeight: 700, margin: '0 0 14px', color: 'var(--text-primary)' }}>Готовые отчёты</h2>
            <div className="reports-grid">
                {reports.map((r, i) => {
                    const Icon = iconMap[r.icon] || BarChart2;
                    return (
                        <div className="report-card" key={i}>
                            <div className="report-card-icon" style={{ background: r.bg }}>
                                <Icon size={20} style={{ color: r.col }} />
                            </div>
                            <h3>{r.title}</h3>
                            <p>{r.desc}</p>
                            <div className="report-meta">
                                <span className="report-updated">за {from} — {to}</span>
                                <div style={{ display: 'flex', gap: 6 }}>
                                    <button className="report-dl-btn" disabled={busy} onClick={() => download('pdf', r.title)}>
                                        <FileText size={13} /> PDF
                                    </button>
                                    <button className="report-dl-btn" disabled={busy} onClick={() => download('xlsx', r.title)}>
                                        <Download size={13} /> Excel
                                    </button>
                                </div>
                            </div>
                        </div>
                    );
                })}
            </div>
        </section>
    );
}

function KpiCard({ label, value, color, loading }) {
    return (
        <div className="card" style={{ padding: 16 }}>
            <div style={{ fontSize: 11, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.04em', marginBottom: 8 }}>{label}</div>
            <div style={{ fontSize: 26, fontWeight: 800, color, lineHeight: 1 }}>
                {loading ? '…' : value}
            </div>
        </div>
    );
}
