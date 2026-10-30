package com.devsecurity.core.service;

import com.devsecurity.core.api.ApiException;
import com.devsecurity.core.domain.ScanJob;
import com.devsecurity.core.domain.ScanStatus;
import com.devsecurity.core.domain.SourceType;
import com.devsecurity.core.repository.ScanJobRepository;
import java.util.UUID;
import java.util.concurrent.Executor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

@Component
public class ScanWorker {
    private static final Logger log = LoggerFactory.getLogger(ScanWorker.class);

    private final ScanJobRepository scans;
    private final ScanClaimService claimService;
    private final SourceService sourceService;
    private final ArtifactStorage storage;
    private final ScanWorkspaceService workspaceService;
    private final ScaClient scaClient;
    private final ScanResultService resultService;
    private final Executor executor;

    public ScanWorker(ScanJobRepository scans,
                      ScanClaimService claimService,
                      SourceService sourceService,
                      ArtifactStorage storage,
                      ScanWorkspaceService workspaceService,
                      ScaClient scaClient,
                      ScanResultService resultService,
                      @Qualifier("scanExecutor") Executor executor) {
        this.scans = scans;
        this.claimService = claimService;
        this.sourceService = sourceService;
        this.storage = storage;
        this.workspaceService = workspaceService;
        this.scaClient = scaClient;
        this.resultService = resultService;
        this.executor = executor;
    }

    @Scheduled(fixedDelayString = "${app.scan.worker-poll-ms:3000}")
    public void dispatch() {
        for (UUID scanId : scans.findIdsByStatus(ScanStatus.QUEUED).stream().limit(8).toList()) {
            if (claimService.claim(scanId)) {
                executor.execute(() -> process(scanId));
            }
        }
    }

    void process(UUID scanId) {
        ScanWorkspaceService.ScanWorkspace workspace = null;
        boolean submitted = false;
        boolean completed = false;
        try {
            ScanJob scan = scans.findById(scanId)
                    .orElseThrow(() -> new IllegalStateException("Claimed scan disappeared"));
            String key = scan.getSourceObjectKey();
            if (scan.getSourceType() == SourceType.GIT) {
                byte[] archive = sourceService.cloneGitArchive(scan.getSourceRef(), scan.getSourceRevision());
                key = "scans/" + scanId + "/source.zip";
                storage.put(key, archive, "application/zip");
                scan.setSourceObjectKey(key);
                scans.save(scan);
            }
            if (key == null) {
                throw new ApiException(org.springframework.http.HttpStatus.BAD_REQUEST,
                        "SOURCE_MISSING", "扫描源码不存在");
            }
            workspace = workspaceService.publish(scanId, storage, key);
            scaClient.submit(scanId);
            submitted = true;
            scaClient.waitForCompletion(scanId);
            ScaClient.ScaResult result = scaClient.readResult(workspace.result());
            resultService.complete(scanId, result);
            completed = true;
            log.info("SCA scan {} completed: components={}, findings={}",
                    scanId,
                    result.components() == null ? 0 : result.components().size(),
                    result.findings() == null ? 0 : result.findings().size());
        } catch (ApiException exception) {
            resultService.fail(scanId, exception.getCode(), exception.getMessage());
            log.warn("SCA scan {} failed: {}", scanId, exception.getCode());
        } catch (Exception exception) {
            resultService.fail(scanId, "SCAN_WORKER_ERROR", "扫描执行失败，请查看服务日志");
            log.error("Unexpected SCA worker failure for scan {}", scanId, exception);
        } finally {
            if (submitted && !completed) {
                try {
                    scaClient.cancelAndWait(scanId);
                } catch (ApiException exception) {
                    log.warn("SCA cancel cleanup failed for scan {}: {}", scanId, exception.getCode());
                }
            }
            workspaceService.cleanup(workspace);
        }
    }
}
