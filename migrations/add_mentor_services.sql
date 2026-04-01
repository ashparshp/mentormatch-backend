-- Add services column to mentor_profiles to store custom offerings
ALTER TABLE public.mentor_profiles 
ADD COLUMN IF NOT EXISTS services JSONB DEFAULT '[]'::jsonb;

-- Example of services JSON structure:
-- [
--   {"id": "1", "title": "Resume Review", "duration": "30 min", "price": 1500, "type": "Career"},
--   {"id": "2", "title": "Mock Interview", "duration": "60 min", "price": 3000, "type": "Interview"}
-- ]
