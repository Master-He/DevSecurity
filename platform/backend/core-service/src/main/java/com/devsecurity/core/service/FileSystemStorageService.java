package com.devsecurity.core.service;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardCopyOption;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.stereotype.Service;

@Service
@ConditionalOnProperty(name = "app.storage.kind", havingValue = "filesystem")
public class FileSystemStorageService implements ArtifactStorage {
    private final Path root;

    public FileSystemStorageService(@Value("${app.storage.directory:.data/artifacts}") String directory)
            throws IOException {
        this.root = Path.of(directory).toAbsolutePath().normalize();
        Files.createDirectories(root);
    }

    @Override
    public void put(String key, byte[] bytes, String contentType) {
        Path target = resolve(key);
        try {
            Files.createDirectories(target.getParent());
            Path temporary = Files.createTempFile(target.getParent(), ".source-", ".tmp");
            Files.write(temporary, bytes);
            Files.move(temporary, target, StandardCopyOption.REPLACE_EXISTING, StandardCopyOption.ATOMIC_MOVE);
        } catch (IOException exception) {
            throw new IllegalStateException("Cannot persist local artifact", exception);
        }
    }

    @Override
    public byte[] get(String key) {
        try {
            return Files.readAllBytes(resolve(key));
        } catch (IOException exception) {
            throw new IllegalStateException("Cannot read local artifact", exception);
        }
    }

    @Override
    public void copyTo(String key, Path target) {
        try {
            Files.createDirectories(target.getParent());
            Files.copy(resolve(key), target, StandardCopyOption.REPLACE_EXISTING);
        } catch (IOException exception) {
            throw new IllegalStateException("Cannot copy local artifact", exception);
        }
    }

    @Override
    public void delete(String key) {
        try {
            Files.deleteIfExists(resolve(key));
        } catch (IOException exception) {
            throw new IllegalStateException("Cannot delete local artifact", exception);
        }
    }

    private Path resolve(String key) {
        Path resolved = root.resolve(key).normalize();
        if (!resolved.startsWith(root)) {
            throw new IllegalArgumentException("Invalid artifact key");
        }
        return resolved;
    }
}
