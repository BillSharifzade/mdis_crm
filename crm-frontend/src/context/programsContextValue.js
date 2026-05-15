import { createContext } from 'react';

export const ProgramsContext = createContext({ programs: [], byId: new Map(), byName: new Map() });
