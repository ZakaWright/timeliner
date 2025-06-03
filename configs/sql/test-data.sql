INSERT INTO users (username, password_hash) VALUES
    ('admin', '$2a$10$/InmzcL8e7.6HM7/y08QVu9MQTKVNlO7NBaxP4im3coV6I6tMWe/K'),
    ('test_user', '$2a$10$/InmzcL8e7.6HM7/y08QVu9MQTKVNlO7NBaxP4im3coV6I6tMWe/K');

-- Create test instances
INSERT INTO incidents (name, description, case_number, created_by) VALUES
    ('Basic Incident', 'Basic Incident for testing', 'case1', (SELECT user_id FROM users WHERE username='test_user')),
    ('Lateral Movement Incident', 'Incident for testing mapping of lateral movement', 'case2', (SELECT user_id FROM users WHERE username='test_user'));

INSERT INTO endpoints (device_name, incident_id, os, ip_address) VALUES
    ('endpoint 1', (SELECT incident_id FROM incidents WHERE name='Basic Incident'), 'Windows 11', '192.168.1.1'),
    ('endpoint 2', (SELECT incident_id FROM incidents WHERE name='Basic Incident'), 'Windows 11', '192.168.1.2'),
    ('initial victim', (SELECT incident_id FROM incidents WHERE name='Lateral Movement Incident'), 'Windows 11', '192.168.2.1'),
    ('second victim', (SELECT incident_id FROM incidents WHERE name='Lateral Movement Incident'), 'Windows 11', '192.168.2.2');

INSERT INTO events (incident_id, event_time, description, created_by, endpoint_id, mitre_tactic) VALUES
    ((SELECT incident_id FROM incidents WHERE name='Lateral Movement Incident'), NOW(), 'lateral movement', 
    (SELECT user_id FROM users WHERE username='test_user'), (SELECT endpoint_id FROM endpoints WHERE device_name='initial victim'),
    (SELECT tactic_id FROM mitre_tactic WHERE name='Lateral Movement')),
    ((SELECT incident_id FROM incidents WHERE name='Basic Incident'), NOW(), 'Initial access', 
    (SELECT user_id FROM users WHERE username='test_user'), (SELECT endpoint_id FROM endpoints WHERE device_name='endpoint 1'),
    (SELECT tactic_id FROM mitre_tactic WHERE name='Initial Access'));