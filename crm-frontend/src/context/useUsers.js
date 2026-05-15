import { useContext } from 'react';
import { UsersContext } from './usersContextValue.js';

export function useUsers() {
    return useContext(UsersContext);
}
