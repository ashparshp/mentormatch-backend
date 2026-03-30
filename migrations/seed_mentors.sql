-- Seed data for Mentors
-- Note: In a production Supabase app, these would be linked to auth.users.
-- For local development/demo, we insert into public.profiles directly.

-- Clear existing data if needed (optional)
-- DELETE FROM public.mentor_profiles;
-- DELETE FROM public.profiles WHERE role = 'mentor';

-- 1. Arpit Verma
INSERT INTO public.profiles (id, email, full_name, avatar_url, profile_headline, organization, skill_tags, role, has_completed_onboarding)
VALUES (
  '00000000-0000-0000-0000-000000000001', 
  'arpit@example.com', 
  'Dr. Arpit Verma', 
  'https://api.dicebear.com/7.x/avataaars/svg?seed=Arpit', 
  'IIT Delhi Alum | Senior SDE @ Microsoft | JEE Physics Expert', 
  'Microsoft', 
  '["Go", "System Design", "Cloud"]'::jsonb, 
  'mentor', 
  true
) ON CONFLICT (id) DO UPDATE SET 
  full_name = EXCLUDED.full_name, 
  avatar_url = EXCLUDED.avatar_url, 
  profile_headline = EXCLUDED.profile_headline, 
  organization = EXCLUDED.organization, 
  skill_tags = EXCLUDED.skill_tags, 
  role = EXCLUDED.role, 
  has_completed_onboarding = EXCLUDED.has_completed_onboarding;

INSERT INTO public.mentor_profiles (profile_id, status, bio, expertise_tags, hourly_rate, availability_slots)
VALUES (
  '00000000-0000-0000-0000-000000000001', 
  'active', 
  'Senior Software Engineer with 8+ years of experience in building scalable cloud systems. Passionate about teaching Physics and mentoring students for competitive exams like JEE.', 
  '["JEE Physics", "System Design", "Cloud Architecture"]'::jsonb, 
  2500.00, 
  '{"monday": ["10:00-11:00", "14:00-15:00"], "wednesday": ["10:00-11:00"]}'::jsonb
) ON CONFLICT (profile_id) DO UPDATE SET 
  status = EXCLUDED.status, 
  bio = EXCLUDED.bio, 
  expertise_tags = EXCLUDED.expertise_tags, 
  hourly_rate = EXCLUDED.hourly_rate, 
  availability_slots = EXCLUDED.availability_slots;

-- 2. Sneha Kapur
INSERT INTO public.profiles (id, email, full_name, avatar_url, profile_headline, organization, skill_tags, role, has_completed_onboarding)
VALUES (
  '00000000-0000-0000-0000-000000000002', 
  'sneha@example.com', 
  'Sneha Kapur', 
  'https://api.dicebear.com/7.x/avataaars/svg?seed=Sneha', 
  'GSoC ''22 Mentor | Open Source Advocate | React Specialist', 
  'Meta', 
  '["React", "TypeScript", "Next.js"]'::jsonb, 
  'mentor', 
  true
) ON CONFLICT (id) DO UPDATE SET 
  full_name = EXCLUDED.full_name, 
  avatar_url = EXCLUDED.avatar_url, 
  profile_headline = EXCLUDED.profile_headline, 
  organization = EXCLUDED.organization, 
  skill_tags = EXCLUDED.skill_tags, 
  role = EXCLUDED.role, 
  has_completed_onboarding = EXCLUDED.has_completed_onboarding;

INSERT INTO public.mentor_profiles (profile_id, status, bio, expertise_tags, hourly_rate, availability_slots)
VALUES (
  '00000000-0000-0000-0000-000000000002', 
  'active', 
  'Frontend specialist and Open Source enthusiast. I''ve helped multiple students crack GSoC and land roles at top tech companies through focused React mentorship.', 
  '["React", "GSoC Prep", "Next.js", "OSS"]'::jsonb, 
  1800.00, 
  '{"tuesday": ["16:00-17:00"], "thursday": ["16:00-17:00"]}'::jsonb
) ON CONFLICT (profile_id) DO UPDATE SET 
  status = EXCLUDED.status, 
  bio = EXCLUDED.bio, 
  expertise_tags = EXCLUDED.expertise_tags, 
  hourly_rate = EXCLUDED.hourly_rate, 
  availability_slots = EXCLUDED.availability_slots;

-- 3. Vikram Singh
INSERT INTO public.profiles (id, email, full_name, avatar_url, profile_headline, organization, skill_tags, role, has_completed_onboarding)
VALUES (
  '00000000-0000-0000-0000-000000000003', 
  'vikram@example.com', 
  'Vikram Singh', 
  'https://api.dicebear.com/7.x/avataaars/svg?seed=Vikram', 
  'IIT Bombay | Competitive Programmer | FAANG Interview Coach', 
  'Google', 
  '["DSA", "Algorithms", "C++"]'::jsonb, 
  'mentor', 
  true
) ON CONFLICT (id) DO UPDATE SET 
  full_name = EXCLUDED.full_name, 
  avatar_url = EXCLUDED.avatar_url, 
  profile_headline = EXCLUDED.profile_headline, 
  organization = EXCLUDED.organization, 
  skill_tags = EXCLUDED.skill_tags, 
  role = EXCLUDED.role, 
  has_completed_onboarding = EXCLUDED.has_completed_onboarding;

INSERT INTO public.mentor_profiles (profile_id, status, bio, expertise_tags, hourly_rate, availability_slots)
VALUES (
  '00000000-0000-0000-0000-000000000003', 
  'active', 
  'Competitive programmer and SDE at Google. I specialize in teaching Algorithms and Data Structures through first principles.', 
  '["DSA", "CP", "Mock Interviews", "Algorithms"]'::jsonb, 
  2200.00, 
  '{"friday": ["18:00-20:00"], "saturday": ["10:00-12:00"]}'::jsonb
) ON CONFLICT (profile_id) DO UPDATE SET 
  status = EXCLUDED.status, 
  bio = EXCLUDED.bio, 
  expertise_tags = EXCLUDED.expertise_tags, 
  hourly_rate = EXCLUDED.hourly_rate, 
  availability_slots = EXCLUDED.availability_slots;
