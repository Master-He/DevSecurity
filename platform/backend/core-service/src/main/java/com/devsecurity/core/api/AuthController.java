package com.devsecurity.core.api;

import com.devsecurity.core.domain.AppUser;
import com.devsecurity.core.repository.UserRepository;
import com.devsecurity.core.security.CurrentUser;
import com.devsecurity.core.security.JwtService;
import jakarta.validation.Valid;
import jakarta.validation.constraints.NotBlank;
import java.util.Map;
import org.springframework.http.HttpStatus;
import org.springframework.security.core.Authentication;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/v1/auth")
public class AuthController {
    private final UserRepository users;
    private final PasswordEncoder passwordEncoder;
    private final JwtService jwtService;
    private final CurrentUser currentUser;

    public AuthController(UserRepository users, PasswordEncoder passwordEncoder,
                          JwtService jwtService, CurrentUser currentUser) {
        this.users = users;
        this.passwordEncoder = passwordEncoder;
        this.jwtService = jwtService;
        this.currentUser = currentUser;
    }

    @PostMapping("/login")
    Map<String, Object> login(@Valid @RequestBody LoginRequest request) {
        AppUser user = users.findByUsername(request.username())
                .filter(candidate -> passwordEncoder.matches(request.password(), candidate.getPasswordHash()))
                .orElseThrow(() -> new ApiException(
                        HttpStatus.UNAUTHORIZED, "INVALID_CREDENTIALS", "用户名或密码错误"));
        return Map.of(
                "token", jwtService.issue(user.getUsername(), user.getRole()),
                "user", userView(user)
        );
    }

    @GetMapping("/me")
    Map<String, Object> me(Authentication authentication) {
        return userView(currentUser.require(authentication));
    }

    private Map<String, Object> userView(AppUser user) {
        return Map.of(
                "id", user.getId(),
                "username", user.getUsername(),
                "displayName", user.getDisplayName(),
                "role", user.getRole()
        );
    }

    public record LoginRequest(@NotBlank String username, @NotBlank String password) {
    }
}
