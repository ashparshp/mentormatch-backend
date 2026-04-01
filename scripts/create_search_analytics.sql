-- Discovery Analytics: Track student search patterns

CREATE TABLE IF NOT EXISTS public.search_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query TEXT NOT NULL,
    results_count INTEGER NOT NULL,
    user_id UUID REFERENCES public.profiles(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for analytics queries (ordering by time)
CREATE INDEX IF NOT EXISTS idx_search_logs_created_at ON public.search_logs (created_at DESC);

-- Index for term aggregation
CREATE INDEX IF NOT EXISTS idx_search_logs_query ON public.search_logs(query);

-- Vacuum analyze for new stats
ANALYZE public.search_logs;
