-- =====================================================
-- Skema PostgreSQL: users, roles, scopes, auth, audit
-- Jalankan di database tujuan (mis. CREATE DATABASE belajar; \c belajar)
-- gen_random_uuid() bawaan PostgreSQL 13+
-- =====================================================

BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- -----------------------------------------------------
-- Tabel master: sites & vendors (dibuat lebih dulu
-- karena direferensikan oleh user_scopes)
-- -----------------------------------------------------
CREATE TABLE sites (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code       varchar(20)  NOT NULL UNIQUE,
    name       varchar(150) NOT NULL,   -- mis. PLTU Paiton
    site_type  varchar(30)              -- mis. PLTU, pabrik, gedung
);

CREATE TABLE vendors (
    id    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code  varchar(20)  NOT NULL UNIQUE,
    name  varchar(150) NOT NULL
);

-- -----------------------------------------------------
-- users
-- -----------------------------------------------------
CREATE TABLE users (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username             varchar(30)  NOT NULL UNIQUE,   -- login, lowercase, tidak bisa diubah
    password_hash        text         NOT NULL,
    full_name            varchar(150),
    email                varchar(150) UNIQUE,            -- opsional, untuk kontak/notifikasi
    phone                varchar(30),
    must_change_password boolean      NOT NULL DEFAULT true,
    failed_login_count   int          NOT NULL DEFAULT 0,
    locked_until         timestamptz,
    is_active            boolean      NOT NULL DEFAULT true,
    last_login_at        timestamptz,
    created_at           timestamptz  NOT NULL DEFAULT now(),
    updated_at           timestamptz  NOT NULL DEFAULT now(),

    CONSTRAINT users_username_lowercase CHECK (username = lower(username))
);

-- Trigger: username tidak bisa diubah
CREATE OR REPLACE FUNCTION prevent_username_change()
RETURNS trigger AS $$
BEGIN
    IF NEW.username IS DISTINCT FROM OLD.username THEN
        RAISE EXCEPTION 'username tidak boleh diubah';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_username_immutable
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION prevent_username_change();

-- Trigger: updated_at otomatis
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- -----------------------------------------------------
-- roles
-- -----------------------------------------------------
CREATE TABLE roles (
    id    smallint PRIMARY KEY,
    code  varchar(30) NOT NULL UNIQUE,  -- SUPER_ADMIN | ADMIN | ADMIN_WORKER | WORKER
    name  varchar(50) NOT NULL,

    CONSTRAINT roles_code_valid
        CHECK (code IN ('SUPER_ADMIN', 'ADMIN', 'ADMIN_WORKER', 'WORKER'))
);

INSERT INTO roles (id, code, name) VALUES
    (1, 'SUPER_ADMIN',  'Super Admin'),
    (2, 'ADMIN',        'Admin'),
    (3, 'ADMIN_WORKER', 'Admin Worker'),
    (4, 'WORKER',       'Worker');

-- -----------------------------------------------------
-- user_roles
-- -----------------------------------------------------
CREATE TABLE user_roles (
    user_id     uuid     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id     smallint NOT NULL REFERENCES roles(id),
    assigned_at timestamptz NOT NULL DEFAULT now(),
    assigned_by uuid     REFERENCES users(id) ON DELETE SET NULL,

    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

-- -----------------------------------------------------
-- user_scopes
-- SUPER_ADMIN tidak butuh baris scope (akses semua)
-- -----------------------------------------------------
CREATE TABLE user_scopes (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    site_id    uuid REFERENCES sites(id)   ON DELETE CASCADE,   -- scope ADMIN
    vendor_id  uuid REFERENCES vendors(id) ON DELETE CASCADE,   -- scope ADMIN_WORKER / WORKER
    created_at timestamptz NOT NULL DEFAULT now(),

    -- minimal salah satu terisi
    CONSTRAINT user_scopes_site_or_vendor
        CHECK (site_id IS NOT NULL OR vendor_id IS NOT NULL)
);

CREATE INDEX idx_user_scopes_user_id   ON user_scopes(user_id);
CREATE INDEX idx_user_scopes_site_id   ON user_scopes(site_id)   WHERE site_id IS NOT NULL;
CREATE INDEX idx_user_scopes_vendor_id ON user_scopes(vendor_id) WHERE vendor_id IS NOT NULL;

-- Cegah scope ganda yang sama untuk satu user
CREATE UNIQUE INDEX uq_user_scopes_site
    ON user_scopes(user_id, site_id) WHERE vendor_id IS NULL;
CREATE UNIQUE INDEX uq_user_scopes_vendor
    ON user_scopes(user_id, vendor_id) WHERE site_id IS NULL;

-- -----------------------------------------------------
-- refresh_tokens
-- -----------------------------------------------------
CREATE TABLE refresh_tokens (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    user_agent text,
    ip_address inet,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user_id    ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- -----------------------------------------------------
-- audit_logs
-- -----------------------------------------------------
CREATE TABLE audit_logs (
    id          bigserial PRIMARY KEY,
    user_id     uuid REFERENCES users(id) ON DELETE SET NULL,  -- kosong untuk login gagal
    action      varchar(30) NOT NULL,   -- LOGIN | LOGOUT | LOGIN_FAILED | CREATE | UPDATE | DELETE
    entity      varchar(50),
    entity_id   uuid,
    before_data jsonb,
    after_data  jsonb,
    ip_address  inet,
    created_at  timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT audit_logs_action_valid
        CHECK (action IN ('LOGIN', 'LOGOUT', 'LOGIN_FAILED', 'CREATE', 'UPDATE', 'DELETE'))
);

CREATE INDEX idx_audit_logs_user_id    ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_entity     ON audit_logs(entity, entity_id);

COMMIT;

