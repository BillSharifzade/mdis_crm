import { useState, useCallback, useEffect } from 'react';
import Sidebar from './components/Sidebar';
import Topbar from './components/Topbar';
import Dashboard from './pages/Dashboard';
import Leads from './pages/Leads';
import Pipeline from './pages/Pipeline';
import Analytics from './pages/Analytics';
import Reports from './pages/Reports';
import Settings from './pages/Settings';
import MyKpi from './pages/MyKpi';
import LeadModal from './components/LeadModal';
import LeadImportModal from './components/LeadImportModal';
import MergeLeadsModal from './components/MergeLeadsModal';
import DetailSidebar from './components/DetailSidebar';
import NotifPanel from './components/NotifPanel';
import ToastContainer from './components/ToastContainer';
import Login from './pages/Login';
import { getStatusObj } from './data/crmData';
import { api, mapServerToClientLead, mapServerToClientInteraction, STATUS_ID_TO_KEY } from './services/api';
import { useNotif } from './context/useNotif';
import { UsersProvider } from './context/UsersContext';
import { ProgramsProvider } from './context/ProgramsContext';
import { StagesProvider } from './context/StagesContext';
import { openEventStream } from './services/eventStream';

const USE_API = api.useApi;

const PAGE_TITLES = {
  dashboard: 'Дашборд',
  leads: 'База лидов',
  pipeline: 'Воронка продаж',
  analytics: 'Аналитика',
  reports: 'Отчёты',
  mykpi: 'Мой KPI',
  settings: 'Настройки',
};

const ROLE_PAGES = {
  admin: ['dashboard', 'leads', 'pipeline', 'analytics', 'reports', 'mykpi', 'settings'],
  // Admissions работает с лидами и видит только свои показатели — без org-аналитики.
  admissions: ['dashboard', 'leads', 'pipeline', 'mykpi'],
  guest: ['dashboard', 'analytics', 'reports'],
};

const PAGE_STORAGE_KEY = 'crm_current_page';
const ALL_PAGES = ['dashboard', 'leads', 'pipeline', 'analytics', 'reports', 'mykpi', 'settings'];

