package com.devsecurity.core.service;

import com.devsecurity.core.api.ApiException;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.net.URI;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Comparator;
import java.util.List;
import java.util.Locale;
import java.util.Set;
import java.util.zip.ZipEntry;
import java.util.zip.ZipInputStream;
import java.util.zip.ZipOutputStream;
import org.eclipse.jgit.api.Git;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.web.multipart.MultipartFile;

@Service
public class SourceService {
    private static final long MAX_ARCHIVE_BYTES = 100L * 1024 * 1024;
    private static final long MAX_UNCOMPRESSED_BYTES = 500L * 1024 * 1024;
    private static final int MAX_ENTRIES = 20_000;

    private final Set<String> allowedGitHosts;
    private final int gitTimeoutSeconds;

    public SourceService(@Value("${app.git.allowed-hosts:}") String allowedHosts,
                         @Value("${app.git.timeout-seconds:120}") int gitTimeoutSeconds) {
        this.allowedGitHosts = Set.copyOf(
                List.of(allowedHosts.split(",")).stream()
                        .map(String::trim)
                        .map(value -> value.toLowerCase(Locale.ROOT))
                        .filter(value -> !value.isBlank())
                        .toList());
        this.gitTimeoutSeconds = gitTimeoutSeconds;
    }

    public byte[] readValidatedUpload(MultipartFile file) {
        if (file == null || file.isEmpty()) {
            throw new ApiException(HttpStatus.BAD_REQUEST, "SOURCE_REQUIRED", "请选择 ZIP 源码包");
        }
        if (file.getSize() > MAX_ARCHIVE_BYTES) {
            throw new ApiException(HttpStatus.PAYLOAD_TOO_LARGE, "UPLOAD_TOO_LARGE", "源码归档不能超过 100 MB");
        }
        try {
            byte[] bytes = file.getBytes();
            requirePom(bytes);
            return bytes;
        } catch (IOException exception) {
            throw new ApiException(HttpStatus.BAD_REQUEST, "INVALID_ARCHIVE", "无法读取源码归档");
        }
    }

    public byte[] cloneGitArchive(String gitUrl, String revision) {
        URI uri = validateGitUri(gitUrl);
        Path workspace = null;
        try {
            workspace = Files.createTempDirectory("devsecurity-git-");
            Path repository = workspace.resolve("repository");
            var command = Git.cloneRepository()
                    .setURI(uri.toString())
                    .setDirectory(repository.toFile())
                    .setCloneAllBranches(false)
                    .setTimeout(gitTimeoutSeconds);
            if (revision != null && !revision.isBlank()) {
                command.setBranch(revision.trim());
            }
            try (Git ignored = command.call()) {
                byte[] archive = zipDirectory(repository);
                requirePom(archive);
                return archive;
            }
        } catch (ApiException exception) {
            throw exception;
        } catch (Exception exception) {
            throw new ApiException(HttpStatus.BAD_REQUEST, "GIT_CLONE_FAILED",
                    "无法获取 Git 源码，请检查地址、分支和网络访问");
        } finally {
            deleteRecursively(workspace);
        }
    }

    private URI validateGitUri(String value) {
        if (value == null || value.isBlank()) {
            throw new ApiException(HttpStatus.BAD_REQUEST, "GIT_URL_REQUIRED", "请输入 Git 地址");
        }
        try {
            URI uri = URI.create(value.trim());
            String scheme = uri.getScheme();
            if (!("http".equalsIgnoreCase(scheme) || "https".equalsIgnoreCase(scheme))) {
                throw new ApiException(HttpStatus.BAD_REQUEST, "GIT_URL_INVALID", "Git 地址仅支持 HTTP/HTTPS");
            }
            if (uri.getHost() == null || uri.getUserInfo() != null
                    || !allowedGitHosts.contains(uri.getHost().toLowerCase(Locale.ROOT))) {
                throw new ApiException(HttpStatus.BAD_REQUEST, "GIT_HOST_NOT_ALLOWED",
                        "Git 主机不在允许列表中，请联系管理员配置 GIT_ALLOWED_HOSTS");
            }
            return uri;
        } catch (IllegalArgumentException exception) {
            throw new ApiException(HttpStatus.BAD_REQUEST, "GIT_URL_INVALID", "Git 地址不合法");
        }
    }

