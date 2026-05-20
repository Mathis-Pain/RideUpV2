-- ============================================================
-- device_tokens — FCM push tokens par utilisateur
-- À exécuter dans Supabase SQL Editor
-- ============================================================

CREATE TABLE IF NOT EXISTS public.device_tokens (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  token      TEXT NOT NULL,
  platform   TEXT NOT NULL CHECK (platform IN ('ios', 'android')),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE (user_id)
);

ALTER TABLE public.device_tokens ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Upsert son propre token"
  ON public.device_tokens FOR ALL
  USING (auth.uid() = user_id)
  WITH CHECK (auth.uid() = user_id);

-- ============================================================
-- Triggers pour déclencher les Edge Functions
-- ============================================================

-- Trigger: nouveau message → send-chat-notification
CREATE OR REPLACE FUNCTION notify_chat_message()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER AS $$
BEGIN
  PERFORM net.http_post(
    url    := current_setting('app.supabase_url') || '/functions/v1/send-chat-notification',
    body   := json_build_object('record', row_to_json(NEW))::text,
    params := '{}',
    headers := json_build_object(
      'Content-Type', 'application/json',
      'Authorization', 'Bearer ' || current_setting('app.service_role_key')
    )::text
  );
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS on_message_insert ON public.messages;
CREATE TRIGGER on_message_insert
  AFTER INSERT ON public.messages
  FOR EACH ROW EXECUTE FUNCTION notify_chat_message();

-- Trigger: nouvelle alerte SOS → send-sos-notification
CREATE OR REPLACE FUNCTION notify_sos_alert()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER AS $$
BEGIN
  PERFORM net.http_post(
    url    := current_setting('app.supabase_url') || '/functions/v1/send-sos-notification',
    body   := json_build_object('record', row_to_json(NEW))::text,
    params := '{}',
    headers := json_build_object(
      'Content-Type', 'application/json',
      'Authorization', 'Bearer ' || current_setting('app.service_role_key')
    )::text
  );
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS on_sos_insert ON public.sos_alerts;
CREATE TRIGGER on_sos_insert
  AFTER INSERT ON public.sos_alerts
  FOR EACH ROW EXECUTE FUNCTION notify_sos_alert();
