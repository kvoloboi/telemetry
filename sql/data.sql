-- ============================================================
-- ROOMS
-- ============================================================

INSERT INTO rooms (id, name, deleted_at) VALUES
  (1, 'room_A', NULL),
  (2, 'room_B', NULL);

-- ============================================================
-- SENSORS
-- ============================================================

-- room_A:
-- 1 voltage sensor (V)
-- 2 resistance sensors (R)
INSERT INTO sensors (id, room_id, name, type, deleted_at) VALUES
  (1, 1, 'A_V1', 'V', NULL),
  (2, 1, 'A_R1', 'R', NULL),
  (3, 1, 'A_R2', 'R', NULL);

-- room_B:
-- 2 voltage sensors (V)
-- 3 resistance sensors (R)
INSERT INTO sensors (id, room_id, name, type, deleted_at) VALUES
  (4, 2, 'B_V1', 'V', NULL),
  (5, 2, 'B_V2', 'V', NULL),
  (6, 2, 'B_R1', 'R', NULL),
  (7, 2, 'B_R2', 'R', NULL),
  (8, 2, 'B_R3', 'R', NULL);

-- ============================================================
-- ROOM (A) MEASUREMENTS
-- ============================================================

-- ------------------------------------------------------------
-- Normal operation (V and R present in same second)
-- room_A @ 12:00:00
-- ------------------------------------------------------------

INSERT INTO measurements (sensor_id, measured_at, value) VALUES
  (1, '2026-02-06 12:00:00.100+00', 10.0),   -- V
  (2, '2026-02-06 12:00:00.300+00', 5.0),    -- R
  (3, '2026-02-06 12:00:00.500+00', 5.2);    -- R (avg R = 5.1)

-- ------------------------------------------------------------
-- Multiple measurements per second (per-sensor avg)
-- room_A @ 12:00:01
-- ------------------------------------------------------------

INSERT INTO measurements (sensor_id, measured_at, value) VALUES
  (1, '2026-02-06 12:00:01.100+00', 10.5),
  (1, '2026-02-06 12:00:01.700+00', 11.5),  -- V avg = 11.0
  (2, '2026-02-06 12:00:01.400+00', 5.1),
  (3, '2026-02-06 12:00:01.800+00', 5.3);   -- R avg = 5.2

-- ------------------------------------------------------------
-- Missing R in this second → carry forward R
-- room_A @ 12:00:02
-- ------------------------------------------------------------

INSERT INTO measurements (sensor_id, measured_at, value) VALUES
  (1, '2026-02-06 12:00:02.200+00', 12.0);   -- V only

-- ------------------------------------------------------------
-- R exists but V does NOT → no output row (V-driven)
-- room_A @ 12:00:03
-- ------------------------------------------------------------

INSERT INTO measurements (sensor_id, measured_at, value) VALUES
  (2, '2026-02-06 12:00:03.100+00', 5.4),
  (3, '2026-02-06 12:00:03.600+00', 5.6);

-- ------------------------------------------------------------
-- Clock skew (V and R slightly different seconds)
-- room_A @ 12:00:04 (R slightly earlier)
-- ------------------------------------------------------------

INSERT INTO measurements (sensor_id, measured_at, value) VALUES
  (2, '2026-02-06 12:00:03.900+00', 5.8),  -- R just before
  (1, '2026-02-06 12:00:04.200+00', 12.5); -- V → uses R=5.8


-- ------------------------------------------------------------
-- Long gap, carry-forward from older lookback window
-- room_A @ 12:00:10
-- ------------------------------------------------------------

INSERT INTO measurements (sensor_id, measured_at, value) VALUES
  (1, '2026-02-06 12:00:10.300+00', 13.0);  -- uses last R from 12:00:03.9


-- ============================================================
-- ROOM (B) MEASUREMENTS
-- ============================================================

-- ------------------------------------------------------------
-- room_B multiple V sensors (avg across sensors)
-- room_B @ 12:00:00
-- ------------------------------------------------------------

INSERT INTO measurements (sensor_id, measured_at, value) VALUES
  (4, '2026-02-06 12:00:00.150+00', 20.0),
  (5, '2026-02-06 12:00:00.650+00', 22.0),  -- V avg = 21.0
  (6, '2026-02-06 12:00:00.300+00', 10.0),
  (7, '2026-02-06 12:00:00.500+00', 10.5),
  (8, '2026-02-06 12:00:00.700+00', 9.5);   -- R avg = 10.0

-- ------------------------------------------------------------
-- Missing V entirely for a period → no rows
-- room_B @ 12:00:01
-- ------------------------------------------------------------

INSERT INTO measurements (sensor_id, measured_at, value) VALUES
  (6, '2026-02-06 12:00:01.200+00', 10.2),
  (7, '2026-02-06 12:00:01.600+00', 10.4);

-- ------------------------------------------------------------
-- V returns, R carried forward across multiple seconds
-- room_B @ 12:00:02
-- ------------------------------------------------------------

INSERT INTO measurements (sensor_id, measured_at, value) VALUES
  (4, '2026-02-06 12:00:02.100+00', 21.5),
  (5, '2026-02-06 12:00:02.900+00', 22.5);  -- V avg = 22.0
