import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { App } from '@/App'
import { Providers } from '@/providers'
import { registerServiceWorker } from '@/lib/pwa'
import { APP_VERSION } from '@/lib/version'
import '@/index.css'

// Logged on every boot so a deployed build is identifiable from the console.
console.info(`Joint Wallet Watch v${APP_VERSION}`)

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
