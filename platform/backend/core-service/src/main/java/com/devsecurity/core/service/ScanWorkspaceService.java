package com.devsecurity.core.service;

import com.devsecurity.core.api.ApiException;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.AtomicMoveNotSupportedException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardCopyOption;
import java.util.Comparator;
import java.util.UUID;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;

@Service
public class ScanWorkspaceService {
    private static final Logger log = LoggerFactory.getLogger(ScanWorkspaceService.class);

    private final Path root;

    public ScanWorkspaceService(@Value("${app.scan.work-root:.data/sca-work}") String directory)
            throws IOException {
        this.root = Path.of(directory).toAbsolutePath().normalize();
        Files.createDirectories(root);
    }

    public ScanWorkspace publish(UUID scanId, ArtifactStorage storage, String sourceKey) {
        Path scans = root.resolve("scans");
        Path target = safeScanPath(scans, scanId);
        if (Files.exists(target)) {
            throw new ApiException(HttpStatus.CONFLICT, "SCA_WORKSPACE_CONFLICT",
                    "扫描临时工作目录已存在");
        }

        Path staging = null;
        boolean published = false;
        try {
            Files.createDirectories(scans);
            staging = Files.createTempDirectory(scans, "." + scanId + "-");
            Path input = Files.createDirectories(staging.resolve("input"));
            Files.createDirectories(staging.resolve("output"));
            Files.createDirectories(staging.resolve("logs"));
            Path source = input.resolve("source.zip");
            storage.copyTo(sourceKey, source);
            if (!Files.isRegularFile(source) || Files.size(source) == 0) {
                throw new IllegalStateException("Published source archive is empty");
            }
            moveAtomically(staging, target);
            published = true;
            writeReadyMarker(target);
            return layout(scanId, target);
        } catch (ApiException exception) {
            deleteQuietly(published ? target : staging);
            throw exception;
        } catch (Exception exception) {
            deleteQuietly(published ? target : staging);
            log.warn("Cannot publish shared scan workspace for {}", scanId,
                    exception.getClass().getSimpleName());
            throw new ApiException(HttpStatus.INTERNAL_SERVER_ERROR, "SCA_WORKSPACE_PUBLISH_FAILED",
                    "无法发布扫描临时工作目录");
        }
    }

    public void cleanup(ScanWorkspace workspace) {
        if (workspace != null) {
            deleteQuietly(workspace.scanRoot());
        }
    }

    private ScanWorkspace layout(UUID scanId, Path scanRoot) {
        return new ScanWorkspace(
                scanId,
                scanRoot,
                scanRoot.resolve("input"),
                scanRoot.resolve("output"),
                scanRoot.resolve("logs"),
                scanRoot.resolve("READY"),
                scanRoot.resolve("output").resolve("result.json"));
    }

    private Path safeScanPath(Path scans, UUID scanId) {
        Path target = scans.resolve(scanId.toString()).normalize();
        if (!target.startsWith(scans)) {
            throw new IllegalArgumentException("Invalid scan workspace path");
        }
        return target;
    }

    private void writeReadyMarker(Path scanRoot) throws IOException {
        Path temporary = Files.createTempFile(scanRoot, ".READY-", ".tmp");
        try {
            Files.writeString(temporary, "ready\n", StandardCharsets.UTF_8);
            moveAtomically(temporary, scanRoot.resolve("READY"));
        } finally {
            Files.deleteIfExists(temporary);
        }
    }

    private void moveAtomically(Path source, Path target) throws IOException {
        try {
            Files.move(source, target, StandardCopyOption.ATOMIC_MOVE);
        } catch (AtomicMoveNotSupportedException exception) {
            Files.move(source, target);
        }
    }

    private void deleteQuietly(Path path) {
        if (path == null || !Files.exists(path)) {
            return;
        }
        try (var paths = Files.walk(path)) {
            paths.sorted(Comparator.reverseOrder()).forEach(candidate -> {
                try {
                    Files.deleteIfExists(candidate);
                } catch (IOException ignored) {
                }
            });
        } catch (IOException ignored) {
        }
    }

    public record ScanWorkspace(
            UUID scanId,
            Path scanRoot,
            Path input,
            Path output,
            Path logs,
            Path ready,
            Path result) {
    }
}
