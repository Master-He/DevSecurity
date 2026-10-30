package com.devsecurity.core.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import java.time.Instant;
import java.util.UUID;

@Entity
@Table(name = "project")
public class Project {
    @Id
    private UUID id;
    @Column(nullable = false, length = 64)
    private String tenantId;
    @Column(nullable = false)
    private String name;
    @Column(nullable = false, length = 1000)
    private String description;
    @Column(nullable = false)
    private UUID ownerId;
    @Column(nullable = false)
    private String ownerUsername;
    @Column(nullable = false)
    private Instant createdAt;

    protected Project() {
    }

    public Project(UUID id, String tenantId, String name, String description, UUID ownerId, String ownerUsername) {
        this.id = id;
        this.tenantId = tenantId;
        this.name = name;
        this.description = description;
        this.ownerId = ownerId;
        this.ownerUsername = ownerUsername;
        this.createdAt = Instant.now();
    }

    public UUID getId() {
        return id;
    }

    public String getTenantId() {
        return tenantId;
    }

    public String getName() {
        return name;
    }

    public String getDescription() {
        return description;
    }

    public UUID getOwnerId() {
        return ownerId;
    }

    public String getOwnerUsername() {
        return ownerUsername;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }
}