function App() {
  const [currentPage, setCurrentPage] = useState(() => {
    try {
      const saved = localStorage.getItem(PAGE_STORAGE_KEY);
      return saved && ALL_PAGES.includes(saved) ? saved : 'dashboard';
    } catch { return 'dashboard'; }
  });
  const [isAuthenticated, setIsAuthenticated] = useState(() => !USE_API || api.isAuthenticated());
  const [user, setUser] = useState(() => api.currentUser());
  const [allLeads, setAllLeads] = useState([]);
  const [isLoading, setIsLoading] = useState(USE_API && isAuthenticated);
  const [loadError, setLoadError] = useState(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [importOpen, setImportOpen] = useState(false);
  const [mergeOpen, setMergeOpen] = useState(false);
  const [detailLead, setDetailLead] = useState(null);
  const [notifOpen, setNotifOpen] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [toasts, setToasts] = useState([]);
  const [globalSearch, setGlobalSearch] = useState('');
  const [unreadByLead, setUnreadByLead] = useState({});

  const { push: pushNotif, unreadCount } = useNotif();

  const role = (user && user.role) || 'admin';
  const allowedPages = ROLE_PAGES[role] || ROLE_PAGES.admin;

  const showToast = useCallback((message, type = 'success') => {
    const id = Date.now() + Math.random();
    setToasts(prev => [...prev, { id, message, type }]);
    setTimeout(() => {
      setToasts(prev => prev.filter(t => t.id !== id));
    }, 3000);
  }, []);

  const navigate = useCallback((pageId) => {
    if (!allowedPages.includes(pageId)) {
      showToast('Недостаточно прав для просмотра раздела', 'warning');
      return;
    }
    setCurrentPage(pageId);
    try { localStorage.setItem(PAGE_STORAGE_KEY, pageId); } catch { /* */ }
    setSidebarOpen(false);
  }, [allowedPages, showToast]);

  const refreshLeads = useCallback(async () => {
    if (!USE_API) return;
    try {
      const data = await api.getLeads();
      setAllLeads(data);
    } catch (error) {
      console.error(error);
      const code = error?.status ?? 0;
      showToast(`Не удалось обновить список лидов (код ${code}): ${error?.message || 'неизвестная ошибка'}`, 'error');
    }
  }, [showToast]);

  // Загрузка лидов. Никаких демо-данных: при ошибке показываем код + описание и кнопку «Повторить».
  const loadLeads = useCallback(async () => {
    if (!USE_API || !isAuthenticated) return;
    setIsLoading(true);
    setLoadError(null);
    try {
      const data = await api.getLeads();
      setAllLeads(data);
    } catch (error) {
      console.error(error);
      setAllLeads([]);
      setLoadError({ code: error?.status ?? 0, message: error?.message || 'Неизвестная ошибка' });
    } finally {
      setIsLoading(false);
    }
  }, [isAuthenticated]);

  useEffect(() => { loadLeads(); }, [loadLeads]);

  // Тост-сводка KPI при входе для роли admissions.
  useEffect(() => {
    if (!USE_API || !isAuthenticated) return;
    if (!user || user.role !== 'admissions') return;
    let cancelled = false;
    (async () => {
      try {
        const targets = await api.getMyKpiTargets();
        if (cancelled) return;
        const period = (Array.isArray(targets) && targets.length > 0)
          ? targets.reduce((m, t) => (t.period_days < m ? t.period_days : m), targets[0].period_days)
          : 30;
        const s = await api.getMyKpi(period);
        if (cancelled || !s) return;
        const periodLabel = s.period_days === 1 ? '1 день' : `${s.period_days} дн.`;
        const lines = [];
        if (s.processed_goal > 0) {
          const pct = Math.round((s.processed / s.processed_goal) * 100);
          lines.push(`Обработка: ${s.processed}/${s.processed_goal} (${pct}%)`);
        }
        if (s.created_goal > 0) {
          const pct = Math.round((s.created / s.created_goal) * 100);
          lines.push(`Создано: ${s.created}/${s.created_goal} (${pct}%)`);
        }
        if (lines.length === 0) {
          showToast(`KPI · ${periodLabel}: цели не назначены`, 'info');
        } else {
          showToast(`Твой KPI · ${periodLabel} — ${lines.join(' · ')}`, 'info');
        }
      } catch { /* ignore */ }
    })();
    return () => { cancelled = true; };
  }, [isAuthenticated, user, showToast]);

  // Telegram unread badges per chat — поллинг 15с.
  useEffect(() => {
    if (!USE_API || !isAuthenticated) return;
    let stopped = false;
    const tick = async () => {
      try {
        const arr = await api.getTelegramUnread();
        if (stopped) return;
        const next = {};
        arr.forEach(it => { next[it.lead_id] = it.unread; });
        setUnreadByLead(next);
      } catch { /* */ }
    };
    tick();
    const iv = setInterval(tick, 15000);
    return () => { stopped = true; clearInterval(iv); };
  }, [isAuthenticated]);

  // ── LIVE EVENTS (SSE) ─────────────────────────────────────────────────
  useEffect(() => {
    if (!USE_API || !isAuthenticated) return;
    const stream = openEventStream({
      onEvent: ({ type, data }) => {
        if (!data) return;
        const payload = data.payload;
        if (type === 'lead.created' && payload) {
          const lead = mapServerToClientLead(payload);
          let alreadyKnown = false;
          setAllLeads(prev => {
            alreadyKnown = prev.some(l => l.id === lead.id);
            return alreadyKnown ? prev : [lead, ...prev];
          });
          if (!alreadyKnown) {
            pushNotif({ type: 'lead_created', text: `<strong>Новый лид:</strong> ${lead.name}${lead.source ? ' · ' + lead.source : ''}`, leadId: lead.id });
          }
        } else if (type === 'lead.updated' && payload) {
          const lead = mapServerToClientLead(payload);
          setAllLeads(prev => prev.map(l => l.id === lead.id ? { ...l, ...lead } : l));
        } else if (type === 'lead.deleted' && payload?.id) {
          setAllLeads(prev => prev.filter(l => l.id !== payload.id));
        } else if (type === 'lead.status_changed' && payload?.lead_id) {
          const statusKey = STATUS_ID_TO_KEY[payload.new_status_id] || 'new';
          setAllLeads(prev => prev.map(l => l.id === payload.lead_id ? { ...l, status: statusKey, refusalReason: payload.refusal_reason || l.refusalReason } : l));
          const stObj = getStatusObj(statusKey);
          pushNotif({ type: statusKey === 'enrolled' ? 'enrolled' : (statusKey === 'lost' ? 'lost' : 'status_change'), text: `Лид #${payload.lead_id} → <strong>${stObj.label}</strong>`, leadId: payload.lead_id });
        } else if (type === 'interaction.created' && payload?.lead_id) {
          const inter = mapServerToClientInteraction(payload);
          pushNotif({ type: 'interaction', text: `<strong>${inter.label}</strong> · лид #${payload.lead_id}`, leadId: payload.lead_id });
        } else if (type === 'telegram.message' && payload?.lead_id) {
          if (payload.direction === 'inbound') {
            pushNotif({ type: 'telegram_chat', text: `<strong>Telegram:</strong> новое сообщение от лида #${payload.lead_id}`, leadId: payload.lead_id });
          }
        }
      },
      onError: (e) => console.warn('SSE closed', e),
    });
    return () => { if (stream) stream.close(); };
  }, [isAuthenticated, pushNotif]);
  // ─────────────────────────────────────────────────────────────────────


  const addLead = useCallback(async (newLead) => {
    try {
      let created;
      if (USE_API) {
        created = await api.createLead(newLead);
        setAllLeads(prev => [created, ...prev]);
      } else {
        created = { ...newLead, id: Date.now() };
        setAllLeads(prev => [created, ...prev]);
      }
      showToast(`Лид «${newLead.name}» добавлен`, 'success');
      pushNotif({
        type: 'lead_created',
        text: `<strong>Новый лид:</strong> ${newLead.name}${newLead.source ? ' из ' + newLead.source : ''}`,
        leadId: created.id,
      });
      return true;
    } catch (error) {
      console.error(error);
      showToast('Ошибка при добавлении лида', 'error');
      return false;
    }
  }, [showToast, pushNotif]);

  const updateLead = useCallback(async (id, updatedLead) => {
    try {
      let saved;
      if (USE_API) {
        saved = await api.updateLead(id, updatedLead);
      } else {
        saved = { ...updatedLead, id };
      }
      setAllLeads(prev => prev.map(l => l.id === id ? { ...l, ...saved } : l));
      showToast('Карточка обновлена', 'success');
      pushNotif({ type: 'user_action', text: `<strong>Карточка</strong> #${id} обновлена`, leadId: id });
      return true;
    } catch (error) {
      console.error(error);
      showToast('Ошибка обновления: ' + error.message, 'error');
      return false;
    }
  }, [showToast, pushNotif]);

  const deleteLead = useCallback(async (id) => {
    try {
      if (USE_API) await api.deleteLead(id);
      setAllLeads(prev => prev.filter(l => l.id !== id));
      showToast('Лид удалён', 'success');
      pushNotif({ type: 'user_action', text: `<strong>Лид</strong> #${id} удалён` });
      return true;
    } catch (error) {
      console.error(error);
      showToast('Ошибка удаления: ' + error.message, 'error');
      return false;
    }
  }, [showToast, pushNotif]);

  const changeLeadStatus = useCallback(async (id, status, refusalReason = '') => {
    if (status === 'lost' && !refusalReason) {
      showToast('Укажите причину отказа', 'warning');
      return false;
    }
    try {
      if (USE_API) {
        await api.updateLeadStatus(id, status, refusalReason);
      }
      let leadName = '';
      setAllLeads(prev => prev.map(l => {
        if (l.id === id) {
          leadName = l.name;
          return { ...l, status, refusalReason };
        }
        return l;
      }));
      const stObj = getStatusObj(status);
      const notifType = status === 'enrolled' ? 'enrolled' : (status === 'lost' ? 'lost' : 'status_change');
      pushNotif({
        type: notifType,
        text: `<strong>${leadName || 'Лид'}</strong> → статус «${stObj.label}»${refusalReason ? ` (${refusalReason})` : ''}`,
        leadId: id,
      });
      showToast('Статус обновлён', 'success');
      return true;
    } catch (error) {
      console.error(error);
      showToast('Ошибка при обновлении статуса', 'error');
      return false;
    }
  }, [showToast, pushNotif]);

  const handleLogout = useCallback(() => {
    api.logout();
    setIsAuthenticated(false);
    setUser(null);
    setAllLeads([]);
    setCurrentPage('dashboard');
    showToast('Вы вышли из системы', 'info');
  }, [showToast]);

  const handleImportDone = useCallback(async (count) => {
    showToast(`Импортировано лидов: ${count}`, 'success');
    pushNotif({ type: 'lead_created', text: `<strong>Импорт Excel:</strong> добавлено ${count} лидов` });
    setImportOpen(false);
    await refreshLeads();
  }, [refreshLeads, showToast, pushNotif]);

  const handleMergeDone = useCallback(async () => {
    showToast('Лиды объединены', 'success');
    pushNotif({ type: 'lead_merged', text: '<strong>Слияние:</strong> дубликаты объединены' });
    setMergeOpen(false);
    setDetailLead(null);
    await refreshLeads();
  }, [refreshLeads, showToast, pushNotif]);

  const openDetail = useCallback((lead) => {
    setDetailLead(lead);
  }, []);

  const renderPage = () => {
    if (!allowedPages.includes(currentPage)) {
      return (
        <div className="card" style={{ padding: 32, textAlign: 'center' }}>
          <h2 style={{ color: 'var(--text-primary)' }}>Доступ запрещён</h2>
          <p style={{ color: 'var(--text-secondary)', marginTop: 8 }}>У вашей роли «{role}» нет прав на этот раздел.</p>
        </div>
      );
    }
    switch (currentPage) {
      case 'dashboard':
        return <Dashboard allLeads={allLeads} navigate={navigate} openDetail={openDetail} user={user} />;
      case 'leads':
        return <Leads
          allLeads={allLeads}
          openDetail={openDetail}
          openModal={() => setModalOpen(true)}
          openImport={() => setImportOpen(true)}
          openMerge={() => setMergeOpen(true)}
          role={role}
          showToast={showToast}
          externalSearch={globalSearch}
          onExternalSearchChange={setGlobalSearch}
          onStatusChange={changeLeadStatus}
          unreadByLead={unreadByLead}
        />;
      case 'pipeline':
        return <Pipeline
          allLeads={allLeads}
          openDetail={openDetail}
          openModal={() => setModalOpen(true)}
          onStatusChange={changeLeadStatus}
          role={role}
        />;
      case 'analytics':
        return <Analytics allLeads={allLeads} />;
      case 'reports':
        return <Reports showToast={showToast} />;
      case 'mykpi':
        return <MyKpi user={user} showToast={showToast} />;
      case 'settings':
        return <Settings allLeads={allLeads} showToast={showToast} role={role} />;
      default:
        return <Dashboard allLeads={allLeads} navigate={navigate} openDetail={openDetail} user={user} />;
    }
  };

  const handleLoginSuccess = useCallback(() => {
    setIsAuthenticated(true);
    setUser(api.currentUser());
    pushNotif({ type: 'user_action', text: '<strong>Вход</strong> в систему' });
  }, [pushNotif]);

  if (!isAuthenticated) {
    return (
      <>
        <Login onLogin={handleLoginSuccess} showToast={showToast} />
        <ToastContainer toasts={toasts} />
      </>
    );
  }

  return (
    <UsersProvider enabled={isAuthenticated}>
    <ProgramsProvider enabled={isAuthenticated}>
    <StagesProvider enabled={isAuthenticated}>
      <div className="bg-orb orb-1"></div>
      <div className="bg-orb orb-2"></div>
      <div className="bg-orb orb-3"></div>

      <Sidebar
        currentPage={currentPage}
        navigate={navigate}
        isOpen={sidebarOpen}
        allLeads={allLeads}
        user={user}
        role={role}
        allowedPages={allowedPages}
        onLogout={handleLogout}
      />

      <div className="main-wrapper">
        <Topbar
          breadcrumb={PAGE_TITLES[currentPage] || currentPage}
          onMenuToggle={() => setSidebarOpen(prev => !prev)}
          onNotifToggle={() => setNotifOpen(prev => !prev)}
          searchValue={globalSearch}
          unreadCount={unreadCount}
          hideSearch={currentPage === 'leads'}
          onSearch={(q) => {
            setGlobalSearch(q);
            if (q && currentPage !== 'leads') navigate('leads');
          }}
        />

        <main className="page-content">
          {loadError ? (
            <div className="card" style={{ padding: 32, textAlign: 'center', maxWidth: 540, margin: '40px auto' }}>
              <h2 style={{ color: 'var(--text-primary)' }}>Не удалось загрузить данные</h2>
              <p style={{ color: '#ef4444', fontSize: 34, fontWeight: 700, margin: '12px 0 4px' }}>
                {loadError.code ? `Код ${loadError.code}` : 'Нет соединения'}
              </p>
              <p style={{ color: 'var(--text-secondary)', marginBottom: 20, wordBreak: 'break-word' }}>
                {loadError.message}
              </p>
              <button className="btn btn-primary" onClick={loadLeads}>Повторить</button>
            </div>
          ) : isLoading ? (
            <div className="flex items-center justify-center p-12">
              <div style={{ color: 'var(--text-primary)' }}>Загрузка данных...</div>
            </div>
          ) : (
            renderPage()
          )}
        </main>
      </div>

      {modalOpen && (
        <LeadModal
          onClose={() => setModalOpen(false)}
          onSave={addLead}
          showToast={showToast}
        />
      )}

      {importOpen && (
        <LeadImportModal
          onClose={() => setImportOpen(false)}
          onDone={handleImportDone}
          showToast={showToast}
        />
      )}

      {mergeOpen && (
        <MergeLeadsModal
          leads={allLeads}
          onClose={() => setMergeOpen(false)}
          onDone={handleMergeDone}
          showToast={showToast}
        />
      )}

      {detailLead && (
        <DetailSidebar
          lead={allLeads.find(l => l.id === detailLead.id) || detailLead}
          onClose={() => setDetailLead(null)}
          onStatusChange={changeLeadStatus}
          onUpdate={updateLead}
          onDelete={async (id) => {
            const ok = await deleteLead(id);
            if (ok) setDetailLead(null);
            return ok;
          }}
          showToast={showToast}
          role={role}
          onNotif={pushNotif}
        />
      )}

      <NotifPanel
        isOpen={notifOpen}
        onClose={() => setNotifOpen(false)}
      />

      <ToastContainer toasts={toasts} />
    </StagesProvider>
    </ProgramsProvider>
    </UsersProvider>
  );
}

export default App;
