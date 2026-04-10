-- +goose Up
-- +goose StatementBegin

-- =============================================
-- RBAC+ Dynamic Permissions System
-- =============================================

-- 1. PERMISSIONS: Define granular permissions
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,        -- e.g., "problem:read", "problem:create", "exam:manage"
    description TEXT,                          -- Human-readable description
    category VARCHAR(50),                       -- e.g., "problem", "exam", "user", "admin"
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. ROLES: System roles (admin, lecturer, student, teaching_assistant)
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,          -- e.g., "admin", "lecturer", "student"
    description TEXT,
    is_system BOOLEAN DEFAULT TRUE,            -- System roles cannot be deleted
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. ROLE_PERMISSIONS: Map roles to permissions (many-to-many)
CREATE TABLE role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

-- 4. USER_ROLES: Assign roles to users (many-to-many for multi-role support)
CREATE TABLE user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    assigned_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    UNIQUE(user_id, role_id)
);

-- 5. PERMISSION_GRANTS: Temporary/special permissions for specific resources
CREATE TABLE permission_grants (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    resource_type VARCHAR(50) NOT NULL,       -- e.g., "exam", "problem", "topic"
    resource_id BIGINT NOT NULL,              -- ID of the resource
    permission VARCHAR(50) NOT NULL,          -- e.g., "read", "write", "delete", "manage"
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    granted_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ,                   -- NULL = never expires
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, resource_type, resource_id, permission)
);

-- 6. AUDIT_LOGS: Track permission changes
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,             -- e.g., "role_assigned", "permission_granted", "resource_modified"
    resource_type VARCHAR(50),
    resource_id BIGINT,
    old_value JSONB,
    new_value JSONB,
    reason TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- =============================================
-- INDEXES
-- =============================================

CREATE INDEX idx_permissions_category ON permissions(category);
CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission_id);
CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);
CREATE INDEX idx_permission_grants_user ON permission_grants(user_id);
CREATE INDEX idx_permission_grants_resource ON permission_grants(resource_type, resource_id);
CREATE INDEX idx_permission_grants_expires ON permission_grants(expires_at);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

-- =============================================
-- SEED DATA: Default Roles & Permissions
-- =============================================

-- Insert default permissions
INSERT INTO permissions (name, description, category) VALUES
-- Problem permissions
('problem:read', 'View problems', 'problem'),
('problem:create', 'Create new problems', 'problem'),
('problem:update', 'Update own problems', 'problem'),
('problem:delete', 'Delete own problems', 'problem'),
('problem:publish', 'Publish problems as public', 'problem'),
('problem:manage_all', 'Manage all problems (admin)', 'problem'),

-- Exam permissions
('exam:read', 'View exams', 'exam'),
('exam:create', 'Create new exams', 'exam'),
('exam:update', 'Update own exams', 'exam'),
('exam:delete', 'Delete own exams', 'exam'),
('exam:view_results', 'View exam results', 'exam'),
('exam:manage_participants', 'Add/remove exam participants', 'exam'),
('exam:manage_all', 'Manage all exams (admin)', 'exam'),

-- Topic permissions
('topic:read', 'View topics', 'topic'),
('topic:create', 'Create topics (admin)', 'topic'),
('topic:update', 'Update topics (admin)', 'topic'),
('topic:delete', 'Delete topics (admin)', 'topic'),

-- Submission permissions
('submission:read', 'View own submissions', 'submission'),
('submission:submit', 'Submit solutions', 'submission'),
('submission:grade', 'Grade submissions (lecturer)', 'submission'),
('submission:view_all', 'View all submissions (admin/lecturer)', 'submission'),

-- User management permissions
('user:read', 'View user profiles', 'user'),
('user:update_self', 'Update own profile', 'user'),
('user:update_all', 'Update all user profiles (admin)', 'user'),
('user:delete', 'Delete users (admin)', 'user'),
('user:manage_roles', 'Manage user roles (admin)', 'user'),

-- System permissions
('admin:access', 'Access admin panel', 'admin'),
('admin:manage_permissions', 'Manage permissions (admin)', 'admin'),
('audit:view', 'View audit logs (admin)', 'admin');

-- Insert default roles
INSERT INTO roles (name, description, is_system) VALUES
('admin', 'System administrator - full access', true),
('lecturer', 'Lecturer - create/manage problems and exams', true),
('teaching_assistant', 'Teaching assistant - manage exams and grade', true),
('student', 'Student - submit solutions and take exams', true);

-- Assign permissions to roles
-- ADMIN role - all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin';

-- LECTURER role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'lecturer' AND p.name IN (
    'problem:read', 'problem:create', 'problem:update', 'problem:delete', 'problem:publish',
    'exam:read', 'exam:create', 'exam:update', 'exam:delete', 'exam:view_results', 'exam:manage_participants',
    'topic:read',
    'submission:read', 'submission:grade', 'submission:view_all',
    'user:read', 'user:update_self',
    'audit:view'
);

-- TEACHING_ASSISTANT role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'teaching_assistant' AND p.name IN (
    'problem:read',
    'exam:read', 'exam:manage_participants', 'exam:view_results',
    'topic:read',
    'submission:read', 'submission:grade', 'submission:view_all',
    'user:read', 'user:update_self'
);

-- STUDENT role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'student' AND p.name IN (
    'problem:read',
    'exam:read',
    'topic:read',
    'submission:read', 'submission:submit',
    'user:read', 'user:update_self'
);

-- =============================================
-- MIGRATION: Assign existing users to roles based on their role column
-- =============================================

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON r.name = u.role
ON CONFLICT (user_id, role_id) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS permission_grants;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;

-- +goose StatementEnd
