CREATE TABLE app_user (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    role VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE project (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description VARCHAR(1000) NOT NULL,
    owner_id UUID NOT NULL REFERENCES app_user(id),
    owner_username VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE scan_job (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    project_id UUID NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    requested_by VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL,
    source_type VARCHAR(32) NOT NULL,
    source_ref VARCHAR(2000),
    source_revision VARCHAR(200),
    source_object_key VARCHAR(255),
    source_mode VARCHAR(255),
    engine_version VARCHAR(255),
    failure_code VARCHAR(255),
    failure_message VARCHAR(2000),
    component_count INTEGER,
    finding_count INTEGER,
    source_expires_at TIMESTAMPTZ NOT NULL,
    source_deleted_at TIMESTAMPTZ,
    result_expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ
);

CREATE INDEX idx_scan_job_queue ON scan_job(status, created_at);
CREATE INDEX idx_scan_job_project ON scan_job(project_id, created_at DESC);
CREATE INDEX idx_scan_job_source_retention ON scan_job(source_expires_at)
    WHERE source_deleted_at IS NULL;
CREATE INDEX idx_scan_job_result_retention ON scan_job(result_expires_at);

CREATE TABLE scan_component (
    id UUID PRIMARY KEY,
    scan_id UUID NOT NULL REFERENCES scan_job(id) ON DELETE CASCADE,
    ecosystem VARCHAR(255) NOT NULL,
    group_id VARCHAR(255) NOT NULL,
    artifact_id VARCHAR(255) NOT NULL,
    version VARCHAR(255) NOT NULL,
    purl VARCHAR(255) NOT NULL,
    direct BOOLEAN NOT NULL,
    dependency_path VARCHAR(4000) NOT NULL,
    scope VARCHAR(255)
);

CREATE INDEX idx_scan_component_scan ON scan_component(scan_id);

CREATE TABLE scan_finding (
    id UUID PRIMARY KEY,
    scan_id UUID NOT NULL REFERENCES scan_job(id) ON DELETE CASCADE,
    component_id UUID NOT NULL REFERENCES scan_component(id) ON DELETE CASCADE,
    vulnerability_id VARCHAR(255) NOT NULL,
    aliases VARCHAR(4000),
    title VARCHAR(1000) NOT NULL,
    severity VARCHAR(32) NOT NULL,
    package_name VARCHAR(255) NOT NULL,
    affected_version VARCHAR(255) NOT NULL,
    affected_range VARCHAR(255),
    fixed_versions VARCHAR(2000),
    dependency_path VARCHAR(4000) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_scan_finding_scan ON scan_finding(scan_id, severity);
CREATE INDEX idx_scan_finding_vulnerability ON scan_finding(vulnerability_id);

CREATE TABLE audit_event (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    actor VARCHAR(255) NOT NULL,
    action VARCHAR(128) NOT NULL,
    resource_type VARCHAR(128) NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_audit_event_created ON audit_event(created_at DESC);
