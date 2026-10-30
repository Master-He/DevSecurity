package com.devsecurity.core.service;

import com.devsecurity.core.api.ApiException;
import com.devsecurity.core.domain.AppUser;
import com.devsecurity.core.domain.Project;
import com.devsecurity.core.domain.ScanJob;
import com.devsecurity.core.domain.ScanStatus;
import com.devsecurity.core.domain.SourceType;
import com.devsecurity.core.repository.ComponentRepository;
import com.devsecurity.core.repository.FindingRepository;
import com.devsecurity.core.repository.ScanJobRepository;
import java.time.Duration;
import java.time.Instant;
import java.util.List;
import java.util.UUID;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.multipart.MultipartFile;

@Service
public class ScanService {
    private final ScanJobRepository scans;
    private final ComponentRepository components;
    private final FindingRepository findings;
    private final ProjectService projectService;
    private final SourceService sourceService;
    private final ArtifactStorage storage;
    private final int sourceRetentionDays;
    private final int resultRetentionDays;

    public ScanService(
            ScanJobRepository scans,
            ComponentRepository components,
            FindingRepository findings,
            ProjectService projectService,
            SourceService sourceService,
            ArtifactStorage storage,
            @Value("${app.scan.source-retention-days:3}") int sourceRetentionDays,
            @Value("${app.scan.result-retention-days:180}") int resultRetentionDays) {
        this.scans = scans;
        this.components = components;
        this.findings = findings;
        this.projectService = projectService;
        this.sourceService = sourceService;
        this.storage = storage;
        this.sourceRetentionDays = sourceRetentionDays;
        this.resultRetentionDays = resultRetentionDays;
    }

    public ScanJob createUpload(UUID projectId, MultipartFile file, AppUser user) {
        Project project = projectService.requireAccess(projectId, user);
        byte[] archive = sourceService.readValidatedUpload(file);
        UUID scanId = UUID.randomUUID();
        ScanJob scan = newScan(project, user, SourceType.UPLOAD, null, null);
        String objectKey = sourceKey(scanId);
        storage.put(objectKey, archive, "application/zip");
        scan.setSourceObjectKey(objectKey);
        try {
            return scans.save(scan);
        } catch (RuntimeException exception) {
            storage.delete(objectKey);
            throw exception;
        }
    }

    public ScanJob createGit(UUID projectId, String gitUrl, String revision, AppUser user) {
        Project project = projectService.requireAccess(projectId, user);
        if (gitUrl == null || gitUrl.isBlank()) {
            throw new ApiException(HttpStatus.BAD_REQUEST, "GIT_URL_REQUIRED", "请输入 Git 地址");
        }
        ScanJob scan = newScan(project, user, SourceType.GIT, gitUrl.trim(), revision);
        return scans.save(scan);
    }

    private ScanJob newScan(Project project, AppUser user, SourceType type, String ref, String revision) {
        Instant now = Instant.now();
        return new ScanJob(
                UUID.randomUUID(), project.getTenantId(), project.getId(), user.getUsername(),
                type, ref, revision,
                now.plus(Duration.ofDays(sourceRetentionDays)),
                now.plus(Duration.ofDays(resultRetentionDays)));
    }

    @Transactional(readOnly = true)
    public List<ScanJob> list(UUID projectId, AppUser user) {
        projectService.requireAccess(projectId, user);
        return scans.findByProjectIdOrderByCreatedAtDesc(projectId);
    }

    @Transactional(readOnly = true)
    public ScanJob requireScan(UUID scanId, AppUser user) {
        ScanJob scan = scans.findById(scanId)
                .orElseThrow(() -> new ApiException(HttpStatus.NOT_FOUND, "SCAN_NOT_FOUND", "扫描任务不存在"));
        projectService.requireAccess(scan.getProjectId(), user);
        return scan;
    }

    @Transactional(readOnly = true)
    public List<com.devsecurity.core.domain.Component> components(UUID scanId, AppUser user) {
        requireScan(scanId, user);
        return components.findByScanIdOrderByGroupIdAscArtifactIdAsc(scanId);
    }

    @Transactional(readOnly = true)
    public List<com.devsecurity.core.domain.Finding> findings(UUID scanId, AppUser user) {
        requireScan(scanId, user);
        return findings.findByScanIdOrderBySeverityAscTitleAsc(scanId);
    }

    @Transactional
    public ScanJob cancel(UUID scanId, AppUser user) {
        ScanJob scan = requireScan(scanId, user);
        if (scan.getStatus() == ScanStatus.QUEUED) {
            scan.cancel();
            return scans.save(scan);
        }
        if (scan.getStatus() == ScanStatus.RUNNING) {
            throw new ApiException(HttpStatus.CONFLICT, "SCAN_ALREADY_RUNNING",
                    "扫描已经开始，当前版本不支持中断正在执行的 SCA 任务");
        }
        throw new ApiException(HttpStatus.CONFLICT, "SCAN_NOT_CANCELABLE", "当前任务状态不可取消");
    }

    public String sourceKey(UUID scanId) {
        return "scans/" + scanId + "/source.zip";
    }
}
