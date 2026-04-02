-- 1. Update mentor_profiles with stats columns
ALTER TABLE public.mentor_profiles 
ADD COLUMN IF NOT EXISTS average_rating DECIMAL(3,2) DEFAULT 0.00,
ADD COLUMN IF NOT EXISTS review_count INTEGER DEFAULT 0;

-- 2. Create reviews table
CREATE TABLE IF NOT EXISTS public.reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL REFERENCES public.bookings(id) UNIQUE,
    student_id UUID NOT NULL REFERENCES public.profiles(id),
    mentor_id UUID NOT NULL REFERENCES public.profiles(id),
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL
);

-- 3. Enable RLS
ALTER TABLE public.reviews ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = 'public'
          AND tablename = 'reviews'
          AND policyname = 'Anyone can read reviews'
    ) THEN
        CREATE POLICY "Anyone can read reviews" ON public.reviews
            FOR SELECT USING (true);
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = 'public'
          AND tablename = 'reviews'
          AND policyname = 'Students can insert reviews for their bookings'
    ) THEN
        CREATE POLICY "Students can insert reviews for their bookings" ON public.reviews
            FOR INSERT WITH CHECK (auth.uid() = student_id);
    END IF;
END $$;

-- 4. Create function to update mentor stats
CREATE OR REPLACE FUNCTION public.update_mentor_stats()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE public.mentor_profiles
    SET 
        average_rating = (
            SELECT COALESCE(AVG(rating), 0.00)
            FROM public.reviews
            WHERE mentor_id = NEW.mentor_id
        ),
        review_count = (
            SELECT COUNT(*)
            FROM public.reviews
            WHERE mentor_id = NEW.mentor_id
        )
    WHERE profile_id = NEW.mentor_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 5. Create trigger
DROP TRIGGER IF EXISTS on_review_added ON public.reviews;
CREATE TRIGGER on_review_added
    AFTER INSERT OR UPDATE OR DELETE ON public.reviews
    FOR EACH ROW EXECUTE FUNCTION public.update_mentor_stats();
