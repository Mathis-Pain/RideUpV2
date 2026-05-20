import { serve } from 'https://deno.land/std@0.177.0/http/server.ts'
import { createClient } from 'https://esm.sh/@supabase/supabase-js@2'

const FCM_PROJECT_ID = Deno.env.get('FCM_PROJECT_ID')!
const FCM_SERVICE_ACCOUNT_KEY = JSON.parse(Deno.env.get('FCM_SERVICE_ACCOUNT_KEY')!)
const SUPABASE_URL = Deno.env.get('SUPABASE_URL')!
const SUPABASE_SERVICE_ROLE_KEY = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY')!

serve(async (req) => {
  const { record } = await req.json()
  // record = new row in messages table
  const eventId: string = record.event_id
  const senderId: string = record.user_id
  const content: string = record.content

  const supabase = createClient(SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY)

  // Get sender username
  const { data: sender } = await supabase
    .from('profiles')
    .select('username')
    .eq('id', senderId)
    .single()

  // Get all participants of this event except the sender
  const { data: participants } = await supabase
    .from('event_participants')
    .select('user_id')
    .eq('event_id', eventId)
    .neq('user_id', senderId)

  if (!participants?.length) return new Response('no recipients', { status: 200 })

  const recipientIds = participants.map((p: { user_id: string }) => p.user_id)

  // Get their FCM tokens
  const { data: tokens } = await supabase
    .from('device_tokens')
    .select('token')
    .in('user_id', recipientIds)

  if (!tokens?.length) return new Response('no tokens', { status: 200 })

  const accessToken = await getAccessToken()

  await Promise.allSettled(
    tokens.map((t: { token: string }) =>
      sendFcmMessage(accessToken, t.token, {
        title: sender?.username ?? 'RideUp',
        body: content,
        data: { type: 'chat', event_id: eventId },
      })
    )
  )

  return new Response('ok', { status: 200 })
})

async function getAccessToken(): Promise<string> {
  const now = Math.floor(Date.now() / 1000)
  const header = btoa(JSON.stringify({ alg: 'RS256', typ: 'JWT' }))
  const payload = btoa(
    JSON.stringify({
      iss: FCM_SERVICE_ACCOUNT_KEY.client_email,
      scope: 'https://www.googleapis.com/auth/firebase.messaging',
      aud: 'https://oauth2.googleapis.com/token',
      iat: now,
      exp: now + 3600,
    })
  )

  const toSign = `${header}.${payload}`
  const key = await importPrivateKey(FCM_SERVICE_ACCOUNT_KEY.private_key)
  const signature = await crypto.subtle.sign(
    { name: 'RSASSA-PKCS1-v1_5' },
    key,
    new TextEncoder().encode(toSign)
  )
  const jwt = `${toSign}.${btoa(String.fromCharCode(...new Uint8Array(signature)))}`

  const resp = await fetch('https://oauth2.googleapis.com/token', {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: `grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer&assertion=${jwt}`,
  })
  const { access_token } = await resp.json()
  return access_token
}

async function importPrivateKey(pem: string): Promise<CryptoKey> {
  const pemContents = pem
    .replace(/-----BEGIN PRIVATE KEY-----/, '')
    .replace(/-----END PRIVATE KEY-----/, '')
    .replace(/\n/g, '')
  const binaryDer = Uint8Array.from(atob(pemContents), (c) => c.charCodeAt(0))
  return crypto.subtle.importKey(
    'pkcs8',
    binaryDer,
    { name: 'RSASSA-PKCS1-v1_5', hash: 'SHA-256' },
    false,
    ['sign']
  )
}

async function sendFcmMessage(
  accessToken: string,
  token: string,
  { title, body, data }: { title: string; body: string; data: Record<string, string> }
) {
  return fetch(
    `https://fcm.googleapis.com/v1/projects/${FCM_PROJECT_ID}/messages:send`,
    {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        message: {
          token,
          notification: { title, body },
          data,
        },
      }),
    }
  )
}
