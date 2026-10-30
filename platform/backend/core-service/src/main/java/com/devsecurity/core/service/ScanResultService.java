package com.devsecurity.core.service;

import com.devsecurity.core.domain.Component;
import com.devsecurity.core.domain.Finding;
import com.devsecurity.core.domain.ScanJob;
import com.devsecurity.core.domain.ScanStatus;
import com.devsecurity.core.repository.ComponentRepository;
import com.devsecurity.core.repository.FindingRepository;
import com.devsecurity.core.repository.ScanJobRepository;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class ScanResultService {
    private final ScanJobRepository scans;
    private final ComponentRepository components;
    private final FindingRepository findings;

    public ScanResultService(ScanJobRepository scans, ComponentRepository components,
                             FindingRepository findings) {
        this.scans = scans;
        this.components = components;
        this.findings = findings;
    }

    @Transactional
    public void complete(UUID scanId, ScaClient.ScaResult result) {
        ScanJob scan = scans.findById(scanId)
                .orElseThrow(() -> new IllegalStateException("Scan disappeared"));
        if (scan.getStatus() != ScanStatus.RUNNING) {
            return;
        }
        findings.deleteByScanId(scanId);
        components.deleteByScanId(scanId);

        Map<String, Component> byPurl = new HashMap<>();
        var savedComponents = new ArrayList<Component>();
        if (result.components() != null) {
            for (ScaClient.ScaComponent item : result.components()) {
                Component component = new Component(
                        scanId,
                        item.ecosystem(),
                        item.groupId(),
                        item.artifactId(),
                        item.version(),
                        item.purl(),
                        item.direct(),
                        item.dependencyPath(),
                        item.scope());
                savedComponents.add(component);
                byPurl.putIfAbsent(item.purl(), component);
            }
        }
        components.saveAll(savedComponents);

        var savedFindings = new ArrayList<Finding>();
        if (result.findings() != null) {
            for (ScaClient.ScaFinding item : result.findings()) {
                Component component = byPurl.get(item.purl());
                if (component == null) {
                    continue;
                }
                savedFindings.add(new Finding(
                        scanId,
                        component.getId(),
                        item.vulnerabilityId(),
                        join(item.aliases()),
                        item.title(),
                        item.severity(),
                        item.packageName(),
                        item.affectedVersion(),
                        item.affectedRange(),
                        join(item.fixedVersions()),
                        item.dependencyPath()));
            }
        }
        findings.saveAll(savedFindings);
        scan.complete(savedComponents.size(), savedFindings.size(),
                result.sourceMode(), result.engineVersion());
        scans.save(scan);
    }

    @Transactional
    public void fail(UUID scanId, String code, String message) {
        scans.findById(scanId).ifPresent(scan -> {
            if (scan.getStatus() == ScanStatus.RUNNING || scan.getStatus() == ScanStatus.QUEUED) {
                scan.fail(code, message);
                scans.save(scan);
            }
        });
    }

    private String join(java.util.List<String> values) {
        return values == null ? "" : String.join(", ", values);
    }
}
