package com.devsecurity.core.service;

import com.devsecurity.core.repository.ComponentRepository;
import com.devsecurity.core.repository.FindingRepository;
import com.devsecurity.core.repository.ScanJobRepository;
import java.time.Instant;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class RetentionService {
    private static final Logger log = LoggerFactory.getLogger(RetentionService.class);

    private final ScanJobRepository scans;
    private final ComponentRepository components;
    private final FindingRepository findings;
    private final ArtifactStorage storage;

    public RetentionService(ScanJobRepository scans, ComponentRepository components,
                            FindingRepository findings, ArtifactStorage storage) {
        this.scans = scans;
        this.components = components;
        this.findings = findings;
        this.storage = storage;
    }

    @Scheduled(fixedDelayString = "${app.scan.source-cleanup-ms:3600000}", initialDelay = 30000)
    @Transactional
    public void deleteExpiredSources() {
        for (var scan : scans.findTop100BySourceDeletedAtIsNullAndSourceExpiresAtBeforeOrderBySourceExpiresAtAsc(
                Instant.now())) {
            if (scan.getSourceObjectKey() != null) {
                try {
                    storage.delete(scan.getSourceObjectKey());
                } catch (RuntimeException exception) {
                    log.warn("Source cleanup failed for scan {}: {}", scan.getId(), exception.getClass().getSimpleName());
                    continue;
                }
            }
            scan.setSourceDeletedAt(Instant.now());
            scans.save(scan);
            log.info("Deleted expired source for scan {}", scan.getId());
        }
    }

    @Scheduled(cron = "${app.scan.result-cleanup-cron:0 15 3 * * *}")
    @Transactional
    public void deleteExpiredResults() {
        for (var scan : scans.findTop100ByResultExpiresAtBeforeOrderByResultExpiresAtAsc(Instant.now())) {
            findings.deleteByScanId(scan.getId());
            components.deleteByScanId(scan.getId());
            scans.delete(scan);
            log.info("Deleted expired scan conclusion {}", scan.getId());
        }
    }
}
