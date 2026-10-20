import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { deleteCredential, listCredentials, registerBegin, registerFinish } from '@/api/webauthn'
import { registerCredential } from '@/lib/webauthn'

const QUERY_KEY = ['webauthn-credentials']

export function useWebAuthnCredentials(enabled = true) {
  return useQuery({ queryKey: QUERY_KEY, queryFn: listCredentials, enabled })
}

/** Runs the full enrollment ceremony: begin -> browser prompt (Face ID/Touch ID/...) -> finish. */
export function useEnrollWebAuthnCredential() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (name: string) => {
      const { options, challenge_token } = await registerBegin()
      const credential = await registerCredential(options)
      return registerFinish(challenge_token, name, credential)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEY })
    },
  })
}

export function useDeleteWebAuthnCredential() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteCredential(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEY })
    },
  })
}
