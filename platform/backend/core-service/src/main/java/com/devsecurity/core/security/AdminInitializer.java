package com.devsecurity.core.security;

import com.devsecurity.core.domain.AppUser;
import com.devsecurity.core.repository.UserRepository;
import java.security.SecureRandom;
import java.util.Base64;
import java.util.UUID;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.ApplicationArguments;
import org.springframework.boot.ApplicationRunner;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Component;

@Component
public class AdminInitializer implements ApplicationRunner {
    private static final Logger log = LoggerFactory.getLogger(AdminInitializer.class);

    private final UserRepository users;
    private final PasswordEncoder passwordEncoder;
    private final String username;
    private final String password;
    private final String displayName;

    public AdminInitializer(
            UserRepository users,
            PasswordEncoder passwordEncoder,
            @Value("${app.admin.username:admin}") String username,
            @Value("${app.admin.password:}") String password,
            @Value("${app.admin.display-name:平台管理员}") String displayName) {
        this.users = users;
        this.passwordEncoder = passwordEncoder;
        this.username = username;
        this.password = password;
        this.displayName = displayName;
    }

    @Override
    public void run(ApplicationArguments args) {
        if (users.findByUsername(username).isPresent()) {
            return;
        }
        String initialPassword = password;
        if (initialPassword == null || initialPassword.isBlank()) {
            byte[] random = new byte[18];
            new SecureRandom().nextBytes(random);
            initialPassword = Base64.getUrlEncoder().withoutPadding().encodeToString(random);
            log.warn("APP_ADMIN_PASSWORD 未设置，本次生成的临时管理密码: {}", initialPassword);
        }
        try {
            users.save(new AppUser(
                    UUID.randomUUID(), "default", username,
                    passwordEncoder.encode(initialPassword), displayName, "ADMIN"));
            log.info("Created initial administrator '{}'", username);
        } catch (DataIntegrityViolationException ignored) {
            log.info("Administrator '{}' was created by another replica", username);
        }
    }
}
