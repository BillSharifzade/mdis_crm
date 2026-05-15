import { useState } from 'react';
import { X, Upload, FileSpreadsheet } from 'lucide-react';
import { api } from '../services/api';

export default function LeadImportModal({ onClose, onDone, showToast }) {
    const [file, setFile] = useState(null);
    const [uploading, setUploading] = useState(false);

    const handleFileChange = (e) => {
        const f = e.target.files && e.target.files[0];
        if (!f) return;
        const ok = /\.(xlsx|xls|csv)$/i.test(f.name);
        if (!ok) {
            showToast('Поддерживаются только файлы .xlsx, .xls, .csv', 'warning');
            return;
        }
        setFile(f);
    };

    const handleUpload = async () => {
        if (!file) {
            showToast('Выберите файл', 'warning');
            return;
        }
        if (!api.useApi) {
            showToast('Импорт работает только при подключённом API', 'warning');
            return;
        }
        setUploading(true);
        try {
            const result = await api.importLeads(file);
            const count = (result && (result.imported || result.count || result.created)) || 0;
            onDone(count);
        } catch (err) {
            console.error(err);
            showToast('Ошибка импорта: ' + err.message, 'error');
        } finally {
            setUploading(false);
        }
    };

    return (
        <div className="modal-overlay open" onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}>
            <div className="modal" style={{ maxWidth: 520 }}>
                <div className="modal-header">
                    <h2>Импорт лидов из Excel</h2>
                    <button className="modal-close" onClick={onClose}><X size={14} /></button>
                </div>
                <div className="modal-body">
                    <p style={{ color: 'var(--text-secondary)', fontSize: 13, marginBottom: 16 }}>
                        Шаблон: <b>«База лидов_03.04.2026.xlsx»</b> — поддерживаются листы PCIE / EET / FYC / MBA / Leads_26
                        с колонками <code>Day, name, number, source, type, status, Programme, Follow-up action</code>.
                        Даты Excel-serial (например, 45953) распознаются автоматически. Привязка к программе — по имени
                        листа, источник — из колонки <code>source</code>. Лиды распределяются round-robin.
                    </p>

                    <label
                        htmlFor="lead-import-file"
                        style={{
                            display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 12,
                            padding: 32, borderRadius: 12,
                            border: '2px dashed var(--glass-border)',
                            background: 'var(--glass)',
                            cursor: 'pointer',
                            transition: 'border-color 0.2s',
                        }}
                    >
                        <FileSpreadsheet size={32} style={{ color: 'var(--text-secondary)' }} />
                        {file ? (
                            <>
                                <div style={{ fontWeight: 700, color: 'var(--text-primary)', fontSize: 14 }}>{file.name}</div>
                                <div style={{ color: 'var(--text-muted)', fontSize: 12 }}>{(file.size / 1024).toFixed(1)} КБ</div>
                            </>
                        ) : (
                            <>
                                <div style={{ fontWeight: 700, color: 'var(--text-primary)', fontSize: 14 }}>Нажмите, чтобы выбрать файл</div>
                                <div style={{ color: 'var(--text-muted)', fontSize: 12 }}>.xlsx, .xls, .csv</div>
                            </>
                        )}
                        <input
                            id="lead-import-file"
                            type="file"
                            accept=".xlsx,.xls,.csv"
                            onChange={handleFileChange}
                            style={{ display: 'none' }}
                        />
                    </label>
                </div>
                <div className="modal-footer">
                    <button className="btn btn-outline" onClick={onClose} disabled={uploading}>Отмена</button>
                    <button className="btn btn-primary" onClick={handleUpload} disabled={!file || uploading}>
                        <Upload size={14} /> {uploading ? 'Загрузка...' : 'Импортировать'}
                    </button>
                </div>
            </div>
        </div>
    );
}
