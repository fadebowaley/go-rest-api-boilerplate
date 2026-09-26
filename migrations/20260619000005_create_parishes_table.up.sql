CREATE TABLE IF NOT EXISTS parishes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE,
    address TEXT,
    area VARCHAR(100),
    zone VARCHAR(100),
    province VARCHAR(100),
    region VARCHAR(100),
    pastor_name VARCHAR(255),
    pastor_phone VARCHAR(50),
    pastor_email VARCHAR(255),
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_parishes_code ON parishes(code);
CREATE INDEX IF NOT EXISTS idx_parishes_region ON parishes(region);
CREATE INDEX IF NOT EXISTS idx_parishes_province ON parishes(province);
CREATE INDEX IF NOT EXISTS idx_parishes_zone ON parishes(zone);
CREATE INDEX IF NOT EXISTS idx_parishes_area ON parishes(area);
CREATE INDEX IF NOT EXISTS idx_parishes_status ON parishes(status);
CREATE INDEX IF NOT EXISTS idx_parishes_deleted_at ON parishes(deleted_at);
