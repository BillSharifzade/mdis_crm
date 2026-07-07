import { useContext } from 'react';
import { SourcesContext } from './sourcesContextValue.js';

export function useSources() {
    return useContext(SourcesContext);
}
