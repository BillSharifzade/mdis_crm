import { useState, useMemo } from 'react';
import { GraduationCap, ChevronDown, ChevronRight, Eye } from 'lucide-react';
import { academicYear, formatDate } from '../data/crmData';

// Архив зачисленных студентов (#7 ТЗ). Лиды со статусом «Зачисление» (enrolled)
// автоматически попадают сюда и раскладываются по папкам учебных годов.
export default function Archive({ allLeads, openDetail }) {
    const enrolled = useMemo(
        () => allLeads.filter(l => l.status === 'enrolled'),
        [allLeads]
    );

    // Группируем по учебному году (по дате зачисления, с фолбэком на дату заявки).
    const groups = useMemo(() => {
        const map = new Map();
        enrolled.forEach(l => {
            const when = l.enrolledAt || l.date;
            const key = academicYear(when);
            if (!map.has(key)) map.set(key, []);
            map.get(key).push(l);
        });
        // Сортируем года по убыванию (свежие сверху).
        return Array.from(map.entries()).sort((a, b) => b[0].localeCompare(a[0]));
    }, [enrolled]);

    const [collapsed, setCollapsed] = useState({});
    const toggle = (key) => setCollapsed(prev => ({ ...prev, [key]: !prev[key] }));

    return (
        <section className="page active">
            <div className="page-header">
                <div>
                    <h1>Архив зачисленных</h1>
                    <p className="page-sub">{enrolled.length} студентов · {groups.length} учебных {groups.length === 1 ? 'год' : 'лет'}</p>
                </div>
            </div>

            {groups.length === 0 && (
                <div className="card" style={{ padding: 40, textAlign: 'center', color: 'var(--text-muted)' }}>
                    Пока нет зачисленных студентов. Как только статус лида меняется на «Зачисление»,
                    он автоматически попадает в архив.
                </div>
            )}

            {groups.map(([year, leads]) => {
                const isCollapsed = collapsed[year];
                return (
                    <div className="card" key={year} style={{ marginBottom: 16, overflow: 'hidden' }}>
                        <button
                            onClick={() => toggle(year)}
                            style={{
                                width: '100%', display: 'flex', alignItems: 'center', gap: 10,
                                padding: '14px 18px', background: 'var(--glass)', border: 'none',
                                borderBottom: isCollapsed ? 'none' : '1px solid var(--glass-border)',
                                cursor: 'pointer', color: 'var(--text-primary)', fontFamily: 'inherit',
                            }}
                        >
                            {isCollapsed ? <ChevronRight size={16} /> : <ChevronDown size={16} />}
                            <GraduationCap size={16} style={{ color: '#10b981' }} />
                            <span style={{ fontWeight: 700, fontSize: 14 }}>{year} учебный год</span>
                            <span style={{ marginLeft: 'auto', fontSize: 12, color: 'var(--text-muted)' }}>{leads.length} студентов</span>
                        </button>

                        {!isCollapsed && (
                            <table className="leads-table">
                                <thead>
                                    <tr>
                                        <th>Студент</th>
                                        <th>Программа</th>
                                        <th>Источник</th>
                                        <th>Дата зачисления</th>
                                        <th>Действия</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {leads.map(l => (
                                        <tr key={l.id} onClick={() => openDetail(l)} style={{ cursor: 'pointer' }}>
                                            <td>
                                                <div className="lead-cell">
                                                    <div className="lead-avatar" style={{ background: l.color, color: '#fff', fontSize: 11, fontWeight: 700 }}>{l.initials}</div>
                                                    <div>
                                                        <div className="lead-cell-name">{l.name}</div>
                                                        <div className="lead-cell-phone">{l.phone}</div>
                                                    </div>
                                                </div>
                                            </td>
                                            <td style={{ color: 'var(--text-secondary)' }}>{l.program}</td>
                                            <td><span className="source-chip">{l.source}</span></td>
                                            <td style={{ color: 'var(--text-muted)' }}>{formatDate(l.enrolledAt || l.date)}</td>
                                            <td onClick={e => e.stopPropagation()}>
                                                <div className="action-btns">
                                                    <button className="action-btn" onClick={() => openDetail(l)} title="Открыть"><Eye size={13} /></button>
                                                </div>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        )}
                    </div>
                );
            })}
        </section>
    );
}
