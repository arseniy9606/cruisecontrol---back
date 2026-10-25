# Расчёт экономии топлива при использовании круиз-контроля

Учебный проект по курсу «Разработка Интернет Приложений», МГТУ им. Н. Э. Баумана, 2026.

## Предметная область

Транспорт, вариант 29. Каталог автомобилей с объёмом бака и расходом топлива. Лабораторные 1 и 2 реализуют просмотр карточек и управление их состоянием; расчёт экономии развивается в следующих работах.

## Лабораторные работы

| № | Содержание | Ветка |
|---|---|---|
| 1 | Шаблонизация, три страницы, MinIO и лайки | [lab1-fuel_consumptions](https://github.com/arseniy9606/cruisecontrol---back/tree/lab1-fuel_consumptions) |
| 2 | PostgreSQL, GORM, SQL-курсор и ER-диаграмма | [lab2-fuel_consumptions](https://github.com/arseniy9606/cruisecontrol---back/tree/lab2-fuel_consumptions) |

[Задание курса](https://github.com/iu5git/Web/blob/main/README.md). Оформление разделов и навигация по веткам выполнены по [примеру](https://github.com/Dovletov-Seyran/wells-risk-backend/tree/main).

## Лабораторная работа 2

## Структура

- `cmd/cars/main.go` — запуск приложения.
- `internal/api/server.go` — HTTP-маршруты.
- `internal/app/car_handler/handler.go` — обработчики страниц.
- `internal/app/cars/` — модели и работа с данными.
- `templates/cars_feed.html`, `cars_draft.html`, `cars_tiles.html` — лента, добавление, плитка.
- `resources/styles/cars.css` — единый стиль, включая выбор файлов и удаление.
- `media/` — исходные файлы для MinIO; бакет `cars`.

## HTTP-маршруты

| Метод | URL | Назначение |
|---|---|---|
| GET | `/cars` | Плитка; `cars_fuel_min`, `cars_fuel_max` фильтруют расход |
| GET | `/cars/feed/{id}` | Одна карточка; `?next=true` открывает следующую |
| GET | `/cars/draft` | Страница добавления |

| POST | `/cars/create` | Создать черновик через GORM |
| POST | `/cars/publish` | Заполнить поля и опубликовать через GORM |
| POST | `/cars/{id}/delete` | Изменить статус через SQL-курсор |

## Запуск

Нужен Docker Desktop. Команды выполняются из корня репозитория.

```sh
cp .env.example .env
docker compose up -d --build
```

Плитка: `http://localhost:8082/cars`; Adminer: `http://localhost:8083`; MinIO: `http://localhost:9101`.

Первый запуск создаёт отдельные тома `cars` и заполняет пустую БД. Старые тома предыдущей версии автоматически не удаляются и не переносятся. Учебные пароли указаны в `.env.example`.

## База данных

- `drivers` — `driver_id`, уникальный `login`, `full_name`, `password_hash`. Демонстрационные учётные записи отключены маркером `!disabled`; настоящие пароли не хранятся открытым текстом. Авторизация пока не реализована.
- `cars` — ID, название, статус, URL фото и видео, описание, бак, расход, создатель, даты создания и формирования.
- `car_likes` — `like_id`, `car_id`, `driver_id`; уникальная пара автомобиль–водитель. Счётчик вычисляется SQL COUNT. Постановка лайка в ЛР2 не предусмотрена заданием.

Четыре nullable-поля автомобиля: `short_description`, `fuel_tank_volume`, `consumption_rate`, `formed_at`. При публикации первые три проверяются, четвёртое выставляет сервер. `created_at` есть только у автомобиля, как в задании и образце. Каскадного удаления нет.

Удаление: транзакция → DECLARE курсора с FOR UPDATE → FETCH → UPDATE WHERE CURRENT OF → CLOSE → COMMIT. Строка и её лайки сохраняются, статус становится «удален».

Выбранные через проводник фото и видео в ЛР2 не передаются серверу по заданию. В БД создаются неверные URL. `internal/app/cars/media.go` заменяет недоступные медиа на `/cars-assets/defaults/cars_default.jpg` и `.mp4`, обслуживаемые из `resources/defaults` самим SSR-сервером.

## Проверка

```sh
docker compose --profile checks run --rm cars-checks
```

Тесты создают отдельную временную схему PostgreSQL. Проверяются фильтрация, CRUD, лайки из БД, недоступность удалённых карточек, ограничение одного черновика и локальные медиа.

[Отчёт об исправлениях](docs/lab2_corrections.docx) · [ER-проект StarUML](docs/cars_lab2.mdj).
