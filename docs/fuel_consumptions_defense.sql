-- Execute individual blocks in Adminer. Read their comments before running.

-- 1. Three tables and current statuses.
SELECT fuel_consumption_id, fuel_consumption_name, fuel_consumption_status,
       fuel_tank_volume, fuel_consumption_rate, fuel_consumption_creator_id,
       fuel_consumption_created_at, fuel_consumption_formed_at
FROM fuel_consumptions ORDER BY fuel_consumption_id;

-- 2. Logical deletion through Adminer. The record and likes remain.
UPDATE fuel_consumptions SET fuel_consumption_status = 'удален'
WHERE fuel_consumption_id = 4;
SELECT fuel_consumption_id, fuel_consumption_name, fuel_consumption_status
FROM fuel_consumptions WHERE fuel_consumption_id = 4;

-- 3. Inspect the newly created draft without guessing its ID.
SELECT fuel_consumption_id, fuel_consumption_name, fuel_consumption_status,
       fuel_consumption_image_url, fuel_consumption_video_url, fuel_consumption_formed_at
FROM fuel_consumptions
WHERE fuel_consumption_creator_id = 1 AND fuel_consumption_status = 'черновик';

-- 4. Inspect creation and publication timestamps after clicking Publish.
SELECT fuel_consumption_id, fuel_consumption_name, fuel_consumption_status,
       fuel_tank_volume, fuel_consumption_rate,
       fuel_consumption_created_at, fuel_consumption_formed_at
FROM fuel_consumptions ORDER BY fuel_consumption_id DESC LIMIT 3;

-- 5. Change subject-specific fields; refresh the feed /fuel_consumptions/feed/1.
UPDATE fuel_consumptions SET fuel_tank_volume = 75, fuel_consumption_rate = 8.4
WHERE fuel_consumption_id = 1;

-- 6. Add a like directly through the junction table. The application has no like POST.
-- If this pair already exists, choose another existing driver who has not liked car 1.
INSERT INTO fuel_consumption_likes (fuel_consumption_id, fuel_consumption_driver_id)
VALUES (1,17) ON CONFLICT DO NOTHING;
SELECT fuel_consumption_id, COUNT(*) AS fuel_consumption_like_count
FROM fuel_consumption_likes GROUP BY fuel_consumption_id ORDER BY fuel_consumption_id;

-- 7. Inspect the three foreign keys: every DELETE rule is RESTRICT.
SELECT tc.table_name, kcu.column_name, ccu.table_name AS referenced_table,
       ccu.column_name AS referenced_column, rc.delete_rule
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON tc.constraint_name = kcu.constraint_name AND tc.constraint_schema = kcu.constraint_schema
JOIN information_schema.constraint_column_usage ccu
  ON tc.constraint_name = ccu.constraint_name AND tc.constraint_schema = ccu.constraint_schema
JOIN information_schema.referential_constraints rc
  ON tc.constraint_name = rc.constraint_name AND tc.constraint_schema = rc.constraint_schema
WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_schema = 'public';

-- 8. The unique partial index enforces at most one draft per driver.
SELECT indexname, indexdef FROM pg_indexes
WHERE schemaname = 'public' AND tablename = 'fuel_consumptions';
