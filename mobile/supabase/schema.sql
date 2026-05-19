-- ============================================================
-- RideUp – Supabase schema
-- Auth gérée par Supabase Auth (table auth.users)
-- ============================================================

-- Profils (extension de auth.users)
CREATE TABLE public.profiles (
  id          UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
  username    TEXT UNIQUE NOT NULL,
  profile_pic TEXT DEFAULT '/static/noprofilpic.png',
  latitude    DOUBLE PRECISION DEFAULT 49.43839,
  longitude   DOUBLE PRECISION DEFAULT 1.10160,
  address     TEXT DEFAULT '22 Place Saint-Marc, 76000 Rouen',
  preference  INTEGER DEFAULT 50,
  role        TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
  created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Activer RLS
ALTER TABLE public.profiles ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Lecture publique des profils"
  ON public.profiles FOR SELECT USING (true);

CREATE POLICY "Mise à jour de son propre profil"
  ON public.profiles FOR UPDATE USING (auth.uid() = id);

-- Trigger : crée automatiquement un profil à l'inscription
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER LANGUAGE plpgsql SECURITY DEFINER AS $$
BEGIN
  INSERT INTO public.profiles (id, username)
  VALUES (NEW.id, COALESCE(NEW.raw_user_meta_data->>'username', split_part(NEW.email, '@', 1)));
  RETURN NEW;
END;
$$;

CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

-- ============================================================
-- Events
-- ============================================================
CREATE TABLE public.events (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title          TEXT NOT NULL,
  description    TEXT,
  created_by     UUID NOT NULL REFERENCES public.profiles(id) ON DELETE CASCADE,
  latitude       DOUBLE PRECISION NOT NULL,
  longitude      DOUBLE PRECISION NOT NULL,
  address        TEXT,
  start_datetime TIMESTAMPTZ NOT NULL,
  end_datetime   TIMESTAMPTZ,
  participants   INTEGER DEFAULT 0,
  created_at     TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE public.events ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Lecture publique des events"
  ON public.events FOR SELECT USING (true);

CREATE POLICY "Créer un event si connecté"
  ON public.events FOR INSERT WITH CHECK (auth.uid() = created_by);

CREATE POLICY "Supprimer son propre event"
  ON public.events FOR DELETE USING (auth.uid() = created_by);

-- ============================================================
-- Participants aux events
-- ============================================================
CREATE TABLE public.event_participants (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id   UUID NOT NULL REFERENCES public.events(id) ON DELETE CASCADE,
  user_id    UUID NOT NULL REFERENCES public.profiles(id) ON DELETE CASCADE,
  joined_at  TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE (event_id, user_id)
);

ALTER TABLE public.event_participants ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Lecture publique des participants"
  ON public.event_participants FOR SELECT USING (true);

CREATE POLICY "Rejoindre un event si connecté"
  ON public.event_participants FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Quitter un event"
  ON public.event_participants FOR DELETE USING (auth.uid() = user_id);

-- Trigger : met à jour le compteur participants
CREATE OR REPLACE FUNCTION public.update_participants_count()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    UPDATE public.events SET participants = participants + 1 WHERE id = NEW.event_id;
  ELSIF TG_OP = 'DELETE' THEN
    UPDATE public.events SET participants = GREATEST(participants - 1, 0) WHERE id = OLD.event_id;
  END IF;
  RETURN NULL;
END;
$$;

CREATE TRIGGER on_participant_change
  AFTER INSERT OR DELETE ON public.event_participants
  FOR EACH ROW EXECUTE FUNCTION public.update_participants_count();

-- ============================================================
-- Messages (par event)
-- ============================================================
CREATE TABLE public.messages (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id   UUID NOT NULL REFERENCES public.events(id) ON DELETE CASCADE,
  user_id    UUID NOT NULL REFERENCES public.profiles(id) ON DELETE CASCADE,
  content    TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE public.messages ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Lire les messages si participant"
  ON public.messages FOR SELECT
  USING (
    EXISTS (
      SELECT 1 FROM public.event_participants ep
      WHERE ep.event_id = messages.event_id AND ep.user_id = auth.uid()
    )
    OR EXISTS (
      SELECT 1 FROM public.events e WHERE e.id = messages.event_id AND e.created_by = auth.uid()
    )
  );

CREATE POLICY "Envoyer un message si participant"
  ON public.messages FOR INSERT
  WITH CHECK (
    auth.uid() = user_id AND (
      EXISTS (
        SELECT 1 FROM public.event_participants ep
        WHERE ep.event_id = messages.event_id AND ep.user_id = auth.uid()
      )
      OR EXISTS (
        SELECT 1 FROM public.events e WHERE e.id = messages.event_id AND e.created_by = auth.uid()
      )
    )
  );

-- ============================================================
-- Posts (feed photos)
-- ============================================================
CREATE TABLE public.posts (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES public.profiles(id) ON DELETE CASCADE,
  event_id   UUID REFERENCES public.events(id) ON DELETE SET NULL,
  image_url  TEXT NOT NULL,
  caption    TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE public.posts ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Lecture publique des posts"
  ON public.posts FOR SELECT USING (true);

CREATE POLICY "Créer un post si connecté"
  ON public.posts FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Supprimer son propre post"
  ON public.posts FOR DELETE USING (auth.uid() = user_id);

-- ============================================================
-- Storage – bucket avatars
-- ============================================================

CREATE POLICY "Upload avatar si connecté"
  ON storage.objects FOR INSERT
  TO authenticated
  WITH CHECK (bucket_id = 'avatars');

CREATE POLICY "Mettre à jour son avatar"
  ON storage.objects FOR UPDATE
  TO authenticated
  USING (bucket_id = 'avatars' AND auth.uid()::text = (storage.foldername(name))[1]);

CREATE POLICY "Lecture publique avatars"
  ON storage.objects FOR SELECT
  TO public
  USING (bucket_id = 'avatars');

-- ============================================================
-- Notifications
-- ============================================================
CREATE TABLE public.notifications (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES public.profiles(id) ON DELETE CASCADE,
  message    TEXT NOT NULL,
  is_read    BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE public.notifications ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Voir ses propres notifications"
  ON public.notifications FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Marquer comme lu"
  ON public.notifications FOR UPDATE USING (auth.uid() = user_id);
