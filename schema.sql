CREATE TABLE IF NOT EXISTS eventos (
    id SERIAL PRIMARY KEY,
    origen VARCHAR(100) NOT NULL,
    tipo VARCHAR(150) NOT NULL,
    severidad VARCHAR(20) NOT NULL,
    fecha TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    atendido BOOLEAN NOT NULL DEFAULT FALSE
);

INSERT INTO eventos (origen, tipo, severidad, fecha, atendido)
VALUES
    ('PC-RH-03', 'Intento de inicio de sesión fallido', 'Media', '2026-08-18 09:14:00', FALSE),
    ('SERVER-01', 'Dirección IP bloqueada', 'Alta', '2026-08-18 11:32:00', TRUE),
    ('PC-DEV-08', 'Acceso a archivo sensible', 'Alta', '2026-08-18 15:47:00', FALSE),
    ('PC-VENTAS-02', 'Inicio de sesión desde dispositivo desconocido', 'Media', '2026-08-19 08:21:00', FALSE),
    ('FIREWALL-01', 'Múltiples intentos de conexión rechazados', 'Crítica', '2026-08-19 10:56:00', TRUE);