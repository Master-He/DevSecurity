package com.devsecurity.core.service;

import java.nio.file.Path;

public interface ArtifactStorage {
    void put(String key, byte[] bytes, String contentType);
    byte[] get(String key);
    void copyTo(String key, Path target);
    void delete(String key);
}
