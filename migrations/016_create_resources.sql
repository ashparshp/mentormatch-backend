-- Create Resources table
CREATE TABLE IF NOT EXISTS public.resources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    category TEXT NOT NULL, -- roadmaps, templates, guides, interviews
    description TEXT,
    author_id UUID REFERENCES public.profiles(id),
    author_name TEXT, -- Fallback for external authors
    downloads INTEGER DEFAULT 0,
    rating DECIMAL(2,1) DEFAULT 0.0,
    file_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL
);

-- Enable RLS
ALTER TABLE public.resources ENABLE ROW LEVEL SECURITY;

-- Allow anyone to read resources
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = 'public'
          AND tablename = 'resources'
          AND policyname = 'Anyone can read resources'
    ) THEN
        CREATE POLICY "Anyone can read resources" ON public.resources
            FOR SELECT USING (true);
    END IF;
END $$;

-- Seed Resources based on MOCK_RESOURCES
INSERT INTO public.resources (title, category, description, author_name, rating, downloads)
VALUES 
('Ultimate JEE 2026 Roadmap', 'roadmaps', 'A month-by-month breakdown of subjects, mock tests, and revision strategies.', 'Dr. Arpit Verma', 4.9, 1200),
('GSoC Proposal Template', 'templates', 'The exact template used by 50+ successful students to crack GSoC.', 'Sneha Kapur', 5.0, 850),
('System Design Primer', 'guides', 'Core concepts of scalability, availability, and distributed systems for beginners.', 'Vikram Singh', 4.8, 2400),
('NEET-PG Strategy Guide', 'guides', 'How to balance clinical duties with high-yield subject revision.', 'Dr. Ishaan', 4.7, 600),
('PM Case Study Framework', 'templates', 'A structured approach to solving product design and estimation problems.', 'Priya Sharma', 4.9, 1100),
('DSA Final Revision Sheet', 'roadmaps', 'All essential algorithms and patterns for FAANG interviews in one place.', 'Vikram Singh', 5.0, 3800);
