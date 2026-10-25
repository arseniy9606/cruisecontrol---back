CREATE TABLE IF NOT EXISTS drivers (
    driver_id BIGSERIAL PRIMARY KEY,
    login VARCHAR(60) NOT NULL UNIQUE,
    full_name VARCHAR(120) NOT NULL,
    password_hash VARCHAR(100) NOT NULL DEFAULT '!disabled'
);

CREATE TABLE IF NOT EXISTS cars (
    car_id BIGSERIAL PRIMARY KEY,
    car_name VARCHAR(120) NOT NULL CHECK (length(trim(car_name)) > 0),
    short_description VARCHAR(1000),
    car_status VARCHAR(20) NOT NULL DEFAULT 'черновик'
        CHECK (car_status IN ('черновик', 'опубликован', 'удален')),
    image_url VARCHAR(500) NOT NULL DEFAULT '',
    video_url VARCHAR(500) NOT NULL DEFAULT '',
    fuel_tank_volume NUMERIC(7,1) CHECK (fuel_tank_volume BETWEEN 0 AND 1500),
    consumption_rate NUMERIC(4,1) CHECK (consumption_rate BETWEEN 0 AND 60),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    creator_id BIGINT NOT NULL,
    formed_at TIMESTAMPTZ,
    CONSTRAINT cars_creator_fk FOREIGN KEY (creator_id)
        REFERENCES drivers(driver_id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT cars_publication_complete CHECK (
        car_status <> 'опубликован' OR
        (short_description IS NOT NULL AND fuel_tank_volume IS NOT NULL AND consumption_rate IS NOT NULL AND length(trim(short_description)) > 0 AND fuel_tank_volume >= 1 AND
         consumption_rate >= 0.1 AND formed_at IS NOT NULL)
    )
);

CREATE TABLE IF NOT EXISTS car_likes (
    like_id BIGSERIAL PRIMARY KEY,
    car_id BIGINT NOT NULL,
    driver_id BIGINT NOT NULL,
    CONSTRAINT car_likes_car_fk FOREIGN KEY (car_id)
        REFERENCES cars(car_id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT car_likes_driver_fk FOREIGN KEY (driver_id)
        REFERENCES drivers(driver_id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT car_likes_unique UNIQUE (car_id, driver_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS cars_one_draft_per_driver
    ON cars(creator_id) WHERE car_status = 'черновик';
CREATE INDEX IF NOT EXISTS cars_published_rate
    ON cars(consumption_rate) WHERE car_status = 'опубликован';
