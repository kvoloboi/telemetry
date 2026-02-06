-- Parameters
WITH params AS (
    SELECT
        '2026-02-06 12:00:00+00'::TIMESTAMPTZ - INTERVAL '1 hour' AS from_ts,
        '2026-02-06 13:00:00+00'::TIMESTAMPTZ AS to_ts
),

-- Step 1: per-sensor, per-second aggregation
sensor_second AS (
    SELECT
        r.id AS room_id,
        r.name AS room,
        s.type AS sensor_type,
        date_trunc('second', m.measured_at) AS ts,
        AVG(m.value) AS sensor_avg
    FROM measurements m
    JOIN sensors s ON s.id = m.sensor_id
    JOIN rooms r   ON r.id = s.room_id
    CROSS JOIN params p
    WHERE r.deleted_at IS NULL
      AND s.deleted_at IS NULL
      -- Lookback to include previous values for carry-forward
      AND m.measured_at >= p.from_ts - INTERVAL '1 hour'
      AND m.measured_at <  p.to_ts
    GROUP BY r.id, r.name, s.type, ts
),

-- Step 2: average across sensors of same type
room_second AS (
    SELECT
        room_id,
        room,
        sensor_type,
        ts,
        AVG(sensor_avg) AS avg_value
    FROM sensor_second
    GROUP BY room_id, room, sensor_type, ts
),

timeline AS (
    SELECT DISTINCT room_id, room, ts
    FROM room_second
    WHERE sensor_type = 'V'
),

-- Step 4: pivot V and R onto the timeline
aligned AS (
    SELECT
        t.room,
        t.room_id,
        t.ts,
        MAX(CASE WHEN rs.sensor_type = 'V' THEN rs.avg_value END) AS V,
        MAX(CASE WHEN rs.sensor_type = 'R' THEN rs.avg_value END) AS R
    FROM timeline t
    LEFT JOIN room_second rs
        ON rs.room_id = t.room_id
       AND rs.ts = t.ts
    GROUP BY t.room, t.room_id, t.ts
),

-- Step 5: carry-forward missing values
filled AS (
    SELECT
        room,
        ts,
        V,
        MAX(R) OVER (
            PARTITION BY room
            ORDER BY ts
            ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
        ) AS R
    FROM aligned
)
SELECT
    room,
    ts AS timestamp,
    ROUND(V::numeric, 2) AS V,
    ROUND(R::numeric, 2) AS R,
    ROUND((V / R)::numeric, 3) AS I
FROM filled, params p
WHERE V IS NOT NULL
    AND R IS NOT NULL
    AND ts >= p.from_ts
    AND ts < p.to_ts
ORDER BY room, ts;
