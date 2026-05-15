// Lightweight wrapper around the SSE endpoint /api/v1/events.
// EventSource не умеет ставить заголовки, поэтому JWT уходит через ?token=…

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081/api/v1';

export function openEventStream({ onEvent, onError } = {}) {
    const token = localStorage.getItem('crm_token');
    if (!token) return null;

    const url = `${API_BASE_URL}/events?token=${encodeURIComponent(token)}`;
    let es;
    try {
        es = new EventSource(url);
    } catch (err) {
        if (onError) onError(err);
        return null;
    }

    const KNOWN_EVENTS = [
        'hello',
        'lead.created',
        'lead.updated',
        'lead.deleted',
        'lead.status_changed',
        'interaction.created',
        'telegram.message',
    ];

    KNOWN_EVENTS.forEach(type => {
        es.addEventListener(type, e => {
            try {
                const data = e.data ? JSON.parse(e.data) : null;
                onEvent && onEvent({ type, data });
            } catch (err) {
                console.warn('SSE parse failed', type, err);
            }
        });
    });

    es.onerror = (err) => {
        // Browser auto-reconnects after readyState=CONNECTING; surface only fatal closes.
        if (es.readyState === EventSource.CLOSED && onError) onError(err);
    };

    return {
        close: () => es.close(),
        raw: es,
    };
}
