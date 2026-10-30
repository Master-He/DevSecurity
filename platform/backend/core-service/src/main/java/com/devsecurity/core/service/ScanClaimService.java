package com.devsecurity.core.service;

import com.devsecurity.core.repository.ScanJobRepository;
import java.util.UUID;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class ScanClaimService {
    private final ScanJobRepository scans;

    public ScanClaimService(ScanJobRepository scans) {
        this.scans = scans;
    }

    @Transactional
    public boolean claim(UUID scanId) {
        return scans.claim(scanId) == 1;
    }
}
