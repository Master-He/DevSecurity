package com.devsecurity.core.service;

import com.devsecurity.core.api.ApiException;
import com.devsecurity.core.domain.AppUser;
import com.devsecurity.core.domain.Project;
import com.devsecurity.core.repository.ProjectRepository;
import com.devsecurity.core.security.CurrentUser;
import java.util.List;
import java.util.UUID;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class ProjectService {
    private final ProjectRepository projects;
    private final CurrentUser currentUser;

    public ProjectService(ProjectRepository projects, CurrentUser currentUser) {
        this.projects = projects;
        this.currentUser = currentUser;
    }

    @Transactional
    public Project create(AppUser user, String name, String description) {
        return projects.save(new Project(
                UUID.randomUUID(), user.getTenantId(), name.trim(),
                description == null ? "" : description.trim(), user.getId(), user.getUsername()));
    }

    @Transactional(readOnly = true)
    public List<Project> list(AppUser user) {
        if (currentUser.isAdmin(user)) {
            return projects.findAllByOrderByCreatedAtDesc();
        }
        return projects.findByOwnerIdOrderByCreatedAtDesc(user.getId());
    }

    @Transactional(readOnly = true)
    public Project requireAccess(UUID projectId, AppUser user) {
        Project project = projects.findById(projectId)
                .orElseThrow(() -> new ApiException(
                        HttpStatus.NOT_FOUND, "PROJECT_NOT_FOUND", "项目不存在"));
        if (!currentUser.isAdmin(user) && !project.getOwnerId().equals(user.getId())) {
            throw new ApiException(HttpStatus.NOT_FOUND, "PROJECT_NOT_FOUND", "项目不存在");
        }
        return project;
    }
}
