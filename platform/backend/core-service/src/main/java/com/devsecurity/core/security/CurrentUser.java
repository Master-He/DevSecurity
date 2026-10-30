package com.devsecurity.core.security;

import com.devsecurity.core.api.ApiException;
import com.devsecurity.core.domain.AppUser;
import com.devsecurity.core.repository.UserRepository;
import org.springframework.http.HttpStatus;
import org.springframework.security.core.Authentication;
import org.springframework.stereotype.Component;

@Component
public class CurrentUser {
    private final UserRepository userRepository;

    public CurrentUser(UserRepository userRepository) {
        this.userRepository = userRepository;
    }

    public AppUser require(Authentication authentication) {
        if (authentication == null || !authentication.isAuthenticated()) {
            throw new ApiException(HttpStatus.UNAUTHORIZED, "UNAUTHENTICATED", "请先登录");
        }
        return userRepository.findByUsername(authentication.getName())
                .orElseThrow(() -> new ApiException(
                        HttpStatus.UNAUTHORIZED, "USER_NOT_FOUND", "用户不存在"));
    }

    public boolean isAdmin(AppUser user) {
        return "ADMIN".equals(user.getRole());
    }
}
