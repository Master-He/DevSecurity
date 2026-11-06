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
    @Column
    private UUID componentId;
    @Column(nullable = false, length = 32)
    private String findingType;
    @Column
    private String vulnerabilityId;
    @Column(length = 255)
    private String ruleId;
    @Column(length = 4000)
    private String aliases;
    @Column(nullable = false, length = 1000)
    private String title;
    @Column(nullable = false, length = 32)
    private String severity;
    @Column
    private String packageName;
    @Column
    private String affectedVersion;
    private String affectedRange;
    @Column(length = 2000)
    private String fixedVersions;
    @Column(length = 4000)
    private String dependencyPath;
    @Column(length = 32)
    private String confidence;
    @Column(length = 128)
    private String category;
    @Column(length = 64)
    private String cwe;
    @Column(length = 2000)
    private String filePath;
    private Integer startLine;
    private Integer endLine;
    @Column(length = 4000)
    private String codeSnippet;
    @Column(length = 2000)
    private String recommendation;
    @Column(nullable = false)
    private Instant createdAt;

    protected Finding() {
    }

    public Finding(UUID scanId, UUID componentId, String vulnerabilityId, String aliases,
                   String title, String severity, String packageName, String affectedVersion,
                   String affectedRange, String fixedVersions, String dependencyPath) {
        this.id = UUID.randomUUID();
        this.scanId = scanId;
        this.findingType = ScanType.SCA.name();
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

    public static Finding sast(UUID scanId, String ruleId, String title, String severity,
                               String confidence, String category, String cwe,
                               String filePath, Integer startLine, Integer endLine,
                               String codeSnippet, String recommendation) {
        Finding finding = new Finding();
        finding.id = UUID.randomUUID();
        finding.scanId = scanId;
        finding.findingType = ScanType.SAST.name();
        finding.ruleId = ruleId;
        finding.title = title;
        finding.severity = severity;
        finding.confidence = confidence;
        finding.category = category;
        finding.cwe = cwe;
        finding.filePath = filePath;
        finding.startLine = startLine;
        finding.endLine = endLine;
        finding.codeSnippet = codeSnippet;
        finding.recommendation = recommendation;
        finding.createdAt = Instant.now();
        return finding;
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

    public String getFindingType() {
        return findingType;
    }

    public String getVulnerabilityId() {
        return vulnerabilityId;
    }

    public String getRuleId() {
        return ruleId;
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

    public String getConfidence() {
        return confidence;
    }

    public String getCategory() {
        return category;
    }

    public String getCwe() {
        return cwe;
    }

    public String getFilePath() {
        return filePath;
    }

    public Integer getStartLine() {
        return startLine;
    }

    public Integer getEndLine() {
        return endLine;
    }

    public String getCodeSnippet() {
        return codeSnippet;
    }

    public String getRecommendation() {
        return recommendation;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }
}
