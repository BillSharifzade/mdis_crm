/* ====================================================
   MDIS CRM — Data & Utilities (shared across components)
   ==================================================== */

export const STATUSES = [
    { key: 'new', label: 'Новая заявка', cls: 's-new', color: '#6366f1' },
    { key: 'consultation', label: 'Консультация', cls: 's-consult', color: '#06b6d4' },
    { key: 'documents', label: 'Сбор документов', cls: 's-docs', color: '#f59e0b' },
    { key: 'exams', label: 'Экзамены', cls: 's-exam', color: '#a855f7' },
    { key: 'payment', label: 'Оплата / договор', cls: 's-payment', color: '#ec4899' },
    { key: 'enrolled', label: 'Зачисление', cls: 's-enrolled', color: '#10b981' },
    { key: 'lost', label: 'Проиграно', cls: 's-lost', color: '#ef4444' },
];

export const SOURCES = ['Сайт', 'Telegram', 'WhatsApp', 'Instagram', 'VK', 'Email', 'Звонок', 'Реклама Яндекс'];
export const PROGRAMS = ['Бизнес-администрирование', 'Информационные технологии', 'Финансы и банковское дело', 'Маркетинг', 'Право', 'Бухгалтерский учёт'];
export const MANAGERS = ['АК', 'ДО', 'СИ'];
export const MANAGER_NAMES = ['Алина Кравцова', 'Денис Орлов', 'Сабина Ибрагимова'];
export const MANAGER_COLORS = ['#6366f1', '#10b981', '#f59e0b'];

export const AVATAR_COLORS = ['#6366f1', '#10b981', '#f59e0b', '#ec4899', '#06b6d4', '#a855f7', '#ef4444', '#8b5cf6'];

const FIRST_NAMES = ['Алия', 'Дмитрий', 'Камила', 'Руслан', 'Зарина', 'Иван', 'Бахром', 'Анна', 'Нигора', 'Максим', 'Диана', 'Тимур', 'Элина', 'Артём', 'Сабина', 'Дилноза', 'Фарида', 'Никита', 'Алишер', 'Юлия'];
const LAST_NAMES = ['Каримова', 'Орлов', 'Хасанова', 'Петров', 'Набиева', 'Смирнов', 'Юсупов', 'Иванова', 'Расулова', 'Козлов', 'Сидорова', 'Мухаммедов', 'Назарова', 'Новиков', 'Ибрагимова', 'Рахимова', 'Тошматова', 'Гришин', 'Азимов', 'Волкова'];

export function randEl(arr) { return arr[Math.floor(Math.random() * arr.length)]; }
export function randInt(a, b) { return Math.floor(Math.random() * (b - a + 1)) + a; }
export function getInitials(name) { return name.split(' ').slice(0, 2).map(p => p[0]).join('').toUpperCase(); }

export function getStatusObj(key) { return STATUSES.find(s => s.key === key) || STATUSES[0]; }

function generateInteractions(n) {
    const types = [
        { icon: 'phone', label: 'Звонок', bg: 'rgba(99,102,241,0.15)', color: '#818cf8' },
        { icon: 'message-circle', label: 'Сообщение', bg: 'rgba(6,182,212,0.15)', color: '#06b6d4' },
        { icon: 'mail', label: 'Email', bg: 'rgba(245,158,11,0.15)', color: '#f59e0b' },
    ];
    const notes = [
        'Клиент заинтересован, просит перезвонить завтра',
        'Отправлен список документов для поступления',
        'Уточнил стоимость программы MBA',
        'Попросил информацию об общежитии',
        'Подтвердил участие в дне открытых дверей',
        'Запросил расписание вступительных испытаний',
        'Интересовался скидками для семейных абитуриентов',
    ];
    const result = [];
    for (let i = 0; i < n; i++) {
        const t = randEl(types);
        const hoursAgo = randInt(1, 240);
        result.push({ ...t, note: randEl(notes), hoursAgo });
    }
    return result.sort((a, b) => a.hoursAgo - b.hoursAgo);
}

export function generateLeads(n) {
    const leads = [];
    const statKeys = STATUSES.map(s => s.key);
    const weights = [30, 25, 15, 10, 8, 7, 5];
    const cumWeights = weights.reduce((acc, w, i) => { acc.push((acc[i - 1] || 0) + w); return acc; }, []);

    for (let i = 0; i < n; i++) {
        const fn = randEl(FIRST_NAMES), ln = randEl(LAST_NAMES);
        const name = `${fn} ${ln}`;
        const r = randInt(1, 100);
        const statIdx = cumWeights.findIndex(c => r <= c);
        const statusKey = statKeys[Math.min(statIdx, statKeys.length - 1)];
        const daysAgo = randInt(0, 90);
        const date = new Date(); date.setDate(date.getDate() - daysAgo);
        leads.push({
            id: i + 1,
            name,
            initials: getInitials(name),
            phone: `+992 ${randInt(90, 99)} ${randInt(100, 999)} ${randInt(10, 99)} ${randInt(10, 99)}`,
            email: `${fn.toLowerCase()}.${ln.toLowerCase()}@mail.ru`,
            source: randEl(SOURCES),
            program: randEl(PROGRAMS),
            status: statusKey,
            managerIdx: randInt(0, 2),
            date,
            color: randEl(AVATAR_COLORS),
            interactions: generateInteractions(randInt(1, 5)),
        });
    }
    return leads;
}

export function formatDate(d) {
    const date = d instanceof Date ? d : new Date(d);
    if (isNaN(date.getTime())) return '—';
    return date.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short' });
}

export function formatHoursAgo(h) {
    if (h < 1) return 'только что';
    if (h < 24) return `${h} ч назад`;
    const days = Math.floor(h / 24);
    return `${days} дн назад`;
}
