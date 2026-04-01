-- Migration: Create Search Logs for Discovery Analytics Engine
-- This table tracks every query on the platform to identify Supply Gaps (zero results)
-- and generate real-time autocomplete suggestions.

CREATE TABLE IF NOT EXISTS public.search_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query TEXT NOT NULL,
    results_count INTEGER NOT NULL DEFAULT 0,
    user_id UUID REFERENCES public.profiles(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL
);

-- Index for fast aggregation on exact search terms (for autocomplete & top searches)
CREATE INDEX IF NOT EXISTS idx_search_logs_query ON public.search_logs(query);

-- Index for time-series filtering (e.g., "Trending in last 24h")
CREATE INDEX IF NOT EXISTS idx_search_logs_created_at ON public.search_logs(created_at);

-- Composite index to rapidly answer "What are the most recent zero-result queries?"
CREATE INDEX IF NOT EXISTS idx_search_logs_zero_results ON public.search_logs(results_count, created_at) WHERE results_count = 0;
