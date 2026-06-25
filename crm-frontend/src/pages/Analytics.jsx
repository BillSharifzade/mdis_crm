import { useEffect, useRef, useState } from 'react';
import { Calendar } from 'lucide-react';
import { Chart, registerables } from 'chart.js';
import { MANAGERS, MANAGER_COLORS } from '../data/crmData';
import { api } from '../services/api';

Chart.register(...registerables);

function todayISO() { return new Date().toISOString().slice(0, 10); }
function daysAgoISO(d) { const dt = new Date(); dt.setDate(dt.getDate() - d); return dt.toISOString().slice(0, 10); }

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

export default function Analytics() {
    const [from, setFrom] = useState(daysAgoISO(30));
    const [to, setTo] = useState(todayISO());
    const [managers, setManagers] = useState([]);
    const [funnel, setFunnel] = useState([]);
    const [dashTotals, setDashTotals] = useState({ totalLeads: 0, totalLost: 0 });
    const [refusalReasons, setRefusalReasons] = useState([]);

    const convRef = useRef(null);
    const lostRef = useRef(null);
    const convInstance = useRef(null);
    const lostInstance = useRef(null);

    useEffect(() => {
        api.getManagerKpi().then(data => {
            if (Array.isArray(data)) {
                setManagers(data.map((m, i) => ({
                    idx: i % MANAGER_COLORS.length,
                    name: m.manager_name,
                    leads: m.leads_count || 0,
                    calls: m.calls_count || 0,
                    enrolled: m.closed_deals || 0,
                    revenue: m.revenue_added || 0,
                })));
            }
        }).catch(() => null);
        api.getFunnel().then(setFunnel).catch(() => setFunnel([]));
        api.getDashboardStats().then(d => {
            if (d) setDashTotals({ totalLeads: d.total_leads || 0, totalLost: d.total_lost || 0 });
        }).catch(() => null);
        // refusal reasons: использовать status_history endpoint когда появится; пока — заглушка
        setRefusalReasons([]);
    }, []);

    // Воронка-бар
    useEffect(() => {
        if (!convRef.current) return;
        if (convInstance.current) convInstance.current.destroy();
        const data = (funnel || []).filter(s => !/отказ|проигр/i.test(s.name));
        convInstance.current = new Chart(convRef.current.getContext('2d'), {
            type: 'bar',
            data: {
                labels: data.map(s => s.name),
                datasets: [{
                    label: 'Кол-во', data: data.map(s => s.count),
                    backgroundColor: ['#6366f1', '#06b6d4', '#f59e0b', '#a855f7', '#ec4899', '#10b981'],
                    borderRadius: 8, borderSkipped: false,
                }]
            },
            options: {
                responsive: true, maintainAspectRatio: false,
                plugins: { legend: { display: false }, tooltip: tooltipDefaults() },
                scales: { x: axisDefaults(), y: { ...axisDefaults(), beginAtZero: true } }
            }
        });
        return () => { if (convInstance.current) convInstance.current.destroy(); };
    }, [funnel]);

    // Pie — отказы (пока — общая статистика)
    useEffect(() => {
        if (!lostRef.current) return;
        if (lostInstance.current) lostInstance.current.destroy();
        const reasons = refusalReasons.length > 0
            ? refusalReasons
            : [{ label: 'Отказ (всего)', value: dashTotals.totalLost || 0, color: '#ef4444' }];
        const total = reasons.reduce((s, r) => s + r.value, 0) || 1;
        lostInstance.current = new Chart(lostRef.current.getContext('2d'), {
            type: 'pie',
            data: {
                labels: reasons.map(r => r.label),
                datasets: [{
                    data: reasons.map(r => Math.max(1, r.value)),
                    backgroundColor: reasons.map(r => r.color || '#64748b'),
                    borderColor: 'transparent', borderWidth: 3, hoverOffset: 6,
                }]
            },
            options: {
                plugins: {
                    legend: { position: 'bottom', labels: { color: '#94a3b8', font: { family: 'Inter', size: 11 }, boxWidth: 10, padding: 14 } },
                    tooltip: tooltipDefaults()
                },
                responsive: true, maintainAspectRatio: false,
                animation: { duration: 900 }
            }
        });
        return () => { if (lostInstance.current) lostInstance.current.destroy(); };
    }, [refusalReasons, dashTotals.totalLost]);

    const totalCalls = managers.reduce((s, m) => s + (m.calls || 0), 0);
    const totalEnrolled = managers.reduce((s, m) => s + (m.enrolled || 0), 0);
    const totalLeads = dashTotals.totalLeads || managers.reduce((s, m) => s + (m.leads || 0), 0);
    const totalLost = dashTotals.totalLost || 0;
    const conv = totalLeads > 0 ? Math.round((totalEnrolled / totalLeads) * 100) : 0;

    const analyticsKPIs = [
        { label: 'Всего звонков', value: totalCalls, pct: Math.min(100, totalCalls), gradient: 'linear-gradient(90deg,#8b5cf6,#a78bfa)' },
        { label: 'Зачислений', value: totalEnrolled, pct: conv, gradient: 'linear-gradient(90deg,#059669,#34d399)' },
        { label: 'Отказов', value: totalLost, pct: totalLeads > 0 ? Math.round((totalLost / totalLeads) * 100) : 0, gradient: 'linear-gradient(90deg,#dc2626,#f87171)' },
        { label: 'Конверсия', value: `${conv}%`, pct: conv, gradient: 'linear-gradient(90deg,#d97706,#fbbf24)' },
    ];

    return (
        <section className="page active">
            <div className="page-header">
                <div>
                    <h1>Аналитика</h1>
                    <p className="page-sub">KPI и эффективность команды</p>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '6px 10px', borderRadius: 8, background: 'var(--glass)', border: '1px solid var(--glass-border)' }}>
                    <Calendar size={14} style={{ color: 'var(--text-muted)' }} />
                    <input type="date" value={from} max={to} onChange={e => setFrom(e.target.value)}
                        className="theme-aware-date"
                        style={{ background: 'transparent', border: 0, color: 'var(--text-primary)', fontSize: 12, fontFamily: 'inherit', width: 120 }} />
                    <span style={{ color: 'var(--text-muted)' }}>—</span>
                    <input type="date" value={to} min={from} max={todayISO()} onChange={e => setTo(e.target.value)}
                        className="theme-aware-date"
                        style={{ background: 'transparent', border: 0, color: 'var(--text-primary)', fontSize: 12, fontFamily: 'inherit', width: 120 }} />
                </div>
            </div>

            <div className="analytics-kpi-grid">
                {analyticsKPIs.map((kpi, i) => (
                    <div className="analytics-kpi" key={i}>
                        <span className="ak-label">{kpi.label}</span>
                        <span className="ak-value">{kpi.value}</span>
                        <div className="ak-bar">
                            <AkBarFill pct={kpi.pct} gradient={kpi.gradient} />
                        </div>
                    </div>
                ))}
            </div>

            <div className="charts-row">
                <div className="chart-card">
                    <div className="chart-header">
                        <div><h3>Конверсия по этапам</h3><p>Воронка выбытия</p></div>
                    </div>
                    <div style={{ position: 'relative', height: '220px', width: '100%' }}>
                        <canvas ref={convRef}></canvas>
                    </div>
                </div>
                <div className="chart-card">
                    <div className="chart-header">
                        <div><h3>Причины отказов</h3><p>Статус — Отказ</p></div>
                    </div>
                    <div style={{ position: 'relative', height: '220px', width: '100%' }}>
                        <canvas ref={lostRef}></canvas>
                    </div>
                </div>
            </div>

            <div className="card managers-card">
                <div className="card-header">
                    <h3>Эффективность менеджеров</h3>
                    <span className="period-tag">{from} — {to}</span>
                </div>
                <table className="managers-table">
                    <thead>
                        <tr><th>Менеджер</th><th>Лидов</th><th>Звонков</th><th>Зачислений</th><th>Конверсия</th></tr>
                    </thead>
                    <tbody>
                        {managers.map((m, i) => {
                            const leadsCount = m.leads || 1;
                            const c = Math.round(m.enrolled / leadsCount * 100) || 0;
                            return (
                                <tr key={m.idx || i}>
                                    <td>
                                        <div className="manager-cell">
                                            <div style={{ width: 32, height: 32, borderRadius: 99, background: MANAGER_COLORS[m.idx] || '#6366f1', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 11, fontWeight: 700, color: '#fff' }}>
                                                {MANAGERS[m.idx] || m.name.substring(0, 2).toUpperCase()}
                                            </div>
                                            <span className="manager-name">{m.name}</span>
                                        </div>
                                    </td>
                                    <td>{m.leads}</td>
                                    <td>{m.calls}</td>
                                    <td>{m.enrolled}</td>
                                    <td><span style={{ color: c > 27 ? 'var(--success)' : c > 22 ? 'var(--warning)' : 'var(--text-secondary)', fontWeight: 700 }}>{c}%</span></td>
                                </tr>
                            );
                        })}
                        {managers.length === 0 && (
                            <tr><td colSpan="5" style={{ textAlign: 'center', padding: 20, color: 'var(--text-muted)' }}>Нет данных</td></tr>
                        )}
                    </tbody>
                </table>
            </div>
        </section>
    );
}

function AkBarFill({ pct, gradient }) {
    const ref = useRef(null);
    useEffect(() => {
        setTimeout(() => {
            if (ref.current) ref.current.style.width = Math.min(100, pct) + '%';
        }, 200);
    }, [pct]);
    return <div ref={ref} style={{ height: '100%', borderRadius: 99, width: 0, transition: 'width 1.2s ease', background: gradient }}></div>;
}
