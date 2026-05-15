import { useContext } from 'react';
import { StagesContext } from './stagesContextValue.js';

export function useStages() {
    return useContext(StagesContext);
}
