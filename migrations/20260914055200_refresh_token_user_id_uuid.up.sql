ALTER TABLE refresh_tokens 
ALTER COLUMN user_id TYPE UUID USING user_id::uuid;