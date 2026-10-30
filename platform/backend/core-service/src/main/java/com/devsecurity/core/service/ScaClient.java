package com.devsecurity.core.service;

import com.devsecurity.core.api.ApiException;
import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.ObjectMapper;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Duration;
import java.time.Instant;
import java.util.List;
import java.util.UUID;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;

@Service
public class ScaClient {
    private static final Duration REQUEST_TIMEOUT = Duration.ofSeconds(30);

    private final ObjectMapper objectMapper;
    private final HttpClient httpClient;
    private final URI scansUri;
    private final String serviceToken;
    private final Duration executionTimeout;
    private final Duration pollInterval;

    public ScaClient(
            ObjectMapper objectMapper,
            @Value("${app.scan.sca-url}") String scaUrl,
            @Value("${app.scan.sca-service-token:}") String serviceToken,
            @Value("${app.scan.execution-timeout-seconds:1800}") long timeoutSeconds,
            @Value("${app.scan.sca-poll-ms:1000}") long pollMilliseconds) {
        this.objectMapper = objectMapper;
        this.serviceToken = serviceToken == null ? "" : serviceToken.trim();
        this.executionTimeout = Duration.ofSeconds(Math.max(1, timeoutSeconds));
        this.pollInterval = Duration.ofMillis(Math.max(100, pollMilliseconds));
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(10))
                .build();
        this.scansUri = URI.create(scaUrl.replaceAll("/+$", "") + "/v1/scans");
    }

    public void submit(UUID scanId) {
        try {
            ScanSubmission payload = new ScanSubmission(
                    "1.0",
                    scanId.toString(),
                    "zip",
                    "java-maven",
                    new ScanLimits((int) executionTimeout.toSeconds()));
            HttpRequest request = request(scansUri)
                    .header("Content-Type", "application/json")
                    .POST(HttpRequest.BodyPublishers.ofByteArray(
                            objectMapper.writeValueAsBytes(payload)))
                    .build();
            HttpResponse<byte[]> response = httpClient.send(
                    request, HttpResponse.BodyHandlers.ofByteArray());
            if (response.statusCode() / 100 != 2) {
                throw responseError(response, "SCA_SUBMIT_FAILED",
                        "SCA 控制服务拒绝扫描任务");
            }
        } catch (ApiException exception) {
            throw exception;
        } catch (InterruptedException exception) {
            Thread.currentThread().interrupt();
            throw new ApiException(HttpStatus.BAD_GATEWAY, "SCA_UNAVAILABLE", "SCA 请求被中断");
        } catch (Exception exception) {
            throw new ApiException(HttpStatus.BAD_GATEWAY, "SCA_UNAVAILABLE",
                    "SCA 控制服务暂时不可用");
        }
    }

    public JobStatus waitForCompletion(UUID scanId) {
        Instant deadline = Instant.now().plus(executionTimeout);
        while (Instant.now().isBefore(deadline)) {
            JobStatus status = status(scanId);
            if ("succeeded".equals(status.state())) {
                return status;
            }
            if ("failed".equals(status.state())) {
                String code = blankToDefault(status.errorCode(), "SCA_FAILED");
                String message = blankToDefault(status.message(), "SCA 引擎执行失败");
                throw new ApiException(HttpStatus.BAD_GATEWAY, code, message);
            }
            if ("canceled".equals(status.state())) {
                throw new ApiException(HttpStatus.CONFLICT, "SCA_CANCELED", "SCA 扫描已取消");
            }
            try {
                Thread.sleep(pollInterval.toMillis());
            } catch (InterruptedException exception) {
                Thread.currentThread().interrupt();
                throw new ApiException(HttpStatus.BAD_GATEWAY, "SCA_UNAVAILABLE", "SCA 状态轮询被中断");
            }
        }
        throw new ApiException(HttpStatus.GATEWAY_TIMEOUT, "SCA_TIMEOUT", "SCA 扫描执行超时");
    }

    public void cancelAndWait(UUID scanId) {
        cancel(scanId);
        Instant deadline = Instant.now().plusSeconds(10);
        while (Instant.now().isBefore(deadline)) {
            JobStatus status = status(scanId);
            if ("succeeded".equals(status.state()) || "failed".equals(status.state())
                    || "canceled".equals(status.state())) {
                return;
            }
            try {
                Thread.sleep(pollInterval.toMillis());
            } catch (InterruptedException exception) {
                Thread.currentThread().interrupt();
                throw new ApiException(HttpStatus.BAD_GATEWAY, "SCA_UNAVAILABLE", "SCA 取消等待被中断");
            }
        }
        throw new ApiException(HttpStatus.BAD_GATEWAY, "SCA_CANCEL_TIMEOUT", "等待 SCA 停止超时");
    }

    private void cancel(UUID scanId) {
        try {
            HttpRequest request = request(URI.create(scansUri + "/" + scanId + "/cancel"))
                    .POST(HttpRequest.BodyPublishers.noBody())
                    .build();
            HttpResponse<Void> response = httpClient.send(
                    request, HttpResponse.BodyHandlers.discarding());
            if (response.statusCode() / 100 != 2) {
                throw new ApiException(HttpStatus.BAD_GATEWAY, "SCA_CANCEL_FAILED",
                        "SCA 取消失败，状态码 " + response.statusCode());
            }
        } catch (ApiException exception) {
            throw exception;
        } catch (InterruptedException exception) {
            Thread.currentThread().interrupt();
            throw new ApiException(HttpStatus.BAD_GATEWAY, "SCA_UNAVAILABLE", "SCA 取消请求被中断");
        } catch (Exception exception) {
            throw new ApiException(HttpStatus.BAD_GATEWAY, "SCA_UNAVAILABLE", "SCA 控制服务暂时不可用");
        }
    }

    public ScaResult readResult(Path result) {
        try {
            if (!Files.isRegularFile(result) || Files.size(result) == 0) {
                throw new ApiException(HttpStatus.BAD_GATEWAY, "SCA_RESULT_MISSING",
                        "SCA 未生成有效扫描结果");
            }
            return objectMapper.readValue(result.toFile(), ScaResult.class);
        } catch (ApiException exception) {
            throw exception;
        } catch (Exception exception) {
            throw new ApiException(HttpStatus.BAD_GATEWAY, "SCA_RESULT_INVALID",
                    "SCA 扫描结果无法解析");
        }
    }

    private JobStatus status(UUID scanId) {
        try {
            HttpRequest request = request(URI.create(scansUri + "/" + scanId)).GET().build();
            HttpResponse<byte[]> response = httpClient.send(
                    request, HttpResponse.BodyHandlers.ofByteArray());
            if (response.statusCode() / 100 != 2) {
                throw responseError(response, "SCA_STATUS_FAILED", "无法查询 SCA 扫描状态");
            }
            return objectMapper.readValue(response.body(), JobStatus.class);
        } catch (ApiException exception) {
            throw exception;
        } catch (InterruptedException exception) {
            Thread.currentThread().interrupt();
            throw new ApiException(HttpStatus.BAD_GATEWAY, "SCA_UNAVAILABLE", "SCA 状态查询被中断");
        } catch (Exception exception) {
            throw new ApiException(HttpStatus.BAD_GATEWAY, "SCA_UNAVAILABLE",
                    "SCA 控制服务暂时不可用");
        }
    }

    private HttpRequest.Builder request(URI uri) {
        HttpRequest.Builder builder = HttpRequest.newBuilder(uri).timeout(REQUEST_TIMEOUT);
        if (!serviceToken.isBlank()) {
            builder.header("Authorization", "Bearer " + serviceToken);
        }
        return builder;
    }

    private String blankToDefault(String value, String fallback) {
        return value == null || value.isBlank() ? fallback : value;
    }

    private ApiException responseError(HttpResponse<byte[]> response, String fallbackCode,
                                       String fallbackMessage) {
        String code = fallbackCode;
        String message = fallbackMessage + "，状态码 " + response.statusCode();
        try {
            ApiErrorEnvelope error = objectMapper.readValue(response.body(), ApiErrorEnvelope.class);
            if (error != null && error.error() != null) {
                if (error.error().code() != null && !error.error().code().isBlank()) {
                    code = error.error().code();
                }
                if ("WORKSPACE_NOT_READY".equals(code)) {
                    message = "SCA 无法读取扫描共享目录；请确认 Core 与 sca-server 使用相同的 SCA_WORK_ROOT";
                } else if ("SCAN_CONFLICT".equals(code)) {
                    message = "SCA 中已存在相同 scan_id 但参数不同的任务";
                } else if (error.error().message() != null && !error.error().message().isBlank()) {
                    message = error.error().message();
                    if (message.length() > 500) {
                        message = message.substring(0, 500);
                    }
                }
            }
        } catch (Exception ignored) {
            // Keep the stable fallback when an intermediate proxy returns a non-JSON error.
        }
        return new ApiException(HttpStatus.BAD_GATEWAY, code, message);
    }

    private record ScanSubmission(
            @JsonProperty("contract_version") String contractVersion,
            @JsonProperty("scan_id") String scanId,
            @JsonProperty("source_kind") String sourceKind,
            String profile,
            ScanLimits limits) {
    }

    private record ScanLimits(@JsonProperty("timeout_seconds") int timeoutSeconds) {
    }

    @JsonIgnoreProperties(ignoreUnknown = true)
    private record ApiErrorEnvelope(ApiError error) {
    }

    @JsonIgnoreProperties(ignoreUnknown = true)
    private record ApiError(String code, String message) {
    }

    @JsonIgnoreProperties(ignoreUnknown = true)
    public record JobStatus(
            @JsonProperty("scan_id") String scanId,
            String state,
            @JsonProperty("error_code") String errorCode,
            String message,
            @JsonProperty("result_path") String resultPath) {
    }

    public record ScaResult(
            String engine,
            String engineVersion,
            String sourceMode,
            List<ScaComponent> components,
            List<ScaFinding> findings,
            List<String> diagnostics) {
    }

    public record ScaComponent(
            String ecosystem,
            String groupId,
            String artifactId,
            String version,
            String purl,
            boolean direct,
            String dependencyPath,
            String scope) {
    }

    public record ScaFinding(
            String vulnerabilityId,
            List<String> aliases,
            String title,
            String severity,
            String packageName,
            String affectedVersion,
            String affectedRange,
            List<String> fixedVersions,
            String purl,
            String dependencyPath) {
    }
}