    private ZipValidation validateZip(byte[] bytes) throws IOException {
        int entries = 0;
        long expanded = 0;
        boolean hasPom = false;
        byte[] buffer = new byte[8192];
        try (ZipInputStream zip = new ZipInputStream(new ByteArrayInputStream(bytes))) {
            ZipEntry entry;
            while ((entry = zip.getNextEntry()) != null) {
                entries++;
                if (entries > MAX_ENTRIES) {
                    throw new ApiException(HttpStatus.BAD_REQUEST, "ARCHIVE_TOO_MANY_FILES",
                            "源码归档文件数量超过限制");
                }
                String name = safeEntryName(entry.getName());
                hasPom = hasPom || name.endsWith("/pom.xml") || "pom.xml".equals(name);
                int read;
                while ((read = zip.read(buffer)) != -1) {
                    expanded += read;
                    if (expanded > MAX_UNCOMPRESSED_BYTES) {
                        throw new ApiException(HttpStatus.BAD_REQUEST, "ARCHIVE_TOO_LARGE",
                                "源码归档解压后体积超过限制");
                    }
                }
            }
        }
        if (entries == 0) {
            throw new ApiException(HttpStatus.BAD_REQUEST, "EMPTY_ARCHIVE", "源码归档为空");
        }
        return new ZipValidation(hasPom);
    }

    private String safeEntryName(String name) {
        if (name == null || name.contains("\0")) {
            throw new ApiException(HttpStatus.BAD_REQUEST, "ARCHIVE_PATH_INVALID", "源码归档路径不合法");
        }
        String normalized = name.replace('\\', '/');
        if (normalized.startsWith("/") || normalized.matches("^[A-Za-z]:.*")
                || List.of(normalized.split("/")).contains("..")) {
            throw new ApiException(HttpStatus.BAD_REQUEST, "ARCHIVE_PATH_INVALID", "源码归档包含不安全路径");
        }
        return normalized;
    }

    private byte[] zipDirectory(Path root) throws IOException {
        ByteArrayOutputStream output = new ByteArrayOutputStream();
        long total = 0;
        List<Path> files;
        try (var paths = Files.walk(root)) {
            files = paths
                    .filter(Files::isRegularFile)
                    .filter(path -> !path.toString().contains(Path.of(".git").toString()))
                    .filter(path -> !path.toString().contains(Path.of("target").toString()))
                    .sorted(Comparator.comparing(Path::toString))
                    .toList();
        }
        try (ZipOutputStream zip = new ZipOutputStream(output)) {
            for (Path file : files) {
                long size = Files.size(file);
                total += size;
                if (total > MAX_UNCOMPRESSED_BYTES) {
                    throw new ApiException(HttpStatus.BAD_REQUEST, "SOURCE_TOO_LARGE", "Git 源码超过体积限制");
                }
                String name = root.relativize(file).toString().replace('\\', '/');
                ZipEntry entry = new ZipEntry(name);
                entry.setTime(0);
                zip.putNextEntry(entry);
                Files.copy(file, zip);
                zip.closeEntry();
            }
        }
        if (files.isEmpty()) {
            throw new ApiException(HttpStatus.BAD_REQUEST, "SOURCE_EMPTY", "Git 仓库为空");
        }
        return output.toByteArray();
    }

    private void requirePom(byte[] archive) throws IOException {
        if (!validateZip(archive).hasPom()) {
            throw new ApiException(HttpStatus.BAD_REQUEST, "POM_NOT_FOUND", "源码中未找到 pom.xml");
        }
    }

    private void deleteRecursively(Path root) {
        if (root == null || !Files.exists(root)) {
            return;
        }
        try (var paths = Files.walk(root)) {
            paths.sorted(Comparator.reverseOrder()).forEach(path -> {
                try {
                    Files.deleteIfExists(path);
                } catch (IOException ignored) {
                    // Temporary cleanup failure is reported by the host filesystem, never content.
                }
            });
        } catch (IOException ignored) {
        }
    }

    private record ZipValidation(boolean hasPom) {
    }
}
