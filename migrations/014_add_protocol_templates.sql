-- Migration: Add protocol_templates column to mentor_profiles
-- Purpose: Enable mentors to define reusable session checklists and resource streams.

ALTER TABLE public.mentor_profiles
ADD COLUMN IF NOT EXISTS protocol_templates JSONB DEFAULT '[]'::jsonb;

COMMENT ON COLUMN public.mentor_profiles.protocol_templates IS 
  'Stores an array of ProtocolTemplate objects (id, title, description, checklist, resources) as JSONB';
