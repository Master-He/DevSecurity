package com.devsecurity.core.service;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;

import com.devsecurity.core.api.ApiException;
import java.io.ByteArrayOutputStream;
import java.nio.charset.StandardCharsets;
import java.util.zip.ZipEntry;
import java.util.zip.ZipOutputStream;
import org.junit.jupiter.api.Test;
import org.springframework.mock.web.MockMultipartFile;

class SourceServiceTest {
    private final SourceService service = new SourceService("gitlab.com", 120);

    @Test
    void acceptsZipContainingPom() throws Exception {
        byte[] archive = zip("demo/pom.xml", "<project/>");
        MockMultipartFile file = new MockMultipartFile(
                "file", "source.zip", "application/zip", archive);

        assertEquals(archive.length, service.readValidatedUpload(file).length);
    }

    @Test
    void rejectsZipWithoutPom() throws Exception {
        MockMultipartFile file = new MockMultipartFile(
                "file", "source.zip", "application/zip",
                zip("demo/readme.txt", "hello"));

        ApiException exception = assertThrows(ApiException.class,
                () -> service.readValidatedUpload(file));
        assertEquals("POM_NOT_FOUND", exception.getCode());
    }

    @Test
    void rejectsPathTraversal() throws Exception {
        MockMultipartFile file = new MockMultipartFile(
                "file", "source.zip", "application/zip",
                zip("../pom.xml", "<project/>"));

        ApiException exception = assertThrows(ApiException.class,
                () -> service.readValidatedUpload(file));
        assertEquals("ARCHIVE_PATH_INVALID", exception.getCode());
    }

    private byte[] zip(String name, String content) throws Exception {
        ByteArrayOutputStream bytes = new ByteArrayOutputStream();
        try (ZipOutputStream zip = new ZipOutputStream(bytes)) {
            zip.putNextEntry(new ZipEntry(name));
            zip.write(content.getBytes(StandardCharsets.UTF_8));
            zip.closeEntry();
        }
        return bytes.toByteArray();
    }
}
