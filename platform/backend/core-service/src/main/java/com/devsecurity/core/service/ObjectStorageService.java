package com.devsecurity.core.service;

import java.io.IOException;
import java.net.URI;
import java.nio.file.Files;
import java.nio.file.Path;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.stereotype.Service;
import software.amazon.awssdk.auth.credentials.AwsBasicCredentials;
import software.amazon.awssdk.auth.credentials.StaticCredentialsProvider;
import software.amazon.awssdk.core.sync.RequestBody;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.s3.S3Client;
import software.amazon.awssdk.services.s3.model.CreateBucketRequest;
import software.amazon.awssdk.services.s3.model.NoSuchBucketException;

@Service
@ConditionalOnProperty(name = "app.storage.kind", havingValue = "minio", matchIfMissing = true)
public class ObjectStorageService implements ArtifactStorage {
    private final S3Client client;
    private final String bucket;

    public ObjectStorageService(
            @Value("${app.storage.endpoint}") String endpoint,
            @Value("${app.storage.access-key}") String accessKey,
            @Value("${app.storage.secret-key}") String secretKey,
            @Value("${app.storage.bucket}") String bucket,
            @Value("${app.storage.region:us-east-1}") String region) {
        this.bucket = bucket;
        this.client = S3Client.builder()
                .endpointOverride(URI.create(endpoint))
                .region(Region.of(region))
                .credentialsProvider(StaticCredentialsProvider.create(
                        AwsBasicCredentials.create(accessKey, secretKey)))
                .forcePathStyle(true)
                .build();
        ensureBucket();
    }

    private void ensureBucket() {
        try {
            client.headBucket(builder -> builder.bucket(bucket));
        } catch (NoSuchBucketException exception) {
            client.createBucket(CreateBucketRequest.builder().bucket(bucket).build());
        }
    }

    @Override
    public void put(String key, byte[] bytes, String contentType) {
        client.putObject(builder -> builder
                        .bucket(bucket)
                        .key(key)
                        .contentType(contentType)
                        .contentLength((long) bytes.length),
                RequestBody.fromBytes(bytes));
    }

    @Override
    public byte[] get(String key) {
        return client.getObjectAsBytes(builder -> builder.bucket(bucket).key(key)).asByteArray();
    }

    @Override
    public void copyTo(String key, Path target) {
        try {
            Files.createDirectories(target.getParent());
            try (var input = client.getObject(builder -> builder.bucket(bucket).key(key))) {
                Files.copy(input, target);
            }
        } catch (IOException exception) {
            throw new IllegalStateException("Cannot copy object artifact", exception);
        }
    }

    @Override
    public void delete(String key) {
        client.deleteObject(builder -> builder.bucket(bucket).key(key));
    }
}
