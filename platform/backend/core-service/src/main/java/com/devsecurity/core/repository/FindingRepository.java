package com.devsecurity.core.repository;

import com.devsecurity.core.domain.Finding;
import java.util.List;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;

public interface FindingRepository extends JpaRepository<Finding, UUID> {
    List<Finding> findByScanIdOrderBySeverityAscTitleAsc(UUID scanId);
    void deleteByScanId(UUID scanId);
}
