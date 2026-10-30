package com.devsecurity.core.api;

import com.devsecurity.core.domain.AppUser;
import com.devsecurity.core.domain.ScanJob;
import com.devsecurity.core.security.CurrentUser;
import com.devsecurity.core.service.ScanService;
import jakarta.validation.Valid;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;
import java.time.Instant;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import org.springframework.http.HttpStatus;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestPart;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.multipart.MultipartFile;

@RestController
@RequestMapping("/api/v1")
public class ScanController {
    private final ScanService scans;
    private final CurrentUser currentUser;

    public ScanController(ScanService scans, CurrentUser currentUser) {
        this.scans = scans;
        this.currentUser = currentUser;
    }

    @GetMapping("/projects/{projectId}/scans")
    List<Map<String, Object>> list(@PathVariable UUID projectId, Authentication authentication) {
        AppUser user = currentUser.require(authentication);
        return scans.list(projectId, user).stream().map(ScanController::view).toList();
    }

    @PostMapping(value = "/projects/{projectId}/scans/upload", consumes = "multipart/form-data")
    @ResponseStatus(HttpStatus.ACCEPTED)
    Map<String, Object> upload(@PathVariable UUID projectId,
                               @RequestPart("file") MultipartFile file,
                               Authentication authentication) {
        AppUser user = currentUser.require(authentication);
        return view(scans.createUpload(projectId, file, user));
    }

    @PostMapping("/projects/{projectId}/scans")
    @ResponseStatus(HttpStatus.ACCEPTED)
    Map<String, Object> git(@PathVariable UUID projectId,
                            @Valid @RequestBody GitScanRequest request,
                            Authentication authentication) {
        AppUser user = currentUser.require(authentication);
        return view(scans.createGit(projectId, request.gitUrl(), request.revision(), user));
    }

    @GetMapping("/scans/{scanId}")
    Map<String, Object> get(@PathVariable UUID scanId, Authentication authentication) {
        return view(scans.requireScan(scanId, currentUser.require(authentication)));
    }

    @GetMapping("/scans/{scanId}/components")
    List<Map<String, Object>> components(@PathVariable UUID scanId, Authentication authentication) {
        return scans.components(scanId, currentUser.require(authentication)).stream()
                .map(component -> Map.<String, Object>of(
                        "id", component.getId(),
                        "ecosystem", component.getEcosystem(),
                        "groupId", component.getGroupId(),
                        "artifactId", component.getArtifactId(),
                        "version", component.getVersion(),
                        "purl", component.getPurl(),
                        "direct", component.isDirect(),
                        "dependencyPath", component.getDependencyPath(),
                        "scope", component.getScope() == null ? "" : component.getScope()))
                .toList();
    }

    @GetMapping("/scans/{scanId}/findings")
    List<Map<String, Object>> findings(@PathVariable UUID scanId, Authentication authentication) {
        return scans.findings(scanId, currentUser.require(authentication)).stream()
                .map(finding -> Map.<String, Object>ofEntries(
                        Map.entry("id", finding.getId()),
                        Map.entry("vulnerabilityId", finding.getVulnerabilityId()),
                        Map.entry("aliases", finding.getAliases()),
                        Map.entry("title", finding.getTitle()),
                        Map.entry("severity", finding.getSeverity()),
                        Map.entry("packageName", finding.getPackageName()),
                        Map.entry("affectedVersion", finding.getAffectedVersion()),
                        Map.entry("affectedRange", finding.getAffectedRange() == null ? "" : finding.getAffectedRange()),
                        Map.entry("fixedVersions", finding.getFixedVersions()),
                        Map.entry("dependencyPath", finding.getDependencyPath()),
                        Map.entry("createdAt", finding.getCreatedAt())))
                .toList();
    }

    @PostMapping("/scans/{scanId}/cancel")
    Map<String, Object> cancel(@PathVariable UUID scanId, Authentication authentication) {
        return view(scans.cancel(scanId, currentUser.require(authentication)));
    }

    static Map<String, Object> view(ScanJob scan) {
        return Map.ofEntries(
                Map.entry("id", scan.getId()),
                Map.entry("projectId", scan.getProjectId()),
                Map.entry("status", scan.getStatus()),
                Map.entry("sourceType", scan.getSourceType()),
                Map.entry("sourceRef", scan.getSourceRef() == null ? "" : scan.getSourceRef()),
                Map.entry("sourceRevision", scan.getSourceRevision() == null ? "" : scan.getSourceRevision()),
                Map.entry("sourceMode", scan.getSourceMode() == null ? "" : scan.getSourceMode()),
                Map.entry("engineVersion", scan.getEngineVersion() == null ? "" : scan.getEngineVersion()),
                Map.entry("failureCode", scan.getFailureCode() == null ? "" : scan.getFailureCode()),
                Map.entry("failureMessage", scan.getFailureMessage() == null ? "" : scan.getFailureMessage()),
                Map.entry("componentCount", scan.getComponentCount() == null ? 0 : scan.getComponentCount()),
                Map.entry("findingCount", scan.getFindingCount() == null ? 0 : scan.getFindingCount()),
                Map.entry("sourceExpiresAt", scan.getSourceExpiresAt()),
                Map.entry("sourceDeleted", scan.getSourceDeletedAt() != null),
                Map.entry("resultExpiresAt", scan.getResultExpiresAt()),
                Map.entry("createdAt", scan.getCreatedAt()),
                Map.entry("startedAt", scan.getStartedAt() == null ? Instant.EPOCH : scan.getStartedAt()),
                Map.entry("finishedAt", scan.getFinishedAt() == null ? Instant.EPOCH : scan.getFinishedAt()));
    }

    public record GitScanRequest(
            @NotBlank @Size(max = 2000) String gitUrl,
            @Size(max = 200) String revision) {
    }
}
