# Лабораторная работа 2 Расход топлива автомобилей

Go/Gin + PostgreSQL + GORM + Adminer + MinIO. Тема из ЛР1: транспорт, вариант 29, расчёт экономии топлива при использовании круиз-контроля. На этом этапе реализован каталог автомобилей с расходом топлива и объёмом бака. Расчёт экономии относится к последующим этапам курса.

Ветка: `lab2-fuel_consumptions`. Основа: актуальная `lab1-fuel_consumptions`; первая лабораторная сохранена в своей ветке.

## Запуск на Mac

Требуется установленный Docker Desktop. Go, PostgreSQL, Adminer и MinIO отдельно на Mac устанавливать не требуется: они запускаются в контейнерах.

```sh
cd ~/Desktop/лаб2-РИП/cruisecontrol---back
./scripts/start-fuel_consumptions-lab2.sh
```

Скрипт открывает Docker Desktop при необходимости, создаёт `.env` из `.env.example`, собирает образы, запускает БД и хранилище, применяет миграцию, заполняет только пустую БД и проверяет доступность приложения. Первый запуск на другом компьютере требует интернета; повторная сборка использует кэш. На этом Mac образы уже загружены и собраны.

| Что открыть | Адрес |
|---|---|
| Плитка автомобилей | http://localhost:8082/fuel_consumptions |
| Лента без ID | http://localhost:8082/fuel_consumptions/feed/ |
| Добавление и текущий черновик | http://localhost:8082/fuel_consumptions/draft |
| Adminer | http://localhost:8083 |
| Консоль MinIO | http://localhost:9101 |
| S3 API MinIO | http://localhost:9100 |
| PostgreSQL с компьютера | localhost:55432 |

На рабочем столе есть `ЗАПУСТИТЬ.command`, `ОСТАНОВИТЬ.command`, `ОТКРЫТЬ ПРОЕКТ.command`, `ОТКРЫТЬ ДИАГРАММУ.command` и `ПРОВЕРИТЬ.command`. Их можно открывать двойным нажатием.

Порты ЛР1 (8080, 9000, 9001) и ЛР2 разные. Имена проекта Compose и постоянных томов ЛР2 отдельные. Команды остановки сохраняют все изменения БД и файлы MinIO.

## Вход в Adminer

- Движок: PostgreSQL.
- Сервер: `fuel_consumptions-postgres` — имя контейнера, а не `localhost:55432`.
- Пользователь: `fuel_consumptions_student`.
- Пароль: `fuel_consumptions_lab2_password`.
- База: `fuel_consumptions_lab2`.
- Схема: `public`.

Это учебные локальные реквизиты, задаваемые `.env.example`. Вход в MinIO: `fuel_consumptions_root` / `fuel_consumptions_media_password`; бакет `fuel-consumptions`.

Для ручного заполнения таблиц: войдите в Adminer, нажмите «SQL-запрос», вставьте содержимое `migrations/sql/002_seed.sql`, нажмите «Выполнить». Запросы проверены через настоящий интерфейс Adminer. Повторный запуск файла не перезаписывает уже существующие записи.

## Как пользоваться приложением

**Просмотр.** Плитка показывает только опубликованные автомобили. Диапазон расхода от 0 до 60 л/100 км задаётся полем «От» и ползунком «До». Нажмите «Показать автомобили»: сервер передаст диапазон в SQL через GORM. Значения фильтра сохраняются в форме. Открытие карточки ведёт в ленту, «Следующий автомобиль» выбирает следующую опубликованную запись и возвращается к первой в конце списка.

**Добавление.** В стартовых данных уже есть черновик ID 5. Поэтому сначала откроется форма с кнопкой «Опубликовать». Завершите её, чтобы перейти к созданию новой карточки. Когда черновика нет, открытие страницы ничего не создаёт. Выберите фото и видео, укажите название и нажмите «Далее». Создаётся только черновик с названием, автором, датой создания и статусом. Затем заполните описание, объём бака и расход топлива и нажмите «Опубликовать». Статус меняется на «опубликован», записывается дата формирования. Карточка появляется в плитке и ленте.

**Фото и видео.** По текущему заданию ЛР2 выбранные файлы не передаются серверу и не сохраняются. У file-input намеренно отсутствует `name`, форма имеет обычную кодировку URL-encoded, сервер отвергает multipart. Новому черновику присваиваются неверные URL файлов MinIO. Перед формированием HTML сервер проверяет доступность медиа и подставляет `/fuel_consumptions-assets/defaults/fuel_consumptions_default.jpg` и `.mp4`. Эти файлы обслуживаются самим Go-сервером. Для пустых URL действует та же подстановка. Существующие доступные фото/видео из ЛР1 продолжают загружаться из MinIO.

