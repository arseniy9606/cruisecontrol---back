-- Executable in Adminer -> SQL command. No destructive statements.
BEGIN;
INSERT INTO drivers (driver_id, login, full_name)
VALUES (1,'arseniy','Арсений'), (11,'driver11','Водитель 11'), (12,'driver12','Водитель 12'),
(13,'driver13','Водитель 13'), (14,'driver14','Водитель 14'), (15,'driver15','Водитель 15'),
(16,'driver16','Водитель 16'), (17,'driver17','Водитель 17'), (18,'driver18','Водитель 18'),
(19,'driver19','Водитель 19'), (20,'driver20','Водитель 20') ON CONFLICT DO NOTHING;

INSERT INTO cars (car_id,car_name,short_description,
car_status,image_url,video_url,fuel_tank_volume,
consumption_rate,creator_id,formed_at)
VALUES
(1,'Седан','Комфортный седан для городской и загородной дороги. Учебный пример расхода и объёма бака; фактические значения зависят от модели, комплектации и условий движения.','опубликован','http://localhost:9100/cars/cars_sedan.jpg','http://localhost:9100/cars/cars_sedan.mp4',70,6.7,1,CURRENT_TIMESTAMP),
(2,'Внедорожник','Учебный пример Toyota Land Cruiser 2.8 D-4D: 10,7 л/100 км в смешанном цикле WLTP и бак 80 л. Видео иллюстрирует тип кузова, а не конкретную комплектацию.','опубликован','http://localhost:9100/cars/cars_suv.jpg','http://localhost:9100/cars/cars_suv.mp4',80,10.7,1,CURRENT_TIMESTAMP),
(3,'Микроавтобус','Пассажирский микроавтобус для групповых поездок. Расход и объём бака — учебные значения, а не паспортные данные машины на видео.','опубликован','http://localhost:9100/cars/cars_minibus.jpg','http://localhost:9100/cars/cars_minibus.mp4',80,6.6,1,CURRENT_TIMESTAMP),
(4,'Фура','Седельный тягач с полуприцепом. Учебные значения расхода и объёма бака. Расход зависит от массы груза, маршрута и скорости.','опубликован','http://localhost:9100/cars/cars_truck.jpg','http://localhost:9100/cars/cars_truck.mp4',480,28.8,1,CURRENT_TIMESTAMP),
(5,'Внедорожник · черновик','Черновик карточки автомобиля. Заполните описание, расход топлива и объём бака, затем опубликуйте.','черновик','http://localhost:9100/cars/cars_missing.jpg','http://localhost:9100/cars/cars_missing.mp4',80,10.7,1,NULL),
(6,'Архивный седан','Удалённая карточка недоступна в ленте и плитке.','удален','http://localhost:9100/cars/cars_sedan.jpg','http://localhost:9100/cars/cars_sedan.mp4',70,6.7,1,CURRENT_TIMESTAMP),
(7,'Автомобиль без медиа','Пример опубликованной карточки с недоступными файлами: сервер подставляет локальные изображение и видео.','опубликован','http://localhost:9100/cars/not_found.jpg','http://localhost:9100/cars/not_found.mp4',55,7.2,1,CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;

INSERT INTO car_likes(car_id,driver_id)
VALUES (1,11),(1,12),(1,13),(1,14),(1,15),(1,16),(2,11),(2,17),(2,18),
(3,12),(3,13),(3,17),(3,19),(4,11),(4,20),(6,11) ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('drivers','driver_id'),(SELECT MAX(driver_id) FROM drivers));
SELECT setval(pg_get_serial_sequence('cars','car_id'),(SELECT MAX(car_id) FROM cars));
COMMIT;
