package com.devsecurity.core.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.EnumType;
import jakarta.persistence.Enumerated;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import java.time.Instant;
import java.util.UUID;

@Entity
@Table(name = "scan_job")
public class ScanJob {
    @Id
    private UUID id;
    @Column(nullable = false, length = 64)
    private String tenantId;
    @Column(nullable = false)
    private UUID projectId;
    @Column(nullable = false)
    private String requestedBy;
    @Enumerated(EnumType.STRING)
    @Column(nullable = false, length = 32)
    private ScanStatus status;
    @Enumerated(EnumType.STRING)
    @Column(nullable = false, length = 32)
    private SourceType sourceType;
    @Column(length = 2000)
    private String sourceRef;
    @Column(length = 200)
    private String sourceRevision;
    private String sourceObjectKey;
    private String sourceMode;
    private String engineVersion;
    private String failureCode;
    @Column(length = 2000)
    private String failureMessage;
    private Integer componentCount;
    private Integer findingCount;
    @Column(nullable = false)
    private Instant sourceExpiresAt;
    private Instant sourceDeletedAt;
    @Column(nullable = false)
    private Instant resultExpiresAt;
    @Column(nullable = false)
    private Instant createdAt;
    private Instant startedAt;
    private Instant finishedAt;

    protected ScanJob() {
    }

    public ScanJob(UUID id, String tenantId, UUID projectId, String requestedBy,
                   SourceType sourceType, String sourceRef, String sourceRevision,
                   Instant sourceExpiresAt, Instant resultExpiresAt) {
        this.id = id;
        this.tenantId = tenantId;
        this.projectId = projectId;
        this.requestedBy = requestedBy;
        this.status = ScanStatus.QUEUED;
        this.sourceType = sourceType;
        this.sourceRef = sourceRef;
        this.sourceRevision = sourceRevision;
        this.sourceExpiresAt = sourceExpiresAt;
        this.resultExpiresAt = resultExpiresAt;
        this.createdAt = Instant.now();
    }

    public UUID getId() {
        return id;
    }

    public String getTenantId() {
        return tenantId;
    }

    public UUID getProjectId() {
        return projectId;
    }

    public String getRequestedBy() {
        return requestedBy;
    }

    public ScanStatus getStatus() {
        return status;
    }

    public SourceType getSourceType() {
        return sourceType;
    }

    public String getSourceRef() {
        return sourceRef;
    }

    public String getSourceRevision() {
        return sourceRevision;
    }

    public String getSourceObjectKey() {
        return sourceObjectKey;
    }

    public void setSourceObjectKey(String sourceObjectKey) {
        this.sourceObjectKey = sourceObjectKey;
    }

    public String getSourceMode() {
        return sourceMode;
    }

    public void setSourceMode(String sourceMode) {
        this.sourceMode = sourceMode;
    }

    public String getEngineVersion() {
        return engineVersion;
    }

    public void setEngineVersion(String engineVersion) {
        this.engineVersion = engineVersion;
    }

    public String getFailureCode() {
        return failureCode;
    }

    public void setFailureCode(String failureCode) {
        this.failureCode = failureCode;
    }

    public String getFailureMessage() {
        return failureMessage;
    }

    public void setFailureMessage(String failureMessage) {
        this.failureMessage = failureMessage;
    }

    public Integer getComponentCount() {
        return componentCount;
    }

    public void setComponentCount(Integer componentCount) {
        this.componentCount = componentCount;
    }

    public Integer getFindingCount() {
        return findingCount;
    }

    public void setFindingCount(Integer findingCount) {
        this.findingCount = findingCount;
    }

    public Instant getSourceExpiresAt() {
        return sourceExpiresAt;
    }

    public Instant getSourceDeletedAt() {
        return sourceDeletedAt;
    }

    public void setSourceDeletedAt(Instant sourceDeletedAt) {
        this.sourceDeletedAt = sourceDeletedAt;
    }

    public Instant getResultExpiresAt() {
        return resultExpiresAt;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }

    public Instant getStartedAt() {
        return startedAt;
    }

    public Instant getFinishedAt() {
        return finishedAt;
    }

    public void complete(int components, int findings, String mode, String engineVersion) {
        this.status = ScanStatus.SUCCEEDED;
        this.componentCount = components;
        this.findingCount = findings;
        this.sourceMode = mode;
        this.engineVersion = engineVersion;
        this.finishedAt = Instant.now();
        this.failureCode = null;
        this.failureMessage = null;
    }

    public void fail(String code, String message) {
        this.status = ScanStatus.FAILED;
        this.failureCode = code;
        this.failureMessage = message;
        this.finishedAt = Instant.now();
    }

    public void cancel() {
        this.status = ScanStatus.CANCELED;
        this.finishedAt = Instant.now();
    }
}
