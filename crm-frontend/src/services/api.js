import { AVATAR_COLORS, getInitials } from '../data/crmData';

// Базовый URL API задаётся через VITE_API_BASE_URL в .env (по умолчанию — localhost для разработки).
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081/api/v1';

// Включение/выключение реального API. Можно отключить через VITE_USE_API=false для работы на моках.
const USE_API_ENV = import.meta.env.VITE_USE_API;
const USE_API = USE_API_ENV === undefined ? true : USE_API_ENV !== 'false';

// Маппинг статусов клиент ↔ сервер. id этапов стабильны (см. миграцию 000012):
// id4 теперь «EET», id8 — «Обращение» (key 'inquiry'). Подписи берутся из
// pipeline_stages, здесь только связь ключ ↔ id.
const STATUS_TO_API = {
    'new': 1, 'consultation': 2, 'documents': 3, 'exams': 4, 'payment': 5, 'enrolled': 6, 'lost': 7, 'inquiry': 8
};
const API_TO_STATUS = {
    1: 'new', 2: 'consultation', 3: 'documents', 4: 'exams', 5: 'payment', 6: 'enrolled', 7: 'lost', 8: 'inquiry'
};

// Динамические карты этап↔id, заполняются из pipeline_stages сервера
// (см. StagesContext.setStageMaps). Нужны, чтобы КАСТОМНЫЕ этапы, созданные
// админом в Settings (id ≥ 9, произвольные названия), корректно
// маппились в обе стороны. Без этого статус таких этапов молча сбрасывался
// в id=1 «Новая заявка», а бейдж показывался неверным.
let DYN_ID_TO_KEY = {};
let DYN_KEY_TO_ID = {};

export function setStageMaps(idToKey, keyToId) {
    DYN_ID_TO_KEY = idToKey || {};
    DYN_KEY_TO_ID = keyToId || {};
}

// statusIdToKey — id этапа → стабильный ключ. Приоритет: динамическая карта
// с сервера → статическая (базовые этапы) → синтетический stage-<id>.
export function statusIdToKey(statusId) {
    if (statusId == null) return 'new';
    if (DYN_ID_TO_KEY[statusId]) return DYN_ID_TO_KEY[statusId];
    if (API_TO_STATUS[statusId]) return API_TO_STATUS[statusId];
    return `stage-${statusId}`;
}

// statusKeyToId — ключ этапа → numeric id для PATCH. Обратный приоритет.
// null означает «неизвестный ключ» (карты ещё не загрузились) — вызывающий
// решает, что делать.
export function statusKeyToId(statusStr) {
    if (statusStr == null) return null;
    if (statusStr in DYN_KEY_TO_ID) return DYN_KEY_TO_ID[statusStr];
    if (statusStr in STATUS_TO_API) return STATUS_TO_API[statusStr];
    const m = /^stage-(\d+)$/.exec(String(statusStr));
    if (m) return Number(m[1]);
    return null;
}

const SOURCE_LABELS = {
    site: 'Сайт', telegram: 'Telegram', whatsapp: 'WhatsApp',
    instagram: 'Instagram', vk: 'VK', email: 'Email',
    phone: 'Звонок', yandex_ads: 'Реклама Яндекс', organic: 'Сайт',
    campus_visit: 'Кампус-визит',
};

