package com.devsecurity.core.repository;

import com.devsecurity.core.domain.ScanJob;
import com.devsecurity.core.domain.ScanStatus;
import java.time.Instant;
import java.util.List;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;

public interface ScanJobRepository extends JpaRepository<ScanJob, UUID> {
    List<ScanJob> findByProjectIdOrderByCreatedAtDesc(UUID projectId);

    @Query("select s.id from ScanJob s where s.status = :status order by s.createdAt asc")
    List<UUID> findIdsByStatus(ScanStatus status);

    @Modifying
    @Query(value = "update scan_job set status='RUNNING', started_at=now() " +
            "where id=:id and status='QUEUED'", nativeQuery = true)
    int claim(UUID id);

    List<ScanJob> findTop100BySourceDeletedAtIsNullAndSourceExpiresAtBeforeOrderBySourceExpiresAtAsc(Instant now);
    List<ScanJob> findTop100ByResultExpiresAtBeforeOrderByResultExpiresAtAsc(Instant now);
}
