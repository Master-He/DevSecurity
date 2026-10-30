package com.devsecurity.core.service;

import static org.assertj.core.api.Assertions.assertThat;

import java.nio.file.Files;
import java.nio.file.Path;
import java.util.UUID;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

class ScanWorkspaceServiceTest {
    @TempDir
    Path temporary;

    @Test
    void publishesAndCleansScanWorkspace() throws Exception {
        ScanWorkspaceService service = new ScanWorkspaceService(temporary.toString());
        UUID scanId = UUID.randomUUID();
        byte[] archive = "zip-content".getBytes();
        ArtifactStorage storage = new FakeStorage(archive);

        ScanWorkspaceService.ScanWorkspace workspace =
                service.publish(scanId, storage, "scans/source.zip");

        assertThat(workspace.scanRoot()).isEqualTo(temporary.resolve("scans").resolve(scanId.toString()));
        assertThat(workspace.input().resolve("source.zip")).hasBinaryContent(archive);
        assertThat(workspace.ready()).exists();
        assertThat(workspace.output()).isDirectory();
        assertThat(workspace.logs()).isDirectory();

        service.cleanup(workspace);
        assertThat(workspace.scanRoot()).doesNotExist();
    }

    private static final class FakeStorage implements ArtifactStorage {
        private final byte[] content;

        private FakeStorage(byte[] content) {
            this.content = content;
        }

        @Override
        public void put(String key, byte[] bytes, String contentType) {
        }

        @Override
        public byte[] get(String key) {
            return content;
        }

        @Override
        public void copyTo(String key, Path target) {
            try {
                Files.write(target, content);
            } catch (Exception exception) {
                throw new IllegalStateException(exception);
            }
        }

        @Override
        public void delete(String key) {
        }
    }
}
