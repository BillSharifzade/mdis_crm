import { LayoutDashboard, Users, Kanban, BarChart2, FileText, Settings, LogOut, Sun, Moon, Target } from 'lucide-react';
import { useTheme } from '../context/ThemeContext';
import { getInitials } from '../data/crmData';

const navItems = [
    {
        section: 'Основное', items: [
            { id: 'dashboard', icon: LayoutDashboard, label: 'Дашборд' },
            { id: 'leads', icon: Users, label: 'Лиды', badge: 'count' },
            { id: 'pipeline', icon: Kanban, label: 'Воронка' },
        ]
    },
    {
        section: 'Аналитика', items: [
            { id: 'analytics', icon: BarChart2, label: 'Аналитика' },
            { id: 'reports', icon: FileText, label: 'Отчёты' },
            { id: 'mykpi', icon: Target, label: 'Мой KPI' },
        ]
    },
    {
        section: 'Система', items: [
            { id: 'settings', icon: Settings, label: 'Настройки' },
        ]
    },
];

const ROLE_LABELS = {
    admin: 'Администратор',
    admissions: 'Менеджер приёма',
    guest: 'Гость (только просмотр)',
};

export default function Sidebar({ currentPage, navigate, isOpen, allLeads, user, role = 'admin', allowedPages, onLogout }) {
    const { theme, toggleTheme } = useTheme();
    const leadCount = (allLeads || []).filter(l => l.status === 'new').length;
    const displayName = (user && user.name) || 'Пользователь';
    const displayEmail = (user && user.email) || '';
    const roleLabel = ROLE_LABELS[role] || role;
    const allowed = allowedPages || ['dashboard', 'leads', 'pipeline', 'analytics', 'reports', 'mykpi', 'settings'];

    return (
        <aside className={`sidebar ${isOpen ? 'mobile-open' : ''}`} id="sidebar">
            <div className="sidebar-logo">
                <img src={`${import.meta.env.BASE_URL}logo-4.png`} alt="MDIS" className="sidebar-logo-img" />
                <span className="logo-crm-text">CRM</span>
            </div>

            <nav className="sidebar-nav">
                {navItems.map(section => {
                    const visible = section.items.filter(it => allowed.includes(it.id));
                    if (visible.length === 0) return null;
                    return (
                        <div key={section.section}>
                            <div className="nav-section-label">{section.section}</div>
                            {visible.map(item => {
                                const Icon = item.icon;
                                return (
                                    <a
                                        href="#"
                                        key={item.id}
                                        className={`nav-item ${currentPage === item.id ? 'active' : ''}`}
                                        onClick={(e) => { e.preventDefault(); navigate(item.id); }}
                                    >
                                        <Icon size={16} />
                                        <span>{item.label}</span>
                                        {item.badge === 'count' && leadCount > 0 && <span className="nav-badge">{leadCount}</span>}
                                    </a>
                                );
                            })}
                        </div>
                    );
                })}
            </nav>

            <div className="theme-toggle-wrapper">
                <button className="theme-toggle-btn" onClick={toggleTheme} title={theme === 'dark' ? 'Светлая тема' : 'Тёмная тема'}>
                    {theme === 'dark' ? <Sun size={16} /> : <Moon size={16} />}
                    <span>{theme === 'dark' ? 'Светлая тема' : 'Тёмная тема'}</span>
                </button>
            </div>

            <div className="sidebar-user">
                <div className="user-avatar">{getInitials(displayName)}</div>
                <div className="user-info">
                    <span className="user-name">{displayName}</span>
                    <span className="user-role" title={displayEmail}>{roleLabel}</span>
                </div>
                <button className="user-logout" title="Выйти" onClick={onLogout}>
                    <LogOut size={14} />
                </button>
            </div>
        </aside>
    );
}
