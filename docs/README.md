# Расход топлива автомобилей — круиз-контроль

Учебный проект РИП, транспорт, вариант 29. Карточки автомобилей содержат расход топлива и объём бака. Лабораторная 4 продолжает каталог, ленту, публикацию и лайки из ЛР1–2.

## Лабораторные работы

| Работа | Содержание | Ветка |
|---|---|---|
| ЛР1 | Шаблоны, лента и MinIO | [lab1-fuel_consumptions](../../tree/lab1-fuel_consumptions) |
| ЛР2 | PostgreSQL и ORM | [lab2-fuel_consumptions](../../tree/lab2-fuel_consumptions) |
| ЛР3 | REST API, публикация, лайки | [lab3-fuel_consumptions](../../tree/lab3-fuel_consumptions) |

## Запуск и показ

```sh
docker compose up -d --build
```

API: `http://localhost:8086/api`. База PostgreSQL: localhost:55434; MinIO: localhost:9300. Все порты привязаны к 127.0.0.1. Учебные локальные пароли находятся в Compose. Используются отдельные тома ЛР3; базы предыдущих лабораторных не меняются.

Импортируйте `collections/api.postman_collection.json` или `collections/api.insomnia.json`: в коллекции 10 методов. `demo.*` содержит последовательность показа. Для создания выберите файлы `media/cars_sedan.jpg` и `media/cars_sedan.mp4`; после переноса проекта пути файлов в клиенте нужно выбрать заново. Все URL записаны полностью, без переменных окружения. В примерах используется ID 11. После создания возьмите ID из ответа и явно впишите его числом в полные адреса публикации, лайка и удаления. При повторном показе удалите оставшийся черновик и смените login регистрации.

В ЛР4 вход выдаёт JWT, выход отзывает его через Redis. Каталог и лента публичны, черновик и изменения требуют JWT.

## HTTP методы

Все адреса начинаются с `/api`.

| Метод | URL | Вход и действие |
|---|---|---|
| GET | http://localhost:8086/api/fuel_consumptions | consumption_min, consumption_max: включительный диапазон расхода; только опубликованные |
| GET | http://localhost:8086/api/fuel_consumptions/feed | без id — первая карточка; id=N — карточка; id=N&next=true — следующая с большим id |
| GET | http://localhost:8086/api/fuel_consumptions/draft | единственный черновик текущего пользователя, без ID |
| POST | http://localhost:8086/api/fuel_consumptions | multipart: name, image, video; необязательные description, consumption_rate, fuel_tank_volume; создать черновик |
| PUT | http://localhost:8086/api/fuel_consumptions/11/publication | JSON: description, consumption_rate, fuel_tank_volume; публикация собственного черновика |
| DELETE | http://localhost:8086/api/fuel_consumptions/11 | логическое удаление собственной карточки |
| POST | http://localhost:8086/api/fuel_consumptions/11/likes | JSON liked=1 поставить, liked=0 отменить; только опубликованные |
| POST | http://localhost:8086/api/users | JSON login, full_name, password; регистрация |
| POST | http://localhost:8086/api/users/login | JSON login, password → JWT |
| POST | http://localhost:8086/api/users/logout | Authorization: Bearer JWT → отзыв токена, пустой ответ |

Лента выбирает одну запись непосредственно запросом БД (`ORDER BY id LIMIT 1`). После последней карточки — 404. Фильтрация каталога также выполняется в БД.

## Контракт JSON и переходы

Карточка всегда содержит одинаковые ключи: id, name, description, consumption_rate, fuel_tank_volume, image_url, video_url, creator, created_at, formed_at, is_owner, liked, likes_count. В незаполненном черновике числовые поля и formed_at равны null, description — пустой строке. is_owner и liked имеют значения 0/1. Статус автомобиля остаётся только в БД и не передаётся в JSON.

Переходы: draft → published или deleted; published → deleted. Обратный переход и повторная публикация запрещены. Системные ID, статус, создатель и даты назначает сервер; лишние JSON-поля отклоняются. Создание — 201, успешное чтение/изменение — 200. Нет JWT или он неверен — 403. Redis недоступен для авторизации — 503. Ошибки ввода, отсутствующие ресурсы и запрещённые переходы — 404 с пустым телом. Непредвиденная ошибка хранилища — 500 с пустым телом. HTTP-код не дублируется внутри JSON, текстовых сообщений об ошибках нет.

Файлы действительно загружаются в MinIO, получают случайные латинские имена; в БД сохраняются имена, в ответе — URL. Регистрация хранит bcrypt-хеш; пароль и хеш не возвращаются.

## Таблицы PostgreSQL

| Таблица | Поля |
|---|---|
| users | id BIGINT PK; login VARCHAR(60) UNIQUE NOT NULL; full_name VARCHAR(120) NOT NULL; password_hash TEXT NOT NULL |
| cars | id BIGINT PK; name VARCHAR(120) NOT NULL; description TEXT; status TEXT NOT NULL; consumption_rate NUMERIC(7,2); fuel_tank_volume NUMERIC(7,2); image_name TEXT NOT NULL; video_name TEXT NOT NULL; creator_id BIGINT FK users NOT NULL; created_at TIMESTAMPTZ NOT NULL; formed_at TIMESTAMPTZ |
| car_likes | id BIGINT PK; car_id BIGINT FK cars NOT NULL; user_id BIGINT FK users NOT NULL; UNIQUE(car_id,user_id) |

Частичный уникальный индекс cars(creator_id) WHERE status='draft' гарантирует один черновик на пользователя. Все связи без каскадного удаления. Публикация, удаление и лайки используют транзакции и блокировку строки автомобиля. Повторный like=1 не увеличивает количество, повторный like=0 безопасен.

## Структура и проверка

`internal/model` — три модели; `internal/api` — маршруты, DTO и обработчики; `internal/auth` — JWT и Redis; `internal/storage` — ORM и MinIO. `docs/lab4.mdj` содержит все диаграммы в одном файле StarUML.

```sh
python3 scripts/check_api.py
docker compose exec postgres psql -U student -d fuel_consumptions
```

Проверки создают свои карточки, удаляют их логически и ожидают отсутствие черновика перед запуском.

```sql
SELECT id,name,status,consumption_rate,fuel_tank_volume,creator_id,created_at,formed_at FROM cars ORDER BY id;
SELECT id,car_id,user_id FROM car_likes ORDER BY id;
SELECT id,login,full_name,left(password_hash,4) AS hash_prefix FROM users ORDER BY id;
```

[Задание ЛР3](https://github.com/iu5git/Web#лабораторная-3) · [Методичка Go](https://github.com/iu5git/Web/tree/main/tutorials/lab3-go)
