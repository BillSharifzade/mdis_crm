import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App.jsx'
import { ThemeProvider } from './context/ThemeContext.jsx'
import { NotifProvider } from './context/NotifContext.jsx'
import './index.css'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <ThemeProvider>
      <NotifProvider>
        <App />
      </NotifProvider>
    </ThemeProvider>
  </StrictMode>,
)