export const mapServerToClientLead = (serverLead) => {
    const fullName = `${serverLead.first_name || ''} ${serverLead.last_name || ''}`.trim() || 'Без имени';
    const created = serverLead.created_at ? new Date(serverLead.created_at) : new Date();
    return {
        id: serverLead.id,
        name: fullName,
        initials: getInitials(fullName),
        phone: serverLead.phone || '',
        email: serverLead.email || '',
        source: SOURCE_LABELS[serverLead.utm_source] || serverLead.utm_source || 'Сайт',
        program: serverLead.program_name || '—',
        programId: serverLead.program_id || null,
        status: statusIdToKey(serverLead.status_id),
        statusId: serverLead.status_id || null,
        managerIdx: serverLead.assignee_id ? Math.abs(serverLead.assignee_id) % 3 : 0,
        assigneeId: serverLead.assignee_id || null,
        date: created,
        color: AVATAR_COLORS[(serverLead.id || 0) % AVATAR_COLORS.length] || AVATAR_COLORS[0],
        interactions: [],
        refusalReason: serverLead.refusal_reason || '',
        englishLevel: serverLead.english_level || '',
        telegramId: serverLead.telegram_id || '',
        whatsappId: serverLead.whatsapp_id || '',
        vkId: serverLead.vk_id || '',
        socialUrl: serverLead.social_url || '',
        paymentStatus: serverLead.payment_status || '',
        reminderAt: serverLead.reminder_at || null,
        reminderNote: serverLead.reminder_note || '',
        reminderDone: !!serverLead.reminder_done,
        workCompany: serverLead.work_company || '',
        workPosition: serverLead.work_position || '',
        enrolledAt: serverLead.enrolled_at || null,
    };
};

export const mapServerToClientInteraction = (serverInt) => {
    let icon = 'phone';
    let bg = 'rgba(124, 58, 237, 0.15)';
    let color = '#a78bfa';
    let label = 'Звонок';

    const t = serverInt.type;
    if (t === 'sms' || t === 'message' || t === 'messenger') {
        icon = 'message-circle'; bg = 'rgba(52, 211, 153, 0.15)'; color = '#34d399'; label = 'Сообщение';
    } else if (t === 'note' || t === 'manual_note') {
        icon = 'message-circle'; bg = 'rgba(245, 158, 11, 0.15)'; color = '#f59e0b'; label = 'Заметка';
    } else if (t === 'email') {
        icon = 'mail'; bg = 'rgba(56, 189, 248, 0.15)'; color = '#38bdf8'; label = 'Email';
    } else if (t === 'task') {
        icon = 'message-circle'; bg = 'rgba(168, 85, 247, 0.15)'; color = '#a855f7'; label = 'Задача';
    }

    const createdTime = new Date(serverInt.created_at || new Date().toISOString()).getTime();
    const hoursAgo = Math.max(0, Math.floor((Date.now() - createdTime) / (1000 * 60 * 60)));

    // Для звонков — дописываем outcome + длительность
    let note = serverInt.content || '';
    if (t === 'call') {
        if (serverInt.outcome === 'answered') {
            label = 'Звонок (успешный)';
            if (serverInt.duration_minutes) note = `${note} · ${serverInt.duration_minutes} мин`;
        } else if (serverInt.outcome === 'no_answer') {
            label = 'Звонок (без ответа)';
            color = '#f59e0b';
        }
    }

    return {
        id: serverInt.id,
        icon, bg, color, label,
        note,
        hoursAgo,
        outcome: serverInt.outcome || '',
        durationMinutes: serverInt.duration_minutes || 0,
    };
};

// Единый тип ошибки API: всегда несёт числовой код и человекочитаемое описание.
// status === 0 — транспортная/сетевая ошибка (DNS, TLS, обрыв соединения, CORS).
export class ApiError extends Error {
    constructor(status, message) {
        super(message);
        this.name = 'ApiError';
        this.status = status;
    }
}

const getAuthHeaders = () => {
    const token = localStorage.getItem('crm_token');
    return token ? { 'Authorization': `Bearer ${token}` } : {};
};

// Обёртка над fetch: сетевые/транспортные сбои превращаем в ApiError(0, ...),
// чтобы наверх ВСЕГДА уходила ошибка с кодом и описанием — никаких тихих фолбэков.
const safeFetch = async (url, options) => {
    try {
        return await fetch(url, options);
    } catch (err) {
        throw new ApiError(0, `Сеть недоступна — ${err?.message || 'не удалось соединиться с сервером'}`);
    }
};

