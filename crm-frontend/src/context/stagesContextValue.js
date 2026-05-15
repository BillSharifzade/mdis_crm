import { createContext } from 'react';

export const StagesContext = createContext({
    stages: [],
    byId: new Map(),
    byKey: new Map(),
    reload: () => {},
});
