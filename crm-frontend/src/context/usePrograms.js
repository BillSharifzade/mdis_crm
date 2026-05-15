import { useContext } from 'react';
import { ProgramsContext } from './programsContextValue.js';

export function usePrograms() {
    return useContext(ProgramsContext);
}
