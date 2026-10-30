package com.devsecurity.core.repository;

import com.devsecurity.core.domain.Project;
import java.util.List;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;

public interface ProjectRepository extends JpaRepository<Project, UUID> {
    List<Project> findByOwnerIdOrderByCreatedAtDesc(UUID ownerId);
    List<Project> findAllByOrderByCreatedAtDesc();
}
