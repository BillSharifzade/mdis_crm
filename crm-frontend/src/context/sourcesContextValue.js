import { createContext } from 'react';

// options: [{ id, name, label }] — уже с человекочитаемой подписью и дедупом.
export const SourcesContext = createContext({ sources: [], options: [] });
