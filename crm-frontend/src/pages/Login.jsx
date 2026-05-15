import { useState } from 'react';
import { api } from '../services/api';
import { Lock, Mail } from 'lucide-react';

export default function Login({ onLogin, showToast }) {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [loading, setLoading] = useState(false);

    const handleLogin = async (e) => {
        e.preventDefault();
        setLoading(true);
        try {
            await api.login(email, password);
            showToast('Успешный вход', 'success');
            onLogin();
        } catch (error) {
            console.error(error);
            showToast('Ошибка авторизации. Проверьте логин и пароль.', 'error');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '100vh', width: '100vw', background: 'var(--bg-card)' }}>
            <div className="bg-orb orb-1"></div>
            <div className="bg-orb orb-2"></div>
            <div className="bg-orb orb-3"></div>

            <div className="card" style={{ maxWidth: '400px', width: '100%', margin: '0 20px', position: 'relative', zIndex: 10 }}>
                <div className="card-header" style={{ display: 'block', textAlign: 'center', borderBottom: 'none', paddingBottom: '0' }}>
                    <img src="/logo-4.png" alt="MDIS Dushanbe" style={{ height: 70, marginBottom: 14, objectFit: 'contain' }} />
                    <h2 style={{ fontSize: '22px', fontWeight: '700', color: 'var(--text-primary)', marginBottom: '6px' }}>Приёмная комиссия</h2>
                    <p style={{ color: 'var(--text-secondary)', fontSize: '13px' }}>Введите данные для доступа к CRM</p>
                </div>

                <div className="card-body" style={{ padding: '24px' }}>
                    <form onSubmit={handleLogin}>
                        <div className="form-group" style={{ marginBottom: '16px' }}>
                            <label className="form-label" style={{ display: 'block', marginBottom: '8px', fontSize: '13px', color: 'var(--text-primary)' }}>Email</label>
                            <div style={{ position: 'relative' }}>
                                <Mail size={16} style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-secondary)' }} />
                                <input
                                    type="email"
                                    className="form-control"
                                    style={{ paddingLeft: '36px', width: '100%' }}
                                    placeholder="admin@example.com"
                                    value={email}
                                    onChange={e => setEmail(e.target.value)}
                                    required
                                />
                            </div>
                        </div>

                        <div className="form-group" style={{ marginBottom: '24px' }}>
                            <label className="form-label" style={{ display: 'block', marginBottom: '8px', fontSize: '13px', color: 'var(--text-primary)' }}>Пароль</label>
                            <div style={{ position: 'relative' }}>
                                <Lock size={16} style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-secondary)' }} />
                                <input
                                    type="password"
                                    className="form-control"
                                    style={{ paddingLeft: '36px', width: '100%' }}
                                    placeholder="••••••••"
                                    value={password}
                                    onChange={e => setPassword(e.target.value)}
                                    required
                                />
                            </div>
                        </div>

                        <button
                            type="submit"
                            className="btn btn-primary"
                            style={{ width: '100%', justifyContent: 'center' }}
                            disabled={loading}
                        >
                            {loading ? 'Вход...' : 'Войти'}
                        </button>
                    </form>
                </div>
            </div>
        </div>
    );
}
