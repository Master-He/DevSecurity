package com.devsecurity.core.repository;

import com.devsecurity.core.domain.Component;
import java.util.List;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;

public interface ComponentRepository extends JpaRepository<Component, UUID> {
    List<Component> findByScanIdOrderByGroupIdAscArtifactIdAsc(UUID scanId);
    void deleteByScanId(UUID scanId);
}
