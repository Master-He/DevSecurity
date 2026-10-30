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
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.UUID;
import java.util.concurrent.atomic.AtomicReference;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

class ScaClientTest {
    private HttpServer server;

    @AfterEach
    void stopServer() {
        if (server != null) {
            server.stop(0);
        }
    }

    @Test
    void submitsMetadataPollsStatusAndReadsSharedResult() throws Exception {
        ObjectMapper mapper = new ObjectMapper();
        AtomicReference<String> submission = new AtomicReference<>();
        AtomicReference<String> authorization = new AtomicReference<>();
        AtomicReference<String> cancelMethod = new AtomicReference<>();
        UUID scanId = UUID.randomUUID();
        server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/v1/scans", exchange -> {
            if ("POST".equals(exchange.getRequestMethod())) {
                authorization.set(exchange.getRequestHeaders().getFirst("Authorization"));
                submission.set(new String(exchange.getRequestBody().readAllBytes(), StandardCharsets.UTF_8));
                respond(exchange, 202, "{\"scan_id\":\"" + scanId + "\",\"state\":\"queued\"}");
                return;
            }
            respond(exchange, 404, "{}");
        });
        server.createContext("/v1/scans/" + scanId, exchange ->
                respond(exchange, 200, "{\"scan_id\":\"" + scanId
                        + "\",\"state\":\"succeeded\",\"result_path\":\"scans/"
                        + scanId + "/output/result.json\"}"));
        server.createContext("/v1/scans/" + scanId + "/cancel", exchange -> {
            cancelMethod.set(exchange.getRequestMethod());
            respond(exchange, 202, "{}");
        });
        server.start();

        ScaClient client = new ScaClient(
                mapper,
                "http://127.0.0.1:" + server.getAddress().getPort(),
                "test-token",
                5,
                100);
        client.submit(scanId);
        ScaClient.JobStatus status = client.waitForCompletion(scanId);
        client.cancelAndWait(scanId);

        assertThat(status.state()).isEqualTo("succeeded");
        assertThat(cancelMethod.get()).isEqualTo("POST");
        assertThat(authorization.get()).isEqualTo("Bearer test-token");
        assertThat(submission.get())
                .contains("\"source_kind\":\"zip\"")
                .doesNotContain("source.zip")
                .doesNotContain("zip-content");
    }

    @Test
    void readsResultFromKnownSharedPath(@TempDir Path temporary) throws Exception {
        Path result = temporary.resolve("result.json");
        Files.writeString(result, """
                {
                  "engine":"sca-maven",
                  "engineVersion":"0.2.0",
                  "sourceMode":"POM_FALLBACK",
                  "components":[],
                  "findings":[],
                  "diagnostics":[]
                }
                """);
        ScaClient client = new ScaClient(
                new ObjectMapper(), "http://127.0.0.1:1", "", 5, 100);

        ScaClient.ScaResult value = client.readResult(result);

        assertThat(value.engine()).isEqualTo("sca-maven");
        assertThat(value.components()).isEmpty();
    }

    @Test
    void surfacesWorkspaceNotReadyFromControlService() throws Exception {
        server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/v1/scans", exchange -> respond(exchange, 409, """
                {"error":{"code":"WORKSPACE_NOT_READY","message":"READY marker missing"}}
                """));
        server.start();
        ScaClient client = new ScaClient(
                new ObjectMapper(),
                "http://127.0.0.1:" + server.getAddress().getPort(),
                "test-token",
                5,
                100);

        assertThatThrownBy(() -> client.submit(UUID.randomUUID()))
                .isInstanceOfSatisfying(ApiException.class, exception -> {
                    assertThat(exception.getCode()).isEqualTo("WORKSPACE_NOT_READY");
                    assertThat(exception.getMessage()).contains("相同的 SCA_WORK_ROOT");
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
