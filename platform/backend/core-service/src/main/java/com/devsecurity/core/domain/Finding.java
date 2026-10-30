package com.devsecurity.core.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import java.time.Instant;
import java.util.UUID;

@Entity
@Table(name = "scan_finding")
public class Finding {
    @Id
    private UUID id;
    @Column(nullable = false)
    private UUID scanId;
    @Column(nullable = false)
    private UUID componentId;
    @Column(nullable = false)
    private String vulnerabilityId;
    @Column(length = 4000)
    private String aliases;
    @Column(nullable = false, length = 1000)
    private String title;
    @Column(nullable = false, length = 32)
    private String severity;
    @Column(nullable = false)
    private String packageName;
    @Column(nullable = false)
    private String affectedVersion;
    private String affectedRange;
    @Column(length = 2000)
    private String fixedVersions;
    @Column(nullable = false, length = 4000)
    private String dependencyPath;
    @Column(nullable = false)
    private Instant createdAt;

    protected Finding() {
    }

    public Finding(UUID scanId, UUID componentId, String vulnerabilityId, String aliases,
                   String title, String severity, String packageName, String affectedVersion,
                   String affectedRange, String fixedVersions, String dependencyPath) {
        this.id = UUID.randomUUID();
        this.scanId = scanId;
        this.componentId = componentId;
        this.vulnerabilityId = vulnerabilityId;
        this.aliases = aliases;
        this.title = title;
        this.severity = severity;
        this.packageName = packageName;
        this.affectedVersion = affectedVersion;
        this.affectedRange = affectedRange;
        this.fixedVersions = fixedVersions;
        this.dependencyPath = dependencyPath;
        this.createdAt = Instant.now();
    }

    public UUID getId() {
        return id;
    }

    public UUID getScanId() {
        return scanId;
    }

    public UUID getComponentId() {
        return componentId;
    }

    public String getVulnerabilityId() {
        return vulnerabilityId;
    }

    public String getAliases() {
        return aliases;
    }

    public String getTitle() {
        return title;
    }

    public String getSeverity() {
        return severity;
    }

    public String getPackageName() {
        return packageName;
    }

    public String getAffectedVersion() {
        return affectedVersion;
    }

    public String getAffectedRange() {
        return affectedRange;
    }

    public String getFixedVersions() {
        return fixedVersions;
    }

    public String getDependencyPath() {
        return dependencyPath;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }
}