const fetchWithAuth = async (endpoint, options = {}) => {
    const headers = { ...options.headers, ...getAuthHeaders() };
    return safeFetch(`${API_BASE_URL}${endpoint}`, { ...options, headers });
};

const handleResponse = async (response) => {
    if (!response.ok) {
        if (response.status === 401) {
            localStorage.removeItem('crm_token');
            localStorage.removeItem('crm_user');
        }
        let body = '';
        try { body = (await response.text()).trim(); } catch { /* ignore */ }
        const desc = body || response.statusText || 'Ошибка запроса';
        throw new ApiError(response.status, `HTTP ${response.status} — ${desc}`);
    }
    const ct = response.headers.get('content-type') || '';
    if (ct.includes('application/json')) return response.json();
    return response.text();
};

export const STATUS_ID_TO_KEY = API_TO_STATUS;

export const api = {
    useApi: USE_API,
    baseUrl: API_BASE_URL,

    // --- AUTH ---
    login: async (email, password) => {
        const response = await safeFetch(`${API_BASE_URL}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password }),
        });
        const data = await handleResponse(response);
        if (data.token) {
            localStorage.setItem('crm_token', data.token);
            if (data.user) localStorage.setItem('crm_user', JSON.stringify(data.user));
        }
        return data;
    },

    logout: () => {
        localStorage.removeItem('crm_token');
        localStorage.removeItem('crm_user');
    },

    isAuthenticated: () => !!localStorage.getItem('crm_token'),

    currentUser: () => {
        try { return JSON.parse(localStorage.getItem('crm_user') || 'null'); }
        catch { return null; }
    },

    // --- LEADS ---
    getLeads: async (limit = 200, offset = 0) => {
        const response = await fetchWithAuth(`/leads?limit=${limit}&offset=${offset}`);
        const result = await handleResponse(response);
        const leads = Array.isArray(result) ? result : (result.data || []);
        return leads.map(mapServerToClientLead);
    },

    getLeadById: async (id) => {
        const response = await fetchWithAuth(`/leads/${id}`);
        const data = await handleResponse(response);
        return mapServerToClientLead(data);
    },

    createLead: async (clientLeadData) => {
        const names = (clientLeadData.name || '').trim().split(/\s+/);
        const sourceKey = Object.keys(SOURCE_LABELS).find(k => SOURCE_LABELS[k] === clientLeadData.source) || 'site';
        const serverPayload = {
            first_name: names[0] || 'Новый',
            last_name: names.slice(1).join(' '),
            email: clientLeadData.email || '',
            phone: clientLeadData.phone || '',
            utm_source: sourceKey,
            utm_medium: 'referral',
            utm_campaign: 'none',
        };
        if (clientLeadData.programId) serverPayload.program_id = clientLeadData.programId;
        if (clientLeadData.socialUrl) serverPayload.social_url = clientLeadData.socialUrl;
        if (clientLeadData.englishLevel) serverPayload.english_level = clientLeadData.englishLevel;
        if (clientLeadData.assigneeId) serverPayload.assignee_id = clientLeadData.assigneeId;
        if (clientLeadData.paymentStatus) serverPayload.payment_status = clientLeadData.paymentStatus;
        if (clientLeadData.reminderAt) serverPayload.reminder_at = new Date(clientLeadData.reminderAt).toISOString();
        if (clientLeadData.reminderNote) serverPayload.reminder_note = clientLeadData.reminderNote;
        if (clientLeadData.workCompany) serverPayload.work_company = clientLeadData.workCompany;
        if (clientLeadData.workPosition) serverPayload.work_position = clientLeadData.workPosition;
        const response = await fetchWithAuth(`/leads`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(serverPayload),
        });
        const data = await handleResponse(response);
        return mapServerToClientLead(data);
    },

    updateLeadStatus: async (id, statusStr, refusalReason = '') => {
        const status_id = statusKeyToId(statusStr);
        if (status_id == null) {
            throw new Error(`Неизвестный этап воронки: ${statusStr}`);
        }
        const payload = { status_id };
        if (refusalReason) payload.refusal_reason = refusalReason;
        const response = await fetchWithAuth(`/leads/${id}/status`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        });
        return await handleResponse(response);
    },

    completeReminder: async (id) => {
        const response = await fetchWithAuth(`/leads/${id}/reminder/complete`, { method: 'POST' });
        return await handleResponse(response);
    },

    mergeLeads: async (primaryId, duplicateId) => {
        const response = await fetchWithAuth(`/leads/merge`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ target_lead_id: primaryId, source_lead_id: duplicateId }),
        });
        return await handleResponse(response);
    },

    importLeads: async (file) => {
        const formData = new FormData();
        formData.append('file', file);
        const response = await safeFetch(`${API_BASE_URL}/leads/import`, {
            method: 'POST',
            headers: { ...getAuthHeaders() },
            body: formData,
        });
        return await handleResponse(response);
    },

    deleteLead: async (id) => {
        const response = await fetchWithAuth(`/leads/${id}`, { method: 'DELETE' });
        return await handleResponse(response);
    },

    updateLead: async (id, clientLead) => {
        const names = (clientLead.name || '').trim().split(/\s+/);
        const sourceKey = Object.keys(SOURCE_LABELS).find(k => SOURCE_LABELS[k] === clientLead.source) || 'site';
        const payload = {
            first_name: names[0] || '—',
            last_name: names.slice(1).join(' '),
            email: clientLead.email || '',
            phone: clientLead.phone || '',
            utm_source: sourceKey,
        };
        if (clientLead.assigneeId != null) payload.assignee_id = clientLead.assigneeId;
        if (clientLead.programId != null) payload.program_id = clientLead.programId;
        if (clientLead.socialUrl !== undefined) payload.social_url = clientLead.socialUrl;
        if (clientLead.englishLevel !== undefined) payload.english_level = clientLead.englishLevel;
        if (clientLead.paymentStatus !== undefined) payload.payment_status = clientLead.paymentStatus;
        if (clientLead.workCompany !== undefined) payload.work_company = clientLead.workCompany;
        if (clientLead.workPosition !== undefined) payload.work_position = clientLead.workPosition;
        // Напоминание: очистка имеет приоритет над установкой.
        if (clientLead.clearReminder) {
            payload.clear_reminder = true;
        } else if (clientLead.reminderAt) {
            payload.reminder_at = new Date(clientLead.reminderAt).toISOString();
            if (clientLead.reminderNote !== undefined) payload.reminder_note = clientLead.reminderNote;
        }
        const response = await fetchWithAuth(`/leads/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        });
        const data = await handleResponse(response);
        return mapServerToClientLead(data);
    },

    getPrograms: async () => {
        const response = await fetchWithAuth(`/programs`);
        const data = await handleResponse(response);
        return Array.isArray(data) ? data : [];
    },

    // ── STAGES (этапы воронки, read-only для всех залогиненных) ──
    getStages: async () => {
        const response = await fetchWithAuth(`/stages`);
        const data = await handleResponse(response);
        return Array.isArray(data) ? data : [];
    },

    // ── SETTINGS (admin) ──
    createProgram: async (body) => {
        const r = await fetchWithAuth(`/settings/programs`, {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
        });
        return handleResponse(r);
    },
    updateProgram: async (id, body) => {
        const r = await fetchWithAuth(`/settings/programs/${id}`, {
            method: 'PUT', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
        });
        return handleResponse(r);
    },
    deleteProgram: async (id) => {
        const r = await fetchWithAuth(`/settings/programs/${id}`, { method: 'DELETE' });
        return handleResponse(r);
    },
    listAllPrograms: async () => {
        const r = await fetchWithAuth(`/settings/programs?all=1`);
        const data = await handleResponse(r);
        return Array.isArray(data) ? data : [];
    },
    listSources: async () => {
        const r = await fetchWithAuth(`/settings/sources?all=1`);
        const data = await handleResponse(r);
        return Array.isArray(data) ? data : [];
    },
    createSource: async (body) => {
        const r = await fetchWithAuth(`/settings/sources`, {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
        });
        return handleResponse(r);
    },
    updateSource: async (id, body) => {
        const r = await fetchWithAuth(`/settings/sources/${id}`, {
            method: 'PUT', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
        });
        return handleResponse(r);
    },
    deleteSource: async (id) => {
        const r = await fetchWithAuth(`/settings/sources/${id}`, { method: 'DELETE' });
        return handleResponse(r);
    },
    listAllStages: async () => {
        const r = await fetchWithAuth(`/settings/stages?all=1`);
        const data = await handleResponse(r);
        return Array.isArray(data) ? data : [];
    },
    createStage: async (body) => {
        const r = await fetchWithAuth(`/settings/stages`, {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
        });
        return handleResponse(r);
    },
    updateStage: async (id, body) => {
        const r = await fetchWithAuth(`/settings/stages/${id}`, {
            method: 'PUT', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
        });
        return handleResponse(r);
    },
    deleteStage: async (id) => {
        const r = await fetchWithAuth(`/settings/stages/${id}`, { method: 'DELETE' });
        return handleResponse(r);
    },
    reorderStages: async (orderedList) => {
        const r = await fetchWithAuth(`/settings/stages/reorder`, {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(orderedList),
        });
        return handleResponse(r);
    },
    getBotSettings: async () => {
        const r = await fetchWithAuth(`/settings/bot`);
        return handleResponse(r);
    },
    setBotSetting: async (key, value) => {
        const r = await fetchWithAuth(`/settings/bot/${encodeURIComponent(key)}`, {
            method: 'PUT', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ value }),
        });
        return handleResponse(r);
    },

    // ── KPI (admin) ──
    getKpiStats: async (userId, period = 30) => {
        const r = await fetchWithAuth(`/kpi/users/${userId}/stats?period=${period}`);
        return handleResponse(r);
    },
    listKpiTargets: async (userId) => {
        const r = await fetchWithAuth(`/kpi/users/${userId}/targets`);
        return handleResponse(r);
    },
    upsertKpiTarget: async (userId, metric, body) => {
        const r = await fetchWithAuth(`/kpi/users/${userId}/targets/${metric}`, {
            method: 'PUT', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
        });
        return handleResponse(r);
    },
    getMyKpi: async (period = 30) => {
        const r = await fetchWithAuth(`/me/kpi?period=${period}`);
        return handleResponse(r);
    },
    getMyKpiTargets: async () => {
        const r = await fetchWithAuth(`/me/kpi/targets`);
        const data = await handleResponse(r);
        return Array.isArray(data) ? data : [];
    },
    deleteKpiTarget: async (targetId) => {
        const r = await fetchWithAuth(`/kpi/targets/${targetId}`, { method: 'DELETE' });
        return handleResponse(r);
    },

    // ── ANALYTICS (расширения) ──
    getSourceBreakdown: async () => {
        const r = await fetchWithAuth(`/analytics/sources`);
        const data = await handleResponse(r);
        return Array.isArray(data) ? data : [];
    },
    getFunnel: async () => {
        const r = await fetchWithAuth(`/analytics/funnel`);
        const data = await handleResponse(r);
        return Array.isArray(data) ? data : [];
    },
    getTimeSeries: async (days = 56) => {
        const r = await fetchWithAuth(`/analytics/timeseries?days=${days}`);
        const data = await handleResponse(r);
        return Array.isArray(data) ? data : [];
    },

    // ── Telegram unread per chat ──
    getTelegramUnread: async () => {
        const r = await fetchWithAuth(`/telegram-chat/unread`);
        const data = await handleResponse(r);
        return Array.isArray(data) ? data : [];
    },

    // ── Call log (T9) ──
    logCall: async ({ leadId, outcome, comment, durationMinutes }) => {
        const payload = {
            lead_id: leadId,
            type: 'call',
            direction: 'outbound',
            content: comment || (outcome === 'answered' ? 'Звонок состоялся' : 'Без ответа'),
            outcome,
        };
        if (outcome === 'answered' && durationMinutes > 0) {
            payload.duration_minutes = durationMinutes;
            payload.duration_seconds = durationMinutes * 60;
        }
        const r = await fetchWithAuth(`/interactions`, {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        });
        return handleResponse(r);
    },

    // --- INTERACTIONS ---
    getInteractions: async (leadId) => {
        const response = await fetchWithAuth(`/interactions/lead/${leadId}`);
        const data = await handleResponse(response);
        return Array.isArray(data) ? data.map(mapServerToClientInteraction) : [];
    },

    createInteraction: async (leadId, type, content) => {
        const response = await fetchWithAuth(`/interactions`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ lead_id: leadId, type, content, duration_seconds: 0 }),
        });
        const data = await handleResponse(response);
        return mapServerToClientInteraction(data);
    },

    getStatusHistory: async (leadId) => {
        const response = await fetchWithAuth(`/status-history/lead/${leadId}`);
        const data = await handleResponse(response);
        return Array.isArray(data) ? data : [];
    },

    // --- USERS ---
    getUsers: async () => {
        const response = await fetchWithAuth(`/users`);
        const data = await handleResponse(response);
        return Array.isArray(data) ? data : [];
    },

    // Слим read-only список назначаемых менеджеров (id/name/role).
    // Доступен всем ролям — в отличие от admin-only getUsers().
    getManagers: async () => {
        const response = await fetchWithAuth(`/managers`);
        const data = await handleResponse(response);
        return Array.isArray(data) ? data : [];
    },

    createUser: async (userData) => {
        const response = await fetchWithAuth(`/users`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(userData),
        });
        return await handleResponse(response);
    },

    updateUser: async (id, userData) => {
        const response = await fetchWithAuth(`/users/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(userData),
        });
        return await handleResponse(response);
    },

    deleteUser: async (id) => {
        const response = await fetchWithAuth(`/users/${id}`, { method: 'DELETE' });
        return await handleResponse(response);
    },

    // --- ANALYTICS ---
    getDashboardStats: async () => {
        const response = await fetchWithAuth(`/analytics/dashboard`);
        return await handleResponse(response);
    },

    getConversionStats: async () => {
        const response = await fetchWithAuth(`/analytics/conversion`);
        return await handleResponse(response);
    },

    getManagerKpi: async () => {
        const response = await fetchWithAuth(`/analytics/kpi`);
        return await handleResponse(response);
    },

    // --- TELEGRAM CHAT ---
    getTelegramStatus: async (leadId) => {
        const response = await fetchWithAuth(`/telegram-chat/${leadId}/status`);
        return await handleResponse(response);
    },

    getTelegramMessages: async (leadId) => {
        const response = await fetchWithAuth(`/telegram-chat/${leadId}/messages`);
        const data = await handleResponse(response);
        return Array.isArray(data) ? data : [];
    },

    sendTelegramMessage: async (leadId, text) => {
        const response = await fetchWithAuth(`/telegram-chat/${leadId}/send`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ text }),
        });
        return await handleResponse(response);
    },

    takeoverTelegram: async (leadId) => {
        const response = await fetchWithAuth(`/telegram-chat/${leadId}/takeover`, {
            method: 'POST',
        });
        return await handleResponse(response);
    },

    exportAnalytics: async (format = 'xlsx', range = {}) => {
        const params = new URLSearchParams({ format });
        if (range.from) params.set('from', range.from);
        if (range.to) params.set('to', range.to);
        const response = await fetchWithAuth(`/analytics/export?${params.toString()}`);
        if (!response.ok) {
            let msg = '';
            try { msg = await response.text(); } catch { /* */ }
            throw new Error(msg || 'Ошибка экспорта');
        }
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `leads_export_${new Date().toISOString().slice(0, 10)}.${format}`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
    },
};

export default api;
