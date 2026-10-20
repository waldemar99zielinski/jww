import { createAsyncThunk, createSlice } from '@reduxjs/toolkit'
import { getMe, login as loginRequest } from '@/api/auth'
import { ApiError } from '@/api/client'
import { MOCK_USER } from '@/api/mock/data'
import type { UserDto } from '@/api/types'
import { loginBegin, loginFinish } from '@/api/webauthn'
import { isDesignMode } from '@/lib/designMode'
import { getAuthToken, setAuthToken } from '@/lib/authToken'
import { authenticateDiscoverable } from '@/lib/webauthn'

export type AuthStatus = 'loading' | 'authenticated' | 'unauthenticated' | 'error'

export interface AuthState {
  status: AuthStatus
  user: UserDto | null
  error: string | null
}

const initialState: AuthState = isDesignMode()
  ? { status: 'authenticated', user: MOCK_USER, error: null }
  : { status: getAuthToken() ? 'loading' : 'unauthenticated', user: null, error: null }

function errorMessage(error: unknown): string {
  if (error instanceof ApiError) return error.status === 401 ? 'Invalid username or password.' : error.message
  return 'Something went wrong — try again.'
}

/** Why a stored-session restore failed — only an auth failure may clear the token. */
export type RestoreFailureReason = 'unauthorized' | 'network'

/** No-ops (via `condition`) in design mode (already "authenticated" from initialState) or when there's no stored token. */
export const restoreSession = createAsyncThunk<UserDto, void, { rejectValue: RestoreFailureReason }>(
  'auth/restoreSession',
  async (_, { rejectWithValue }) => {
    try {
      return await getMe()
    } catch (error) {
      // A network failure (offline) rejects exactly like a 401 — distinguish them so a blip
      // doesn't destroy the stored session.
      const unauthorized = error instanceof ApiError && (error.status === 401 || error.status === 403)
      return rejectWithValue(unauthorized ? 'unauthorized' : 'network')
    }
  },
  { condition: () => !isDesignMode() && getAuthToken() !== null },
)

export const login = createAsyncThunk<UserDto, { username: string; password: string }, { rejectValue: string }>(
  'auth/login',
  async ({ username, password }, { rejectWithValue }) => {
    if (isDesignMode()) return MOCK_USER
    try {
      const { access_token } = await loginRequest(username, password)
      setAuthToken(access_token)
      return await getMe()
    } catch (error) {
      return rejectWithValue(errorMessage(error))
    }
  },
)

/** Face ID/Touch ID login via the discoverable WebAuthn ceremony — see `lib/webauthn.ts`. */
export const loginWithPasskey = createAsyncThunk<UserDto, void, { rejectValue: string }>(
  'auth/loginWithPasskey',
  async (_, { rejectWithValue }) => {
    try {
      const { options, challenge_token } = await loginBegin()
      const credential = await authenticateDiscoverable(options)
      const { access_token } = await loginFinish(challenge_token, credential)
      setAuthToken(access_token)
      return await getMe()
    } catch (error) {
      if (error instanceof ApiError) {
        return rejectWithValue(
          error.status === 401 ? 'Face ID sign-in failed — try again or use your password.' : error.message,
        )
      }
      // A thrown DOMException from navigator.credentials.get (the user cancelled the
      // prompt, no platform authenticator is enrolled, etc.) is not an ApiError —
      // never a server rejection, so there's no anti-enumeration reason to collapse it.
      return rejectWithValue('Face ID sign-in was cancelled or is unavailable.')
    }
  },
)

export const logout = createAsyncThunk('auth/logout', () => {
  setAuthToken(null)
})

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder
      .addCase(restoreSession.pending, (state) => {
        state.status = 'loading'
      })
      .addCase(restoreSession.fulfilled, (state, action) => {
        state.status = 'authenticated'
        state.user = action.payload
        state.error = null
      })
      .addCase(restoreSession.rejected, (state, action) => {
        // Only a real auth failure clears the session. A network failure keeps the stored token
        // and surfaces the existing 'error' state, so an offline blip cannot sign the user out.
        const unauthorized = action.payload === 'unauthorized'
        if (unauthorized) setAuthToken(null)
        state.status = unauthorized ? 'unauthenticated' : 'error'
        state.user = null
      })
      .addCase(login.pending, (state) => {
        state.status = 'loading'
        state.error = null
      })
      .addCase(login.fulfilled, (state, action) => {
        state.status = 'authenticated'
        state.user = action.payload
        state.error = null
      })
      .addCase(login.rejected, (state, action) => {
        state.status = 'error'
        state.user = null
        state.error = action.payload ?? 'Something went wrong — try again.'
      })
      .addCase(loginWithPasskey.pending, (state) => {
        state.status = 'loading'
        state.error = null
      })
      .addCase(loginWithPasskey.fulfilled, (state, action) => {
        state.status = 'authenticated'
        state.user = action.payload
        state.error = null
      })
      .addCase(loginWithPasskey.rejected, (state, action) => {
        state.status = 'error'
        state.user = null
        state.error = action.payload ?? 'Something went wrong — try again.'
      })
      .addCase(logout.fulfilled, (state) => {
        state.status = 'unauthenticated'
        state.user = null
        state.error = null
      })
  },
})

export const authReducer = authSlice.reducer
