CREATE TABLE IF NOT EXISTS eventos (
    id SERIAL PRIMARY KEY,
    origen VARCHAR(100) NOT NULL,
    tipo VARCHAR(150) NOT NULL,
    severidad VARCHAR(20) NOT NULL,
    fecha TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    atendido BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS alertas (
    id SERIAL PRIMARY KEY,
    evento_id INT NOT NULL,
    regla VARCHAR(150) NOT NULL,
    severidad VARCHAR(20) NOT NULL,
    fecha TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (evento_id) REFERENCES eventos(id)
);

CREATE TABLE IF NOT EXISTS incidentes (
    id SERIAL PRIMARY KEY,
    alerta_id INT NOT NULL,
    descripcion TEXT NOT NULL,
    estado VARCHAR(30) NOT NULL DEFAULT 'Abierto',
    fecha_creacion TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    fecha_cierre TIMESTAMP,
    FOREIGN KEY (alerta_id) REFERENCES alertas(id)
);

INSERT INTO eventos (origen, tipo, severidad, fecha, atendido)
VALUES
    ('PC-RH-03', 'Intento de inicio de sesión fallido', 'Media', '2026-08-18 09:14:00', FALSE),
    ('SERVER-01', 'Dirección IP bloqueada', 'Alta', '2026-08-18 11:32:00', TRUE),
    ('PC-DEV-08', 'Acceso a archivo sensible', 'Alta', '2026-08-18 15:47:00', FALSE),
    ('PC-VENTAS-02', 'Inicio de sesión desde dispositivo desconocido', 'Media', '2026-08-19 08:21:00', FALSE),
    ('FIREWALL-01', 'Múltiples intentos de conexión rechazados', 'Crítica', '2026-08-19 10:56:00', TRUE),
    ('SERVER-DB-01', 'Intento de acceso no autorizado a la base de datos', 'Crítica', '2026-08-19 13:20:00', FALSE),
    ('PC-RH-07', 'Archivo ejecutable descargado desde fuente desconocida', 'Alta', '2026-08-20 09:03:00', FALSE),
    ('VPN-01', 'Inicio de sesión desde ubicación inusual', 'Media', '2026-08-20 11:41:00', FALSE),
    ('SERVER-WEB-02', 'Modificación inesperada de archivo del sistema', 'Crítica', '2026-08-20 14:17:00', FALSE),
    ('PC-DEV-11', 'Dispositivo USB conectado', 'Media', '2026-08-21 08:32:00', TRUE),
    ('FIREWALL-02', 'Escaneo de múltiples puertos detectado', 'Alta', '2026-08-21 12:08:00', FALSE),
    ('SERVER-AUTH-01', 'Múltiples intentos de autenticación privilegiada', 'Crítica', '2026-08-21 16:45:00', FALSE);

INSERT INTO alertas (evento_id, regla, severidad, fecha)
SELECT
    id,
    'Evento de severidad crítica',
    'Crítica',
    fecha
FROM eventos
WHERE severidad = 'Crítica';

INSERT INTO alertas (evento_id, regla, severidad, fecha)
VALUES
    (5, 'Evento de severidad crítica', 'Crítica', '2026-08-19 10:57:00'),
    (6, 'Evento de severidad crítica', 'Crítica', '2026-08-19 13:21:00');

INSERT INTO incidentes (alerta_id, descripcion, estado, fecha_creacion)
SELECT
    a.id,
    CONCAT(
        'Investigar evento crítico en ',
        e.origen,
        ': ',
        e.tipo
    ),
    'Abierto',
    a.fecha
FROM alertas a
JOIN eventos e ON a.evento_id = e.id;

INSERT INTO incidentes (alerta_id, descripcion, estado, fecha_creacion, fecha_cierre)
VALUES
    (1, 'Investigar múltiples intentos de conexión rechazados en FIREWALL-01', 'Cerrado', 
        '2026-08-19 11:00:00', '2026-08-19 13:35:00'),
    (2, 'Investigar intento de acceso no autorizado a la base de datos', 'Cerrado',
        '2026-08-19 13:25:00', '2026-08-19 16:10:00');