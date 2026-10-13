CREATE TABLE IF NOT EXISTS fuel_consumption_drivers (
    fuel_consumption_driver_id BIGSERIAL PRIMARY KEY,
    fuel_consumption_driver_login VARCHAR(60) NOT NULL UNIQUE,
    fuel_consumption_driver_name VARCHAR(120) NOT NULL,
    fuel_consumption_driver_created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS fuel_consumptions (
    fuel_consumption_id BIGSERIAL PRIMARY KEY,
    fuel_consumption_name VARCHAR(120) NOT NULL CHECK (length(trim(fuel_consumption_name)) > 0),
    fuel_consumption_description VARCHAR(1000) NOT NULL DEFAULT '',
    fuel_consumption_status VARCHAR(20) NOT NULL DEFAULT 'черновик'
        CHECK (fuel_consumption_status IN ('черновик', 'опубликован', 'удален')),
    fuel_consumption_image_url VARCHAR(500) NOT NULL DEFAULT '',
    fuel_consumption_video_url VARCHAR(500) NOT NULL DEFAULT '',
    fuel_tank_volume NUMERIC(7,1) NOT NULL DEFAULT 0 CHECK (fuel_tank_volume BETWEEN 0 AND 1500),
    fuel_consumption_rate NUMERIC(4,1) NOT NULL DEFAULT 0 CHECK (fuel_consumption_rate BETWEEN 0 AND 60),
    fuel_consumption_created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    fuel_consumption_creator_id BIGINT NOT NULL,
    fuel_consumption_formed_at TIMESTAMPTZ,
    CONSTRAINT fuel_consumptions_creator_fk FOREIGN KEY (fuel_consumption_creator_id)
        REFERENCES fuel_consumption_drivers(fuel_consumption_driver_id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT fuel_consumptions_publication_complete CHECK (
        fuel_consumption_status <> 'опубликован' OR
        (length(trim(fuel_consumption_description)) > 0 AND fuel_tank_volume >= 1 AND
         fuel_consumption_rate >= 0.1 AND fuel_consumption_formed_at IS NOT NULL)
    )
);

CREATE TABLE IF NOT EXISTS fuel_consumption_likes (
    fuel_consumption_like_id BIGSERIAL PRIMARY KEY,
    fuel_consumption_id BIGINT NOT NULL,
    fuel_consumption_driver_id BIGINT NOT NULL,
    CONSTRAINT fuel_consumption_likes_car_fk FOREIGN KEY (fuel_consumption_id)
        REFERENCES fuel_consumptions(fuel_consumption_id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT fuel_consumption_likes_driver_fk FOREIGN KEY (fuel_consumption_driver_id)
        REFERENCES fuel_consumption_drivers(fuel_consumption_driver_id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT fuel_consumption_likes_unique UNIQUE (fuel_consumption_id, fuel_consumption_driver_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS fuel_consumptions_one_draft_per_driver
    ON fuel_consumptions(fuel_consumption_creator_id) WHERE fuel_consumption_status = 'черновик';
CREATE INDEX IF NOT EXISTS fuel_consumptions_published_rate
    ON fuel_consumptions(fuel_consumption_rate) WHERE fuel_consumption_status = 'опубликован';
