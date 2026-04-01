-- Optimization: Add GIN indexes for fast categorical matching

-- 1. Index for Mentor Expertise
CREATE INDEX IF NOT EXISTS idx_mentor_profiles_expertise ON public.mentor_profiles USING GIN (expertise_tags);

-- 2. Index for Student Interests
CREATE INDEX IF NOT EXISTS idx_student_profiles_interests ON public.student_profiles USING GIN (interests);

-- 3. Vacuum Analyze for fresh stats
ANALYZE public.mentor_profiles;
ANALYZE public.student_profiles;
