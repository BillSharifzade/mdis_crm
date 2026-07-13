// Небольшой стор черновиков форм в localStorage. Нужен, чтобы случайно закрытый
// диалог (клик по фону / крестик / «Отмена») не терял введённые данные —
// при повторном открытии форма восстанавливается. Черновик чистится только
// после успешного сохранения.

export function loadDraft(key, fallback) {
    try {
        const raw = localStorage.getItem(key);
        if (!raw) return fallback;
        const parsed = JSON.parse(raw);
        if (!parsed || typeof parsed !== 'object') return fallback;
        // Мержим поверх дефолтов: если структура формы поменялась, новые поля
        // возьмутся из fallback, а сохранённые — из черновика.
        return { ...fallback, ...parsed };
    } catch { return fallback; }
}

export function hasDraft(key) {
    try { return !!localStorage.getItem(key); } catch { return false; }
}

export function saveDraft(key, data) {
    try { localStorage.setItem(key, JSON.stringify(data)); } catch { /* quota — ignore */ }
}

export function clearDraft(key) {
    try { localStorage.removeItem(key); } catch { /* ignore */ }
}
