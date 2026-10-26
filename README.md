# Расчёт экономии топлива при использовании круиз контроля

Учебный проект по курсу «Разработка Интернет Приложений», кафедра ИУ5, МГТУ им. Н. Э. Баумана, 2026. Транспорт, вариант 29.

## Лабораторные работы

| № | Содержание | Ветка |
|---|---|---|
| 1 | Шаблонизация и MinIO | [lab1-fuel_consumptions](../../tree/lab1-fuel_consumptions) |
| 2 | PostgreSQL, GORM и курсор | [lab2-fuel_consumptions](../../tree/lab2-fuel_consumptions) |
| 3 | REST API и расчёт экономии | [lab3-fuel_consumptions](../../tree/lab3-fuel_consumptions) |

## Лабораторная работа 3

REST API на Go/Gin с PostgreSQL, GORM и MinIO. Шаблонов и HTML-маршрутов нет. Показ выполняется в Postman или Insomnia. Авторизация относится к ЛР4; в ЛР3 фиксированные создатель и модератор выдаются функцией `service.CurrentUser()` с `sync.Once`.

Автомобиль содержит расход `consumption_rate` (л/100 км) и объём бака `fuel_tank_volume` (л). Домен `fuel_consumptions` управляет заявками на расчёт. В составе заявки у каждой машины задаётся расстояние `distance_km`.

## Запуск

```sh
docker compose up -d --build
```

API: `http://localhost:8085/api`. Для ЛР3 используются отдельные тома, порт PostgreSQL 55433 и MinIO 9200. Перенос данных из ЛР2 не требуется: миграция создаёт 4 таблицы и начальный каталог. Учебные учётные данные заданы в Compose; они предназначены только для локального показа. Порты опубликованы на 127.0.0.1.

Импортируйте `collections/api.postman_collection.json` (16 методов) либо `collections/demo.postman_collection.json` (порядок показа). При загрузке автомобиля выберите `media/cars_sedan.jpg` и `media/cars_sedan.mp4`. Адрес настраивается переменной `base_url`; `request_id` обновляется после добавления автомобиля. Перед удалением прежнего черновика скопируйте его ID из ответа корзины в `old_draft_id`; если корзина пустая, пропустите удаление. Для повторной регистрации используйте новый login.

## Методы API

| Метод | URL после /api | Действие |
|---|---|---|
| GET | /cars | Каталог; consumption_min, consumption_max |
| GET | /cars/:id | Один автомобиль |
| POST | /cars | multipart/form-data: name, description, consumption_rate, fuel_tank_volume, image, video |
| POST | /fuel_consumptions/draft/cars/:carID | Добавить машину; автоматически создать черновик |
| PUT | /fuel_consumptions/:id/cars/:carID | Изменить distance_km |
| DELETE | /fuel_consumptions/:id/cars/:carID | Удалить машину из черновика |
| GET | /fuel_consumptions/cart | ID черновика и количество машин |
| GET | /fuel_consumptions | Список; status, formed_from, formed_to |
| GET | /fuel_consumptions/:id | Заявка с машинами и медиа |
| PUT | /fuel_consumptions/:id | route, saving_percent |
| PUT | /fuel_consumptions/:id/formation | Проверить поля и рассчитать |
| PUT | /fuel_consumptions/:id/completion | decision: complete или reject |
| DELETE | /fuel_consumptions/:id | Логически удалить черновик |
| POST | /users | login, full_name, password |
| POST | /users/login | Заглушка ЛР4 |
| POST | /users/logout | Заглушка ЛР4 |

Диапазон дат задаётся как YYYY-MM-DD, обе границы включены, календарь UTC. Список заявок исключает draft и deleted. Статус заявки — предметное поле, необходимое для фильтрации; он не дублирует HTTP-код. Статус автомобиля из БД в JSON не передаётся.

## Правила и формула

Допустимые переходы: draft → formed или deleted; formed → completed или rejected. Остальные переходы возвращают 404 с пустым телом. Ошибки ввода и отсутствующие ресурсы — 404 без сообщения; успешные чтение и изменение — 200, создание — 201. Непредвиденные ошибки инфраструктуры — 500 с пустым телом. Системные поля не входят в DTO ввода, неизвестные поля JSON отклоняются.

Для расстояния D, расхода R, объёма бака V и заданного в заявке процента S:

- Расход без круиз-контроля: B = D × R / 100.
- Расход с круиз-контролем: C = B × (1 − S / 100).
- Экономия: E = B × S / 100.
- Запас хода: L = V / (R × (1 − S / 100)) × 100.

S задаётся пользователем в диапазоне 0 ≤ S < 100. Это сценарный параметр, а не универсальная фактическая эффективность круиз-контроля. Результаты округляются до двух знаков; итог заявки — сумма округлённых значений экономии её машин. До формирования результаты равны null, ключи JSON остаются теми же. calculated_count считает строки состава с рассчитанной экономией.

Пример: седан 7,2 л/100 км и внедорожник 10,5 л/100 км, по 500 км, S = 10. Экономия 3,60 + 5,25 = 8,85 л. Запас хода седана с баком 50 л — 771,60 км.

## Структура

- `cmd/server` — запуск HTTP-сервера; `cmd/migrate` — миграция и начальные данные.
- `internal/api` — маршруты, обработчики и DTO сериализации.
- `internal/model` — User, Car, FuelConsumption, TripCar.
- `internal/service` — формула и singleton текущего пользователя.
- `internal/storage` — GORM, миграция, S3-клиент MinIO.
- `collections` — две коллекции Postman, также импортируются в Insomnia.

Состав заявки имеет составной ключ (request_id, car_id), отдельного PK связи нет. Транзакции и блокировки защищают изменения заявки; частичный уникальный индекс допускает только один черновик на создателя. Регистрация хранит bcrypt-хеш, пароль и хеш в ответ не включаются. Файлы получают случайные латинские имена, в БД записываются имена, в MinIO — сами файлы.

## Проверка

```sh
docker build --target build -f docker/cars.Dockerfile -t cruisecontrol-lab3-checks .
docker run --rm cruisecontrol-lab3-checks sh -ec 'go test -v ./... && go vet ./...'
python3 scripts/check_api.py
```

HTTP-проверки создают демонстрационные записи и ожидают пустую корзину перед запуском. Для SQL-показа:

```sh
docker compose exec postgres psql -U student -d fuel_consumptions
```

```sql
SELECT id, status, creator_id, moderator_id, formed_at, completed_at, total_saved_liters FROM fuel_consumptions ORDER BY id;
SELECT request_id, car_id, distance_km, saved_liters, range_km FROM trip_cars ORDER BY request_id, car_id;
SELECT id, name, consumption_rate, fuel_tank_volume, image_name, video_name FROM cars ORDER BY id;
SELECT id, login, full_name, left(password_hash, 4) AS hash_prefix FROM users ORDER BY id;
```

[Задание](https://github.com/iu5git/Web#лабораторная-3) · [Методические указания Go](https://github.com/iu5git/Web/tree/main/tutorials/lab3-go).
