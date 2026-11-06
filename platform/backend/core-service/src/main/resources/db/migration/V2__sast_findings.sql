ALTER TABLE scan_job
    ADD COLUMN scan_type VARCHAR(32) NOT NULL DEFAULT 'SCA';

ALTER TABLE scan_finding
    ADD COLUMN finding_type VARCHAR(32) NOT NULL DEFAULT 'SCA',
    ADD COLUMN rule_id VARCHAR(255),
    ADD COLUMN confidence VARCHAR(32),
    ADD COLUMN category VARCHAR(128),
    ADD COLUMN cwe VARCHAR(64),
    ADD COLUMN file_path VARCHAR(2000),
    ADD COLUMN start_line INTEGER,
    ADD COLUMN end_line INTEGER,
    ADD COLUMN code_snippet VARCHAR(4000),
    ADD COLUMN recommendation VARCHAR(2000);

ALTER TABLE scan_finding ALTER COLUMN component_id DROP NOT NULL;
ALTER TABLE scan_finding ALTER COLUMN vulnerability_id DROP NOT NULL;
ALTER TABLE scan_finding ALTER COLUMN package_name DROP NOT NULL;
ALTER TABLE scan_finding ALTER COLUMN affected_version DROP NOT NULL;
ALTER TABLE scan_finding ALTER COLUMN dependency_path DROP NOT NULL;

CREATE INDEX idx_scan_finding_type ON scan_finding(scan_id, finding_type, severity);