**Удаление.** Нажмите «Удалить автомобиль» в плитке. Параметризованный `SQL UPDATE` через `database/sql` меняет статус; физического удаления нет. Прямой адрес удалённой карточки возвращает HTTP 404. Лайки и сама строка остаются в БД.

**Лайки.** Интерфейс отображает количество строк в таблице связи. Кнопки постановки лайка и POST лайков в ЛР2 отсутствуют. Изменить счётчик для показа можно добавлением строки через Adminer и обновлением страницы.

## Шесть HTTP методов

Статические CSS, иконки, запасные фото/видео обслуживаются отдельно и не являются дополнительными контроллерами страниц.

| Метод | URL | Что делает | Работа с БД |
|---|---|---|---|
| GET | `/fuel_consumptions` | Плитка с диапазоном `fuel_consumptions_fuel_min/max` | GORM Where + Find |
| GET | `/fuel_consumptions/feed/{id}` или `/fuel_consumptions/feed/` | Одна опубликованная карточка, `?next=true` — следующая | GORM First, LIMIT 1 |
| GET | `/fuel_consumptions/draft` | Единственный черновик водителя 1 либо пустая форма | GORM First |
| POST | `/fuel_consumptions/create` | Создать черновик по `fuel_consumptions_name` | GORM Create |
| POST | `/fuel_consumptions/publish` | Записать описание, бак, расход и опубликовать текущий черновик | GORM Updates |
| POST | `/fuel_consumptions/{id}/delete` | Логически удалить свою карточку | database/sql ExecContext, UPDATE без ORM |

POST перенаправляет на GET с кодом 303. Некорректные поля возвращают 400; отсутствующие/удалённые карточки — 404; попытка второго черновика — 409; отправка файлов — 415. Пользователь фиксирован: `FuelConsumptionCurrentDriverID = 1`. Авторизация в этой работе не реализуется; поля автора, статуса и дат вычисляет сервер.

## Таблицы и поля

`fuel_consumption_drivers` — пользователи-водители:

| Поле | Тип | Ограничение |
|---|---|---|
| fuel_consumption_driver_id | BIGINT, последовательность BIGSERIAL | PK |
| fuel_consumption_driver_login | VARCHAR(60) | NOT NULL, UNIQUE |
| fuel_consumption_driver_name | VARCHAR(120) | NOT NULL |
| fuel_consumption_driver_created_at | TIMESTAMPTZ | NOT NULL |

`fuel_consumptions` — карточки расхода топлива:

| Поле | Тип | Назначение |
|---|---|---|
| fuel_consumption_id | BIGINT, последовательность BIGSERIAL | PK |
| fuel_consumption_name | VARCHAR(120) | Название |
| fuel_consumption_description | VARCHAR(1000) | Краткое описание |
| fuel_consumption_status | VARCHAR(20) | черновик / опубликован / удален |
| fuel_consumption_image_url | VARCHAR(500) | URL фото |
| fuel_consumption_video_url | VARCHAR(500) | URL видео |
| fuel_tank_volume | NUMERIC(7,1) | Объём бака, л |
| fuel_consumption_rate | NUMERIC(4,1) | Расход, л/100 км |
| fuel_consumption_created_at | TIMESTAMPTZ | Дата создания |
| fuel_consumption_creator_id | BIGINT | FK на водителя |
| fuel_consumption_formed_at | TIMESTAMPTZ, NULL допустим | Дата публикации |

`fuel_consumption_likes` — связь многие-ко-многим:

| Поле | Тип | Ограничение |
|---|---|---|
| fuel_consumption_like_id | BIGINT, последовательность BIGSERIAL | PK |
| fuel_consumption_id | BIGINT | FK на карточку |
| fuel_consumption_driver_id | BIGINT | FK на водителя |

Сочетание ID карточки и ID водителя уникально. Все три внешних ключа имеют `ON DELETE RESTRICT`, каскадного удаления нет. Частичный уникальный индекс запрещает два черновика одного водителя, включая одновременные запросы. CHECK ограничивает статусы, диапазоны чисел и полноту опубликованной карточки. NUMERIC хранит предметные числа точно до десятых. TIMESTAMPTZ — дата и время с часовым поясом; серверные часы могут отображаться в UTC.

