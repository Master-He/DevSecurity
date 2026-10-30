package com.devsecurity.core;

import java.util.concurrent.Executor;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;
import org.springframework.scheduling.annotation.EnableAsync;
import org.springframework.scheduling.annotation.EnableScheduling;
import org.springframework.scheduling.concurrent.ThreadPoolTaskExecutor;

@EnableAsync
@EnableScheduling
@SpringBootApplication
public class CoreServiceApplication {

    public static void main(String[] args) {
        // 启动需要配置环境变量： APP_ADMIN_PASSWORD=admin123;JWT_SECRET=01234567890123456789012345678912;SPRING_PROFILES_ACTIVE=local
        SpringApplication.run(CoreServiceApplication.class, args);
    }

    @Bean("scanExecutor")
    Executor scanExecutor() {
        ThreadPoolTaskExecutor executor = new ThreadPoolTaskExecutor();
        executor.setCorePoolSize(4);
        executor.setMaxPoolSize(8);
        executor.setQueueCapacity(100);
        executor.setThreadNamePrefix("scan-worker-");
        executor.initialize();
        return executor;
    }
}
