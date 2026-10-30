package com.devsecurity.core.api;

import com.devsecurity.core.domain.AppUser;
import com.devsecurity.core.domain.Project;
import com.devsecurity.core.security.CurrentUser;
import com.devsecurity.core.service.ProjectService;
import jakarta.validation.Valid;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;
import java.time.Instant;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/v1/projects")
public class ProjectController {
    private final ProjectService projects;
    private final CurrentUser currentUser;

    public ProjectController(ProjectService projects, CurrentUser currentUser) {
        this.projects = projects;
        this.currentUser = currentUser;
    }

    @GetMapping
    List<Map<String, Object>> list(Authentication authentication) {
        AppUser user = currentUser.require(authentication);
        return projects.list(user).stream().map(ProjectController::view).toList();
    }

    @PostMapping
    Map<String, Object> create(@Valid @RequestBody CreateProjectRequest request,
                               Authentication authentication) {
        AppUser user = currentUser.require(authentication);
        return view(projects.create(user, request.name(), request.description()));
    }

    @GetMapping("/{projectId}")
    Map<String, Object> get(@PathVariable UUID projectId, Authentication authentication) {
        AppUser user = currentUser.require(authentication);
        return view(projects.requireAccess(projectId, user));
    }

    static Map<String, Object> view(Project project) {
        return Map.of(
                "id", project.getId(),
                "name", project.getName(),
                "description", project.getDescription(),
                "ownerUsername", project.getOwnerUsername(),
                "createdAt", project.getCreatedAt()
        );
    }

    public record CreateProjectRequest(
            @NotBlank @Size(max = 120) String name,
            @Size(max = 1000) String description) {
    }
}