## Архитектура и расположение кода

- `cmd/fuel_consumptions` — запуск HTTP-сервера.
- `cmd/migrate` — отдельный запуск миграции; `-seed` заполняет только пустую БД.
- `internal/api/fuel_consumptions_server.go` — маршруты и общие проверки POST.
- `internal/app/fuel_consumptions/fuel_consumptions.go` — три модели GORM.
- `internal/app/fuel_consumptions/fuel_consumptions_repository.go` — запросы к PostgreSQL.
- `internal/app/fuel_consumptions/fuel_consumptions_media.go` — проверка доступности медиа и запасные URL.
- `internal/app/fuel_consumptions_handler/fuel_consumptions_handler.go` — шесть контроллеров и данные для шаблонов.
- `templates` — три серверных HTML-шаблона без JavaScript.
- `resources` — CSS, иконки и запасные медиа.
- `media` — исходные фото/видео для загрузки в MinIO; сведения об источниках сохранены.
- `migrations/sql/001_schema.sql` — версионированная миграция схемы, повторное применение безопасно.
- `migrations/sql/002_seed.sql` — начальные данные для Adminer.
- `docs` — отчёт, схема StarUML, контрольные вопросы, порядок защиты и скриншоты.

В ЛР1 источник данных был массивом и JSON-файлом лайков. В ЛР2 PostgreSQL — единственный источник данных приложения. Шаблоны не выполняют SQL; обработчики не фильтруют массив для ленты. Получение карточки и переход к следующей выполняются запросами с LIMIT 1.

## Проверка

```sh
./scripts/check-fuel_consumptions-lab2.sh
```

Go запускается внутри Docker. Проверяются реальные PostgreSQL-запросы, публикация, сохранение неверных URL, фильтр, 404, запрет физического/каскадного удаления связанных данных, уникальность лайков и черновиков, восемь одновременных попыток создания, отказ от файлов и чужого Origin. После тестов выполняется `go vet`.

Тесты создают уникальную временную схему `fuel_consumptions_test_*` и удаляют только её. Обычная схема `public` и твои карточки не очищаются. При запуске `go test` вручную без `FUEL_CONSUMPTIONS_TEST_DATABASE_URL` интеграционный тест пропускается; проверочный скрипт устанавливает её автоматически.

## Остановка и диагностика

```sh
./scripts/stop-fuel_consumptions-lab2.sh
```

Для ручных команд на Mac Docker находится в `/Applications/Docker.app/Contents/Resources/bin/docker`. Если команда `docker` не найдена, добавь этот каталог в PATH:

```sh
export PATH="/Applications/Docker.app/Contents/Resources/bin:$PATH"
docker compose ps -a
docker compose logs --tail=50 fuel_consumptions-backend
docker compose logs --tail=50 fuel_consumptions-migrate
docker compose logs --tail=50 fuel_consumptions-media-init
```

Если порт занят, останови другое приложение на этом порту. Если нет черновика — это нормальная первая стадия добавления. Если черновик существует — сначала опубликуй его. Если медиа отсутствует — показывается локальный запасной файл. Повторный запуск сохраняет состояние; запуск seed на непустой БД не возвращает карточки в исходные статусы.

## Материалы защиты

Порядок: `docs/fuel_consumptions_defense.md`; команды: `docs/fuel_consumptions_defense.sql`; схема: `docs/fuel_consumptions_lab2.mdj`; инструкция по StarUML: `docs/fuel_consumptions_staruml.md`; ответы: `docs/fuel_consumptions_answers.md`; отчёт: `docs/fuel_consumptions_lab2_report.docx` и `.doc`.

На этом Mac дополнительно сохранена опубликованная карточка «Седан для дальних поездок» ID 8, созданная при проверке, и один дополнительный лайк седану. Стартовый черновик ID 5 снова подготовлен для показа. Скриншоты фиксируют последовательные этапы проверки, поэтому их промежуточные статусы могут отличаться от текущего состояния.

Источники: [задание ЛР2](https://github.com/iu5git/Web/blob/main/README.md), [методичка Go](https://github.com/iu5git/Web/tree/main/tutorials/lab2-go), [пример структуры проекта](https://github.com/Dovletov-Seyran/wells-risk-backend/tree/criteria-postgres-gorm). Проверено по актуальному README 13 октября 2026 года. Пример использован как справочный материал; реализация адаптирована под требования текущего задания и свою предметную область.
