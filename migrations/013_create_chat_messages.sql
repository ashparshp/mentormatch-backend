-- Migration: Create chat_messages table
-- Purpose: Enable high-fidelity technical communication and protocol logging for mentorship sessions.

CREATE TABLE IF NOT EXISTS public.chat_messages (
    id UUID PRIMARY KEY,
    booking_id UUID NOT NULL REFERENCES public.bookings(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES public.profiles(id),
    content TEXT NOT NULL,
    is_protocol_item BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for optimized session history retrieval
CREATE INDEX IF NOT EXISTS idx_chat_messages_booking_id ON public.chat_messages(booking_id);
