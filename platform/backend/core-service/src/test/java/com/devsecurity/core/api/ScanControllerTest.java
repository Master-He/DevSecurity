package com.devsecurity.core.api;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import com.devsecurity.core.domain.AppUser;
import com.devsecurity.core.domain.Finding;
import com.devsecurity.core.domain.ScanJob;
import com.devsecurity.core.domain.ScanType;
import com.devsecurity.core.domain.SourceType;
import com.devsecurity.core.security.CurrentUser;
import com.devsecurity.core.service.ScanService;
import java.util.List;
import java.util.UUID;
import java.time.Instant;
import org.junit.jupiter.api.Test;
import org.springframework.security.core.Authentication;

class ScanControllerTest {
    @Test
    void serializesSastFindingWithoutScAFields() {
        ScanService scans = mock(ScanService.class);
        CurrentUser currentUser = mock(CurrentUser.class);
        Authentication authentication = mock(Authentication.class);
        AppUser user = new AppUser(
                UUID.randomUUID(), "default", "admin", "hash", "管理员", "ADMIN");
        UUID scanId = UUID.randomUUID();
        when(currentUser.require(authentication)).thenReturn(user);
        when(scans.requireScan(scanId, user)).thenReturn(null);
        when(scans.findings(scanId, user)).thenReturn(List.of(
                Finding.sast(scanId, "SAST-JAVA-001", "潜在命令执行", "HIGH",
                        "MEDIUM", "command-injection", "CWE-78",
                        "Demo.java", 8, 8,
                        "Runtime.getRuntime().exec(cmd)", "使用固定命令")));

        ScanController controller = new ScanController(scans, currentUser);
        var response = controller.findings(scanId, authentication);

        assertThat(response).singleElement().satisfies(finding -> {
            assertThat(finding.get("findingType")).isEqualTo("SAST");
            assertThat(finding.get("ruleId")).isEqualTo("SAST-JAVA-001");
            assertThat(finding.get("vulnerabilityId")).isEqualTo("");
            assertThat(finding.get("filePath")).isEqualTo("Demo.java");
        });
    }

    @Test
    void omittedGitScanTypeDefaultsToScA() {
        ScanService scans = mock(ScanService.class);
        CurrentUser currentUser = mock(CurrentUser.class);
        Authentication authentication = mock(Authentication.class);
        AppUser user = new AppUser(
                UUID.randomUUID(), "default", "admin", "hash", "管理员", "ADMIN");
        UUID projectId = UUID.randomUUID();
        UUID scanId = UUID.randomUUID();
        Instant now = Instant.now();
        ScanJob scan = new ScanJob(
                scanId, "default", projectId, "admin", ScanType.SCA,
                SourceType.GIT, "https://gitlab.com/team/repo.git", "main",
                now.plusSeconds(3600), now.plusSeconds(7200));
        when(currentUser.require(authentication)).thenReturn(user);
        when(scans.createGit(
                projectId, "https://gitlab.com/team/repo.git", "main", user, ScanType.SCA))
                .thenReturn(scan);

        ScanController controller = new ScanController(scans, currentUser);
        controller.git(projectId,
                new ScanController.GitScanRequest(
                        "https://gitlab.com/team/repo.git", "main", null),
                authentication);

        verify(scans).createGit(
                projectId, "https://gitlab.com/team/repo.git", "main", user, ScanType.SCA);
    }
}
