-- =============================================
-- ROLES AND PERMISSIONS SYSTEM
-- =============================================

-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    is_extensible BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    resource_type VARCHAR(100) NOT NULL, -- exam, problem, submission, user, role, class, etc
    action VARCHAR(50) NOT NULL, -- add, view, update, delete, import, export
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(resource_type, action)
);

-- Map roles to permissions (many-to-many)
CREATE TABLE IF NOT EXISTS role_permissions (
    id SERIAL PRIMARY KEY,
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, permission_id)
);

-- Assign roles to users (many-to-many)
-- User can have multiple roles (e.g., admin and lecturer)
CREATE TABLE IF NOT EXISTS user_roles (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    assigned_by INT REFERENCES users(id), -- admin who assigned this role
    UNIQUE(user_id, role_id)
);

-- Resource-specific access control
-- Track ownership and access level for specific resources
CREATE TABLE IF NOT EXISTS resource_access_control (
    id SERIAL PRIMARY KEY,
    resource_type VARCHAR(100) NOT NULL, -- exam, problem, class, etc
    resource_id BIGINT NOT NULL, -- id of exam, problem, class, etc
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission_type VARCHAR(50) NOT NULL, -- owner, viewer, editor
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    granted_by INT REFERENCES users(id), -- admin or resource owner who granted access
    UNIQUE(resource_type, resource_id, user_id, permission_type)
);

-- Audit trail for permission changes
CREATE TABLE IF NOT EXISTS permission_audit_log (
    id SERIAL PRIMARY KEY,
    action VARCHAR(100) NOT NULL, -- role_assigned, role_revoked, permission_granted, permission_revoked
    target_user_id INT REFERENCES users(id) ON DELETE SET NULL,
    target_role_id INT REFERENCES roles(id) ON DELETE SET NULL,
    target_resource_type VARCHAR(100),
    target_resource_id BIGINT,
    performed_by INT NOT NULL REFERENCES users(id), -- admin who made change
    details JSONB, -- additional context
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_resource_access_control_user_id ON resource_access_control(user_id);
CREATE INDEX idx_resource_access_control_resource ON resource_access_control(resource_type, resource_id);
CREATE INDEX idx_permission_audit_log_user_id ON permission_audit_log(target_user_id);
CREATE INDEX idx_permission_audit_log_created_at ON permission_audit_log(created_at);
CREATE INDEX idx_permissions_resource_type ON permissions(resource_type);

-- =============================================
-- SEED INITIAL ROLES
-- =============================================

INSERT INTO roles (name, description, is_extensible) VALUES 
    ('admin', 'Administrator - Full system access', true),
    ('lecturer', 'Lecturer - Can create/manage exams, view submissions', true),
    ('student', 'Student - Can take exams and view submissions', true)
ON CONFLICT (name) DO NOTHING;

-- =============================================
-- SEED INITIAL PERMISSIONS
-- =============================================

INSERT INTO permissions (resource_type, action, description) VALUES 
    -- Exam permissions
    ('exam', 'add', 'Create new exam'),
    ('exam', 'view', 'View exams'),
    ('exam', 'update', 'Edit exams'),
    ('exam', 'delete', 'Delete exams'),
    ('exam', 'import', 'Import exams'),
    ('exam', 'export', 'Export exams'),
    
    -- Problem permissions
    ('problem', 'add', 'Create new problem'),
    ('problem', 'view', 'View problems'),
    ('problem', 'update', 'Edit problems'),
    ('problem', 'delete', 'Delete problems'),
    ('problem', 'import', 'Import problems'),
    ('problem', 'export', 'Export problems'),
    
    -- Submission permissions
    ('submission', 'add', 'Create submission'),
    ('submission', 'view', 'View submissions'),
    ('submission', 'update', 'Edit submissions'),
    ('submission', 'delete', 'Delete submissions'),
    ('submission', 'export', 'Export submissions'),
    
    -- User permissions
    ('user', 'add', 'Create users'),
    ('user', 'view', 'View users'),
    ('user', 'update', 'Edit users'),
    ('user', 'delete', 'Delete users'),
    
    -- Role permissions
    ('role', 'add', 'Create roles'),
    ('role', 'view', 'View roles'),
    ('role', 'update', 'Edit roles'),
    ('role', 'delete', 'Delete roles'),
    
    -- Permission permissions
    ('permission', 'add', 'Create permissions'),
    ('permission', 'view', 'View permissions'),
    ('permission', 'update', 'Edit permissions'),
    ('permission', 'delete', 'Delete permissions'),
    
    -- Class permissions
    ('class', 'add', 'Create classes'),
    ('class', 'view', 'View classes'),
    ('class', 'update', 'Edit classes'),
    ('class', 'delete', 'Delete classes'),
    
    -- Report/Analytics permissions
    ('report', 'view', 'View reports and analytics')
ON CONFLICT (resource_type, action) DO NOTHING;

-- =============================================
-- ASSIGN PERMISSIONS TO ROLES
-- =============================================

-- Admin: All permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Lecturer: Exam, Problem, Submission (view), Class, Report (view)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'lecturer'
  AND (
    p.resource_type IN ('exam', 'problem', 'class') 
    OR (p.resource_type = 'submission' AND p.action IN ('view', 'export'))
    OR (p.resource_type = 'report' AND p.action = 'view')
  )
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Student: Exam (view), Submission (add, view), Problem (view)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'student'
  AND (
    (p.resource_type = 'exam' AND p.action = 'view')
    OR (p.resource_type = 'submission' AND p.action IN ('add', 'view'))
    OR (p.resource_type = 'problem' AND p.action = 'view')
  )
ON CONFLICT (role_id, permission_id) DO NOTHING;
