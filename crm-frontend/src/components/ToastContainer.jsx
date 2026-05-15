import { CheckCircle, Info, AlertTriangle, XCircle } from 'lucide-react';

const iconMap = {
    success: CheckCircle,
    info: Info,
    warning: AlertTriangle,
    error: XCircle,
};

const colorMap = {
    success: { bg: 'rgba(16,185,129,0.15)', col: '#10b981' },
    info: { bg: 'rgba(99,102,241,0.15)', col: '#818cf8' },
    warning: { bg: 'rgba(245,158,11,0.15)', col: '#f59e0b' },
    error: { bg: 'rgba(239,68,68,0.15)', col: '#ef4444' },
};

export default function ToastContainer({ toasts }) {
    if (!toasts.length) return null;

    return (
        <div className="toast-container">
            {toasts.map(t => {
                const Icon = iconMap[t.type] || CheckCircle;
                const colors = colorMap[t.type] || colorMap.success;
                return (
                    <div className="toast" key={t.id}>
                        <div className="toast-icon" style={{ background: colors.bg }}>
                            <Icon size={15} style={{ color: colors.col }} />
                        </div>
                        <span className="toast-text">{t.message}</span>
                    </div>
                );
            })}
        </div>
    );
}
