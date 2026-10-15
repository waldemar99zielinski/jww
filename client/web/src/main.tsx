import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { App } from '@/App'
import { Providers } from '@/providers'
import { registerServiceWorker } from '@/lib/pwa'
import '@/index.css'

// Registered before render so an installed app can boot its shell offline. No-op outside prod.
registerServiceWorker()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <Providers>
        <App />
      </Providers>
    </BrowserRouter>
  </StrictMode>,
)
