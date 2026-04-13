-- Текущие статусы
CREATE TABLE IF NOT EXISTS statuses (
    user_id UUID PRIMARY KEY,
    text VARCHAR(140) NOT NULL,
    emoji VARCHAR(10),
    type VARCHAR(20) DEFAULT 'normal',
    activity VARCHAR(50),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- История статусов
CREATE TABLE IF NOT EXISTS status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    text VARCHAR(140) NOT NULL,
    emoji VARCHAR(10),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_statuses_expires_at ON statuses(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_status_history_user_id ON status_history(user_id);
CREATE INDEX IF NOT EXISTS idx_status_history_created_at ON status_history(created_at DESC);

-- Триггер для updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_statuses_updated_at ON statuses;

CREATE TRIGGER update_statuses_updated_at 
    BEFORE UPDATE ON statuses 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();