import { useEffect, useRef, useMemo, useState } from 'react';
import { UserPlus, CheckCircle, Percent, Clock, TrendingUp, TrendingDown, Calendar, Download } from 'lucide-react';
import { Chart, registerables } from 'chart.js';
import { getStatusObj, formatDate } from '../data/crmData';
import { useStages } from '../context/useStages';
import { api } from '../services/api';

Chart.register(...registerables);

function todayISO() { return new Date().toISOString().slice(0, 10); }
function daysAgoISO(d) { const dt = new Date(); dt.setDate(dt.getDate() - d); return dt.toISOString().slice(0, 10); }

export default function Dashboard({ allLeads, navigate, openDetail, user }) {
    const greetingName = (user?.name || '').split(/\s+/)[0] || 'коллега';
    const { stages: stagesAll } = useStages();
    const stages = (stagesAll && stagesAll.length > 0 ? stagesAll : []).filter(s => s.key !== 'lost');

    const [stats, setStats] = useState({
        newLeads: 0, inSchool: 0, conversion: 0,
        avgLeadDays: 0, totalLost: 0,
    });
    const [from, setFrom] = useState(daysAgoISO(56));
    const [to, setTo] = useState(todayISO());
    const [series, setSeries] = useState([]);
    const [sourceBreakdown, setSourceBreakdown] = useState([]);

    const leadsChartRef = useRef(null);
    const sourceChartRef = useRef(null);
    const leadsChartInstance = useRef(null);
    const sourceChartInstance = useRef(null);

    // Загрузка реальных метрик
    useEffect(() => {
        Promise.all([
            api.getDashboardStats().catch(() => null),
            api.getConversionStats().catch(() => null),
            api.getSourceBreakdown().catch(() => []),
        ]).then(([dashData, convData, srcData]) => {
            setStats({
                newLeads: dashData?.total_new_leads ?? 0,
                inSchool: dashData?.total_in_school ?? 0,
                conversion: convData?.conversion_rate ?? 0,
                avgLeadDays: dashData?.average_lead_to_enrolled_days ?? 0,
                totalLost: dashData?.total_lost ?? 0,
            });
            setSourceBreakdown(srcData);
        });
    }, []);

    // Загрузка временного ряда (по неделям)
    useEffect(() => {
        const days = Math.max(7, Math.ceil((new Date(to) - new Date(from)) / 86400000) + 1);
        api.getTimeSeries(days).then(setSeries).catch(() => setSeries([]));
    }, [from, to]);

    // Leads chart
    useEffect(() => {
        if (!leadsChartRef.current) return;
        if (leadsChartInstance.current) leadsChartInstance.current.destroy();
        const labels = series.map(s => {
            const d = new Date(s.bucket);
            return d.toLocaleDateString('ru-RU', { month: 'short', day: '2-digit' });
        });
        leadsChartInstance.current = new Chart(leadsChartRef.current.getContext('2d'), {
            type: 'line',
            data: {
                labels: labels.length > 0 ? labels : ['—'],
                datasets: [
                    {
                        label: 'Заявки', data: series.map(s => s.leads),
                        borderColor: '#6366f1', backgroundColor: 'rgba(99,102,241,0.1)',
                        fill: true, tension: 0.4, pointRadius: 4, pointBackgroundColor: '#6366f1', borderWidth: 2.5,
                    },
                    {
                        label: 'Зачисления', data: series.map(s => s.enrolled),
                        borderColor: '#10b981', backgroundColor: 'rgba(16,185,129,0.08)',
                        fill: true, tension: 0.4, pointRadius: 4, pointBackgroundColor: '#10b981', borderWidth: 2,
                    }
                ]
            },
            options: {
                responsive: true, maintainAspectRatio: true,
                plugins: { legend: { display: false }, tooltip: tooltipDefaults() },
                scales: { x: axisDefaults(), y: { ...axisDefaults(), beginAtZero: true } }
            }
        });
        return () => { if (leadsChartInstance.current) leadsChartInstance.current.destroy(); };
    }, [series]);

    // Source chart
    useEffect(() => {
        if (!sourceChartRef.current) return;
        if (sourceChartInstance.current) sourceChartInstance.current.destroy();
        const palette = ['#6366f1', '#06b6d4', '#ec4899', '#10b981', '#f59e0b', '#a855f7', '#64748b', '#ef4444'];
        const data = sourceBreakdown.length > 0
            ? sourceBreakdown.map((s, i) => ({ label: s.source, value: s.count, color: palette[i % palette.length] }))
            : [{ label: 'Нет данных', value: 1, color: '#64748b' }];
        sourceChartInstance.current = new Chart(sourceChartRef.current.getContext('2d'), {
            type: 'doughnut',
            data: {
                labels: data.map(d => d.label),
                datasets: [{
                    data: data.map(d => d.value),
                    backgroundColor: data.map(d => d.color),
                    borderColor: 'transparent', borderWidth: 3, hoverOffset: 8,
                }]
            },
            options: {
                cutout: '68%',
                plugins: { legend: { display: false }, tooltip: tooltipDefaults() },
                animation: { animateRotate: true, duration: 1000 }
            }
        });
        return () => { if (sourceChartInstance.current) sourceChartInstance.current.destroy(); };
    }, [sourceBreakdown]);

    const recentLeads = useMemo(() => allLeads.slice(0, 6), [allLeads]);

    const funnelData = useMemo(() => {
        const totalNonLost = allLeads.filter(l => l.status !== 'lost').length;
        return stages.map(st => {
            const count = allLeads.filter(l => l.status === st.key).length;
            const pct = totalNonLost ? Math.round(count / totalNonLost * 100) : 0;
            return { ...st, count, pct };
        });
    }, [allLeads, stages]);

    const kpiData = [
        { label: 'Новых заявок', value: stats.newLeads, Icon: UserPlus, color: 'kpi-purple' },
        { label: 'Успешно зачислено', value: stats.inSchool, Icon: CheckCircle, color: 'kpi-green' },
        { label: 'Конверсия', value: `${(stats.conversion || 0).toFixed(1)}%`, Icon: Percent, color: 'kpi-orange' },
        {
            label: 'Ср. время до зачисления',
            value: stats.avgLeadDays > 0 ? `${stats.avgLeadDays.toFixed(1)} дн.` : '—',
            Icon: Clock, color: 'kpi-pink',
        },
    ];

    const presets = [
        { label: '7 дн', days: 7 },
        { label: '30 дн', days: 30 },
        { label: '90 дн', days: 90 },
    ];

    const palette = ['#6366f1', '#06b6d4', '#ec4899', '#10b981', '#f59e0b', '#a855f7', '#64748b', '#ef4444'];

    return (
        <section className="page active">
            <div className="page-header">
                <div>
                    <h1>Добро пожаловать, {greetingName} 👋</h1>
                    <p className="page-sub">Вот что происходит с вашими заявками сегодня</p>
                </div>
                <div style={{ display: 'flex', gap: '10px', alignItems: 'center', flexWrap: 'wrap' }}>
                    <button
                        onClick={() => api.exportAnalytics('xlsx', { from, to })}
                        className="btn btn-outline"
                    >
                        <Download size={14} /> Скачать Excel
                    </button>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '6px 10px', borderRadius: 8, background: 'var(--glass)', border: '1px solid var(--glass-border)' }}>
                        <Calendar size={14} style={{ color: 'var(--text-muted)' }} />
                        <input
                            type="date" value={from} max={to}
                            onChange={e => setFrom(e.target.value)}
                            className="theme-aware-date"
                            style={{ background: 'transparent', border: 0, color: 'var(--text-primary)', fontSize: 12, fontFamily: 'inherit', width: 120 }}
                        />
                        <span style={{ color: 'var(--text-muted)' }}>—</span>
                        <input
                            type="date" value={to} min={from} max={todayISO()}
                            onChange={e => setTo(e.target.value)}
                            className="theme-aware-date"
                            style={{ background: 'transparent', border: 0, color: 'var(--text-primary)', fontSize: 12, fontFamily: 'inherit', width: 120 }}
                        />
                    </div>
                    {presets.map(p => (
                        <button key={p.days} className="btn btn-outline" style={{ padding: '6px 10px', fontSize: 12 }}
                            onClick={() => { setFrom(daysAgoISO(p.days)); setTo(todayISO()); }}>
                            {p.label}
                        </button>
                    ))}
                </div>
            </div>

            <div className="kpi-grid">
                {kpiData.map((kpi, i) => (
                    <div className={`kpi-card ${kpi.color}`} key={i}>
                        <div className="kpi-icon"><kpi.Icon size={20} /></div>
                        <div className="kpi-body">
                            <span className="kpi-label">{kpi.label}</span>
                            <span className="kpi-value">{kpi.value}</span>
                        </div>
                        <div className="kpi-glow"></div>
                    </div>
                ))}
            </div>

            <div className="charts-row">
                <div className="chart-card wide">
                    <div className="chart-header">
                        <div>
                            <h3>Динамика заявок</h3>
                            <p>Поступление новых лидов по неделям</p>
                        </div>
                        <div className="chart-legend">
                            <span className="legend-dot" style={{ background: '#6366f1' }}></span>Заявки
                            <span className="legend-dot" style={{ background: '#10b981' }}></span>Зачисления
                        </div>
                    </div>
                    <canvas ref={leadsChartRef} height="100"></canvas>
                </div>
                <div className="chart-card">
                    <div className="chart-header">
                        <div>
                            <h3>Источники</h3>
                            <p>Каналы привлечения</p>
                        </div>
                    </div>
                    <canvas ref={sourceChartRef} height="200"></canvas>
                    <div className="source-legend">
                        {sourceBreakdown.length === 0 && <p style={{ color: 'var(--text-muted)', fontSize: 12 }}>Нет данных</p>}
                        {sourceBreakdown.map((d, i) => (
                            <div className="source-item" key={d.source}>
                                <div className="source-item-left">
                                    <div className="source-dot" style={{ background: palette[i % palette.length] }}></div>{d.source}
                                </div>
                                <span className="source-pct">{d.pct}%</span>
                            </div>
                        ))}
                    </div>
                </div>
            </div>

            <div className="bottom-row">
                <div className="card recent-leads-card">
                    <div className="card-header">
                        <h3>Последние заявки</h3>
                        <button className="link-btn" onClick={() => navigate('leads')}>Все лиды →</button>
                    </div>
                    <div className="leads-list">
                        {recentLeads.map(l => {
                            const st = getStatusObj(l.status);
                            return (
                                <div className="lead-list-item" key={l.id} onClick={() => openDetail(l)}>
                                    <div className="lead-avatar" style={{ background: l.color, color: '#fff', fontSize: 11, fontWeight: 700 }}>{l.initials}</div>
                                    <div className="lead-info">
                                        <div className="lead-name">{l.name}</div>
                                        <div className="lead-meta">{l.program} · <span className={`status-badge ${st.cls}`} style={{ display: 'inline-flex' }}>{st.label}</span></div>
                                    </div>
                                    <span className="lead-time">{formatDate(l.date)}</span>
                                </div>
                            );
                        })}
                    </div>
                </div>
                <div className="card pipeline-mini-card">
                    <div className="card-header">
                        <h3>Статус воронки</h3>
                        <button className="link-btn" onClick={() => navigate('pipeline')}>Открыть →</button>
                    </div>
                    <div className="funnel-bars">
                        {funnelData.map(fd => (
                            <div className="funnel-row" key={fd.key}>
                                <div className="funnel-label-row">
                                    <span className="funnel-stage-name">{fd.label}</span>
                                    <span className="funnel-stage-count">{fd.count}</span>
                                </div>
                                <div className="funnel-track">
                                    <FunnelFill pct={fd.pct} color={fd.color} />
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </section>
    );
}

function FunnelFill({ pct, color }) {
    const ref = useRef(null);
    useEffect(() => {
        setTimeout(() => {
            if (ref.current) ref.current.style.width = pct + '%';
        }, 100);
    }, [pct]);
    return <div className="funnel-fill" ref={ref} style={{ width: 0, background: color }}></div>;
}

function tooltipDefaults() {
    return {
        backgroundColor: '#22263a', titleColor: '#f1f5f9', bodyColor: '#94a3b8',
        borderColor: 'rgba(255,255,255,0.07)', borderWidth: 1,
        padding: 12, cornerRadius: 8,
        titleFont: { family: 'Inter', weight: '700', size: 13 },
        bodyFont: { family: 'Inter', size: 12 },
    };
}
function axisDefaults() {
    return {
        grid: { color: 'rgba(255,255,255,0.05)', drawBorder: false },
        ticks: { color: '#64748b', font: { family: 'Inter', size: 11 } },
        border: { display: false },
    };
}
