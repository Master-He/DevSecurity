package com.devsecurity.core.api;

import jakarta.validation.ConstraintViolationException;
import java.util.Map;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;
import org.springframework.web.multipart.MaxUploadSizeExceededException;

@RestControllerAdvice
public class GlobalExceptionHandler {
    private static final Logger log = LoggerFactory.getLogger(GlobalExceptionHandler.class);

    @ExceptionHandler(ApiException.class)
    ResponseEntity<Map<String, Object>> apiException(ApiException exception) {
        return ResponseEntity.status(exception.getStatus())
                .body(Map.of("error", Map.of(
                        "code", exception.getCode(),
                        "message", exception.getMessage()
                )));
    }

    @ExceptionHandler({MethodArgumentNotValidException.class, ConstraintViolationException.class})
    ResponseEntity<Map<String, Object>> validation(Exception exception) {
        return ResponseEntity.badRequest().body(Map.of("error", Map.of(
                "code", "VALIDATION_ERROR",
                "message", "请求参数不合法"
        )));
    }

    @ExceptionHandler(MaxUploadSizeExceededException.class)
    ResponseEntity<Map<String, Object>> uploadTooLarge() {
        return ResponseEntity.status(HttpStatus.PAYLOAD_TOO_LARGE).body(Map.of("error", Map.of(
                "code", "UPLOAD_TOO_LARGE",
                "message", "源码归档不能超过 100 MB"
        )));
    }

    @ExceptionHandler(Exception.class)
    ResponseEntity<Map<String, Object>> unexpected(Exception exception) {
        log.error("Unhandled API error", exception);
        return ResponseEntity.internalServerError().body(Map.of("error", Map.of(
                "code", "INTERNAL_ERROR",
                "message", "服务暂时不可用，请稍后重试"
        )));
    }
}
