SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;
SET default_tablespace = '';
SET default_table_access_method = heap;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" SCHEMA public; -- To use uuid_generate_v7 custom function.

-- Creating uuid generate v7 method - this feature is not implemented in postgres 17 yet.
-- Will be implemented in 18.
create or replace function public.uuid_generate_v7()
returns uuid
as $$
begin
  -- use random v4 uuid as starting point (which has the same variant we need)
  -- then overlay timestamp
  -- then set version 7 by flipping the 2 and 1 bit in the version 4 string
  return encode(
    set_bit(
      set_bit(
        overlay(uuid_send(gen_random_uuid())
                placing substring(int8send(floor(extract(epoch from clock_timestamp()) * 1000)::bigint) from 3)
                from 1 for 6
        ),
        52, 1
      ),
      53, 1
    ),
    'hex')::uuid;
end
$$
language plpgsql
volatile;

-- Messages table - stores the actual chat messages
CREATE TABLE IF NOT EXISTS public.message (
    id UUID PRIMARY KEY DEFAULT public.uuid_generate_v7(),
    chat_id UUID NOT NULL,
    sender_id UUID NOT NULL,
    content TEXT NOT NULL,
    message_type VARCHAR(20) NOT NULL DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'video', 'audio', 'file')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    read_status BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS public.chat (
    id uuid NOT NULL,
    type CHARACTER VARYING(20) NOT NULL CHECK (type IN ('private', 'group', 'container')),
    usergroup_id bigint,
    container_id uuid,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_chat PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.chat_participant (
    id UUID PRIMARY KEY DEFAULT public.uuid_generate_v7(),
    chat_id UUID NOT NULL REFERENCES public.chat(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    role VARCHAR(10) CHECK (role IN ('admin', 'member')) DEFAULT 'member',
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'left', 'banned', 'muted', 'pending'))
);

INSERT INTO public.chat(id, type, usergroup_id, container_id, created_at) VALUES ('01987073-0a87-7b32-9439-86868dfe9bd2', 'group', 1, NULL, CURRENT_TIMESTAMP);
INSERT INTO public.chat(id, type, usergroup_id, container_id, created_at) VALUES ('01987073-cf13-7621-af36-54ce20056d18', 'group', 2, NULL, CURRENT_TIMESTAMP);
INSERT INTO public.chat(id, type, usergroup_id, container_id, created_at) VALUES ('01987075-16cb-7337-af15-cd28f64c93a3', 'group', 3, NULL, CURRENT_TIMESTAMP);
INSERT INTO public.chat(id, type, usergroup_id, container_id, created_at) VALUES ('01987074-1f7f-7aad-ad76-a4b83544fa2d', 'group', 4, NULL, CURRENT_TIMESTAMP);
INSERT INTO public.chat(id, type, usergroup_id, container_id, created_at) VALUES ('01987074-440c-73f8-aa5b-ba2b50a19395', 'group', 5, NULL, CURRENT_TIMESTAMP);

-- Chat 1 (User Group 1): kevin, macy, testing1 are members (based on usergroup_user table)
INSERT INTO public.chat_participant(chat_id, user_id, role, status) VALUES
('01987073-0a87-7b32-9439-86868dfe9bd2', '01959b38-b3f9-7ec5-8ac8-e353bfe08a2d', 'admin',  'active'),
('01987073-0a87-7b32-9439-86868dfe9bd2', '01959b39-febd-770d-9e1b-e5ee392fce54', 'member', 'active'),
('01987073-0a87-7b32-9439-86868dfe9bd2', '01959b3a-405b-7591-86dd-87174e2453fd', 'member', 'active');

-- Chat 2 (User Group 2): testing2 is member
INSERT INTO public.chat_participant(chat_id, user_id, role, status) VALUES ('01987073-cf13-7621-af36-54ce20056d18', '0195c388-d0f4-77d5-be90-971d38344c74', 'admin', 'active');
-- Chat 3 (User Group 3): kevin is member
INSERT INTO public.chat_participant(chat_id, user_id, role, status) VALUES ('01987075-16cb-7337-af15-cd28f64c93a3', '01959b38-b3f9-7ec5-8ac8-e353bfe08a2d', 'admin', 'active');
-- Chat 4 (User Group 4): macy is member
INSERT INTO public.chat_participant(chat_id, user_id, role, status) VALUES ('01987074-1f7f-7aad-ad76-a4b83544fa2d', '01959b39-febd-770d-9e1b-e5ee392fce54', 'admin', 'active');
-- Chat 5 (User Group 5): kevin is member
INSERT INTO public.chat_participant(chat_id, user_id, role, status) VALUES ('01987074-440c-73f8-aa5b-ba2b50a19395', '01959b38-b3f9-7ec5-8ac8-e353bfe08a2d', 'admin', 'active');
