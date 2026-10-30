package com.devsecurity.core.security;

import static org.junit.jupiter.api.Assertions.assertEquals;

import org.junit.jupiter.api.Test;

class JwtServiceTest {
    @Test
    void issuedTokenCarriesSubjectAndRole() {
        JwtService service = new JwtService(
                "0123456789012345678901234567890123456789", 1);
        String token = service.issue("alice", "ADMIN");

        var claims = service.parse(token);
        assertEquals("alice", claims.getSubject());
        assertEquals("ADMIN", claims.get("role", String.class));
    }
}
