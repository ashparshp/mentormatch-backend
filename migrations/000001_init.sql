-- 1. Core Data Enums
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM ('student', 'mentor', 'admin');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'mentor_status') THEN
        CREATE TYPE mentor_status AS ENUM ('pending', 'active', 'inactive', 'blocked');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'booking_status') THEN
        CREATE TYPE booking_status AS ENUM ('pending', 'accepted', 'rejected', 'completed', 'cancelled');
    END IF;
END $$;

-- 2. Shared Profiles Table
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'profiles') THEN
        CREATE TABLE public.profiles (
          id UUID PRIMARY KEY, -- Linked to auth.users(id)
          email TEXT UNIQUE NOT NULL,
          full_name TEXT,
          avatar_url TEXT,
          phone_number TEXT,
          profile_headline TEXT,
          organization TEXT,
          social_links JSONB DEFAULT '{}'::jsonb,
          skill_tags JSONB DEFAULT '[]'::jsonb,
          preferred_languages JSONB DEFAULT '[]'::jsonb,
          state TEXT,
          role user_role DEFAULT 'student',
          has_completed_onboarding BOOLEAN DEFAULT false,
          created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL,
          updated_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL
        );
    END IF;
END $$;

-- 3. Student Specific Data
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'student_profiles') THEN
        CREATE TABLE public.student_profiles (
          profile_id UUID PRIMARY KEY REFERENCES public.profiles(id) ON DELETE CASCADE,
          interests JSONB DEFAULT '[]'::jsonb,
          current_education TEXT,
          learning_goals TEXT
        );
    END IF;
END $$;

-- 4. Mentor Specific Data
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'mentor_profiles') THEN
        CREATE TABLE public.mentor_profiles (
          profile_id UUID PRIMARY KEY REFERENCES public.profiles(id) ON DELETE CASCADE,
          status mentor_status DEFAULT 'pending',
          bio TEXT,
          expertise_tags JSONB DEFAULT '[]'::jsonb,
          verification_url TEXT,
          hourly_rate NUMERIC(10, 2),
          availability_slots JSONB DEFAULT '{}'::jsonb,
          verified_at TIMESTAMP WITH TIME ZONE,
          last_online_at TIMESTAMP WITH TIME ZONE
        );
    END IF;
END $$;

-- 5. Bookings Table
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'bookings') THEN
        CREATE TABLE public.bookings (
          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
          student_id UUID REFERENCES public.profiles(id) NOT NULL,
          mentor_id UUID REFERENCES public.mentor_profiles(profile_id) NOT NULL,
          start_time TIMESTAMP WITH TIME ZONE NOT NULL,
          end_time TIMESTAMP WITH TIME ZONE NOT NULL,
          status booking_status DEFAULT 'pending' NOT NULL,
          total_price NUMERIC(10, 2) NOT NULL,
          meeting_link TEXT,
          created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL,
          updated_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL
        );
    END IF;
END $$;

-- 6. Tracking Changes (Audit Logs)
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'profile_audit_logs') THEN
        CREATE TABLE public.profile_audit_logs (
          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
          profile_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE,
          changed_by UUID REFERENCES public.profiles(id),
          old_status TEXT,
          new_status TEXT,
          reason TEXT,
          created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        );
    END IF;
END $$;

-- 7. RLS Policies
ALTER TABLE public.profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.mentor_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.bookings ENABLE ROW LEVEL SECURITY;

-- Visibility Policies
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE policyname = 'Public can view basic profiles') THEN
        CREATE POLICY "Public can view basic profiles" ON public.profiles FOR SELECT USING (true);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE policyname = 'Public can view active mentors') THEN
        CREATE POLICY "Public can view active mentors" ON public.mentor_profiles FOR SELECT USING (status = 'active');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE policyname = 'Users can update their basic info') THEN
        CREATE POLICY "Users can update their basic info" ON public.profiles FOR UPDATE USING (auth.uid() = id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE policyname = 'Students can view their own bookings') THEN
        CREATE POLICY "Students can view their own bookings" ON public.bookings FOR SELECT USING (auth.uid() = student_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE policyname = 'Mentors can view their own bookings') THEN
        CREATE POLICY "Mentors can view their own bookings" ON public.bookings FOR SELECT USING (auth.uid() = mentor_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE policyname = 'Students can create bookings') THEN
        CREATE POLICY "Students can create bookings" ON public.bookings FOR INSERT WITH CHECK (auth.uid() = student_id);
    END IF;
END $$;

-- 8. Automation: Triggers for Auth Integration
CREATE OR REPLACE FUNCTION public.handle_new_user() 
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO public.profiles (id, email, full_name, avatar_url)
  VALUES (new.id, new.email, new.raw_user_meta_data->>'full_name', new.raw_user_meta_data->>'avatar_url');
  RETURN new;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'on_auth_user_created') THEN
        CREATE TRIGGER on_auth_user_created
          AFTER INSERT ON auth.users
          FOR EACH ROW EXECUTE PROCEDURE public.handle_new_user();
    END IF;
END $$;

-- 9. Updated At Trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = timezone('utc'::text, now());
    RETURN NEW;
END;
$$ language 'plpgsql';

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_profiles_updated_at') THEN
        CREATE TRIGGER update_profiles_updated_at BEFORE UPDATE ON public.profiles FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_bookings_updated_at') THEN
        CREATE TRIGGER update_bookings_updated_at BEFORE UPDATE ON public.bookings FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
    END IF;
END $$;
