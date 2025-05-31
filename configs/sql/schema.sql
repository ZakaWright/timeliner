-- Users table
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,  -- Store hashed passwords only
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Incidents table
CREATE TABLE incidents (
    incident_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    case_number VARCHAR(100),
    status VARCHAR(50) DEFAULT 'open',  -- open, closed, etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by INTEGER REFERENCES users(user_id),
    closed_at TIMESTAMP WITH TIME ZONE
);

-- Endpoints table
CREATE TABLE endpoints (
    endpoint_id SERIAL PRIMARY KEY,
    device_name VARCHAR(255) NOT NULL,
    incident_id INTEGER REFERENCES incidents(incident_id) ON DELETE CASCADE,
    os VARCHAR(100),
    os_version VARCHAR(100),
    ip_address VARCHAR(45),
    mac_address VARCHAR(17),
    last_seen TIMESTAMP WITH TIME ZONE,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- IOC types reference table
CREATE TABLE ioc_types (
    ioc_type_id SERIAL PRIMARY KEY,
    type_name VARCHAR(50) NOT NULL UNIQUE,  -- ip, domain, file_hash, process, etc.
    description TEXT
);

-- IOCs table - unified approach
CREATE TABLE iocs (
    ioc_id SERIAL PRIMARY KEY,
    ioc_type_id INTEGER REFERENCES ioc_types(ioc_type_id),
    value TEXT NOT NULL,  -- The primary IOC value (hash, IP, etc.)
    added_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    added_by INTEGER REFERENCES users(user_id),
    is_malicious BOOLEAN,
    UNIQUE (ioc_type_id, value)  -- Prevent duplicate IOCs of same type
);

-- IOC attributes table - for flexible attribute storage
CREATE TABLE ioc_attributes (
    attribute_id SERIAL PRIMARY KEY,
    ioc_id INTEGER REFERENCES iocs(ioc_id) ON DELETE CASCADE,
    attribute_name VARCHAR(100) NOT NULL,
    attribute_value TEXT,
    UNIQUE (ioc_id, attribute_name)  -- Each IOC can have each attribute only once
);

-- Timeline events table
CREATE TABLE events (
    event_id SERIAL PRIMARY KEY,
    incident_id INTEGER REFERENCES incidents(incident_id) ON DELETE CASCADE,
    event_time TIMESTAMP WITH TIME ZONE NOT NULL,
    event_type VARCHAR(100) NOT NULL,  -- detection, analysis, containment, etc.
    description TEXT,
    created_by INTEGER REFERENCES users(user_id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    endpoint_id INTEGER REFERENCES endpoints(endpoint_id)
);

-- Event-IOC relationship table
CREATE TABLE event_iocs (
    event_id INTEGER REFERENCES events(event_id) ON DELETE CASCADE,
    ioc_id INTEGER REFERENCES iocs(ioc_id) ON DELETE CASCADE,
    PRIMARY KEY (event_id, ioc_id)
);


-- Collaborative edits/comments
CREATE TABLE event_comments (
    comment_id SERIAL PRIMARY KEY,
    event_id INTEGER REFERENCES events(event_id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(user_id),
    comment TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

INSERT INTO ioc_types (type_name, description) VALUES
    ('ip_address', 'IPv4 or IPv6 address'),
    ('domain', 'Domain name'),
    ('file_name', 'File name or path'),
    ('email', 'Email address'),
    ('process', 'Process name or path'),
    ('registry', 'Windows registry key or value');