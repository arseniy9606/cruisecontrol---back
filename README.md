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

## Лабораторная работа 1

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

| POST | `/cars/{id}/like` | Поставить или снять собственный лайк |

## Запуск

Нужны Go 1.24+ и Docker Desktop.

```sh
docker compose --profile media up -d --build
go run ./cmd/cars
```

Открыть `http://localhost:8080/cars`. MinIO: `http://localhost:9001`.

ЛР1 использует коллекцию автомобилей с массивами ID пользователей в `CarLikedBy`; количество вычисляется через `len` в обработчике. Лайки работают без JavaScript; локальное сохранение находится в `.runtime/cars_likes.json`. POST лайка оставлен по отдельному замечанию преподавателя. Режим строго трёх GET без изменения лайков: `CARS_STRICT_LAB1=true`.

Добавление содержит выбор фото, видео и название. Все элементы расположены друг под другом. Кнопки «Далее» нет, введённая карточка не сохраняется.

## Проверка

```sh
go test ./...
go vet ./...
```

Отчёт об исправлениях: [ЛР1](docs/lab1_corrections.docx).
