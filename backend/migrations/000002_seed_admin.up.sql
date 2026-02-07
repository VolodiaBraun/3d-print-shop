INSERT INTO users (email, password_hash, first_name, role, is_active)
VALUES ('admin@3dprint.shop', '$2a$10$ROlEfsm1gjUgGjSeI0faUOVhXku.gTiDuNMMs0YLSVLkhVToVi2.S', 'Admin', 'admin', true)
ON CONFLICT DO NOTHING;
