import pkg from '../../package.json'

/** The app version, read from package.json so the manifest stays the single source of truth. */
export const APP_VERSION: string = pkg.version
