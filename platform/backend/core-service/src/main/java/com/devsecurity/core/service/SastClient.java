package com.devsecurity.core.service;

import com.devsecurity.core.api.ApiException;
import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.databind.ObjectMapper;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.List;
import java.util.UUID;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;

@Service
public class SastClient {
    private final ObjectMapper objectMapper;
    private final HttpClient httpClient;
    private final URI analyzeUri;
    private final String serviceToken;
    private final Duration timeout;

    public SastClient(
            ObjectMapper objectMapper,
            @Value("${app.scan.sast-url:http://localhost:8087}") String sastUrl,
            @Value("${app.scan.sast-service-token:}") String serviceToken,
            @Value("${app.scan.execution-timeout-seconds:1800}") long timeoutSeconds) {
        this.objectMapper = objectMapper;
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(10))
                .build();
        this.serviceToken = serviceToken == null ? "" : serviceToken.trim();
        this.timeout = Duration.ofSeconds(Math.max(1, timeoutSeconds));
        this.analyzeUri = URI.create(sastUrl.replaceAll("/+$", "") + "/v1/analyze");
    }

    public SastResult analyze(UUID scanId, byte[] sourceArchive) {
        try {
            HttpRequest.Builder builder = HttpRequest.newBuilder(analyzeUri)
                    .timeout(timeout)
                    .header("Content-Type", "application/zip")
                    .header("X-Scan-Id", scanId.toString())
                    .POST(HttpRequest.BodyPublishers.ofByteArray(sourceArchive));
            if (!serviceToken.isBlank()) {
                builder.header("Authorization", "Bearer " + serviceToken);
            }
            HttpResponse<byte[]> response = httpClient.send(
                    builder.build(), HttpResponse.BodyHandlers.ofByteArray());
            if (response.statusCode() / 100 != 2) {
                throw responseError(response);
            }
            return objectMapper.readValue(response.body(), SastResult.class);
        } catch (ApiException exception) {
            throw exception;
        } catch (InterruptedException exception) {
            Thread.currentThread().interrupt();
            throw new ApiException(HttpStatus.BAD_GATEWAY, "SAST_UNAVAILABLE", "SAST 请求被中断");
        } catch (Exception exception) {
            throw new ApiException(HttpStatus.BAD_GATEWAY, "SAST_RESULT_INVALID",
                    "SAST 服务返回无法解析的结果");
        }
    }

    private ApiException responseError(HttpResponse<byte[]> response) {
        String code = response.statusCode() == 401 ? "SAST_UNAUTHORIZED" : "SAST_FAILED";
        String message = "SAST 引擎执行失败，状态码 " + response.statusCode();
        try {
            ApiErrorEnvelope error = objectMapper.readValue(response.body(), ApiErrorEnvelope.class);
            if (error != null && error.error() != null) {
                if (error.error().code() != null && !error.error().code().isBlank()) {
                    code = error.error().code();
                }
                if (error.error().message() != null && !error.error().message().isBlank()) {
                    message = error.error().message();
                }
            }
        } catch (Exception ignored) {
            // Keep a stable error when an intermediary returns non-JSON output.
        }
        return new ApiException(HttpStatus.BAD_GATEWAY, code,
                message.length() > 500 ? message.substring(0, 500) : message);
    }

    @JsonIgnoreProperties(ignoreUnknown = true)
    private record ApiErrorEnvelope(ApiError error) {
    }

    @JsonIgnoreProperties(ignoreUnknown = true)
    private record ApiError(String code, String message) {
    }

    @JsonIgnoreProperties(ignoreUnknown = true)
    public record SastResult(
            String engine,
            String engineVersion,
            String sourceMode,
            List<SastFinding> findings,
            List<String> diagnostics) {
    }

    @JsonIgnoreProperties(ignoreUnknown = true)
    public record SastFinding(
            String ruleId,
            String title,
            String severity,
            String confidence,
            String category,
            String cwe,
            String message,
            String filePath,
            Integer startLine,
            Integer endLine,
            String codeSnippet,
            String recommendation) {
    }
}
