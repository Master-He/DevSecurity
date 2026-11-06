package com.devsecurity.core.service;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import com.devsecurity.core.api.ApiException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;
import java.io.IOException;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.UUID;
import java.util.concurrent.atomic.AtomicReference;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;

class SastClientTest {
    private HttpServer server;

    @AfterEach
    void stopServer() {
        if (server != null) {
            server.stop(0);
        }
    }

    @Test
    void sendsArchiveWithServiceTokenAndMapsFindings() throws Exception {
        AtomicReference<String> authorization = new AtomicReference<>();
        AtomicReference<String> scanHeader = new AtomicReference<>();
        AtomicReference<byte[]> requestBody = new AtomicReference<>();
        server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/v1/analyze", exchange -> {
            authorization.set(exchange.getRequestHeaders().getFirst("Authorization"));
            scanHeader.set(exchange.getRequestHeaders().getFirst("X-Scan-Id"));
            requestBody.set(exchange.getRequestBody().readAllBytes());
            respond(exchange, 200, """
                    {
                      "engine":"joern",
                      "engineVersion":"4.0.674",
                      "sourceMode":"JOERN_CPGQL",
                      "findings":[{
                        "ruleId":"SAST-JAVA-001",
                        "title":"潜在命令执行",
                        "severity":"HIGH",
                        "confidence":"MEDIUM",
                        "category":"command-injection",
                        "cwe":"CWE-78",
                        "filePath":"Demo.java",
                        "startLine":8,
                        "endLine":8,
                        "codeSnippet":"Runtime.getRuntime().exec(cmd)",
                        "recommendation":"使用固定命令和参数白名单"
                      }],
                      "diagnostics":[]
                    }
                    """);
        });
        server.start();

        UUID scanId = UUID.randomUUID();
        SastClient client = new SastClient(
                new ObjectMapper(),
                "http://127.0.0.1:" + server.getAddress().getPort(),
                "test-token",
                10);
        SastClient.SastResult result = client.analyze(scanId, new byte[]{1, 2, 3});

        assertThat(authorization.get()).isEqualTo("Bearer test-token");
        assertThat(scanHeader.get()).isEqualTo(scanId.toString());
        assertThat(requestBody.get()).containsExactly(1, 2, 3);
        assertThat(result.engineVersion()).isEqualTo("4.0.674");
        assertThat(result.findings()).singleElement()
                .satisfies(finding -> {
                    assertThat(finding.ruleId()).isEqualTo("SAST-JAVA-001");
                    assertThat(finding.startLine()).isEqualTo(8);
                    assertThat(finding.codeSnippet()).contains("Runtime.getRuntime().exec");
                });
    }

    @Test
    void surfacesServiceErrorCode() throws Exception {
        server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/v1/analyze", exchange -> respond(exchange, 502, """
                {"error":{"code":"JOERN_FAILED","message":"Joern CPG 构建失败"}}
                """));
        server.start();
        SastClient client = new SastClient(
                new ObjectMapper(),
                "http://127.0.0.1:" + server.getAddress().getPort(),
                "test-token",
                10);

        assertThatThrownBy(() -> client.analyze(UUID.randomUUID(), new byte[]{1}))
                .isInstanceOfSatisfying(ApiException.class, exception -> {
                    assertThat(exception.getCode()).isEqualTo("JOERN_FAILED");
                    assertThat(exception.getMessage()).contains("Joern");
                });
    }

    private static void respond(HttpExchange exchange, int status, String body) throws IOException {
        byte[] bytes = body.getBytes(StandardCharsets.UTF_8);
        exchange.getResponseHeaders().set("Content-Type", "application/json");
        exchange.sendResponseHeaders(status, bytes.length);
        exchange.getResponseBody().write(bytes);
        exchange.close();
    }
}
