package com.devsecurity.core.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import java.util.UUID;

@Entity
@Table(name = "scan_component")
public class Component {
    @Id
    private UUID id;
    @Column(nullable = false)
    private UUID scanId;
    @Column(nullable = false)
    private String ecosystem;
    @Column(nullable = false)
    private String groupId;
    @Column(nullable = false)
    private String artifactId;
    @Column(nullable = false)
    private String version;
    @Column(nullable = false)
    private String purl;
    @Column(nullable = false)
    private boolean direct;
    @Column(nullable = false, length = 4000)
    private String dependencyPath;
    private String scope;

    protected Component() {
    }

    public Component(UUID scanId, String ecosystem, String groupId, String artifactId,
                     String version, String purl, boolean direct, String dependencyPath, String scope) {
        this.id = UUID.randomUUID();
        this.scanId = scanId;
        this.ecosystem = ecosystem;
        this.groupId = groupId;
        this.artifactId = artifactId;
        this.version = version;
        this.purl = purl;
        this.direct = direct;
        this.dependencyPath = dependencyPath;
        this.scope = scope;
    }

    public UUID getId() {
        return id;
    }

    public UUID getScanId() {
        return scanId;
    }

    public String getEcosystem() {
        return ecosystem;
    }

    public String getGroupId() {
        return groupId;
    }

    public String getArtifactId() {
        return artifactId;
    }

    public String getVersion() {
        return version;
    }

    public String getPurl() {
        return purl;
    }

    public boolean isDirect() {
        return direct;
    }

    public String getDependencyPath() {
        return dependencyPath;
    }

    public String getScope() {
        return scope;
    }
}
