import { createContext } from 'react';

export const UsersContext = createContext({ users: [], byId: new Map(), reload: () => { } });
