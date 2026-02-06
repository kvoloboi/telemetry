CREATE TABLE rooms (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    deleted_at TIMESTAMPTZ NULL
);

CREATE TYPE sensor_type AS ENUM ('V', 'R');

CREATE TABLE sensors (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    room_id BIGINT NOT NULL,
    type sensor_type NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT fk_sensors_room FOREIGN KEY (room_id) REFERENCES rooms(id)
);

CREATE TABLE measurements (
    id BIGSERIAL PRIMARY KEY,
    sensor_id BIGINT NOT NULL,
    measured_at TIMESTAMPTZ NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    CONSTRAINT fk_measurements_sensor FOREIGN KEY (sensor_id) REFERENCES sensors(id)
);

CREATE INDEX idx_rooms_active ON rooms(id)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_sensors_active ON sensors(room_id)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_measurements_sensor_time ON measurements(sensor_id, measured_at);