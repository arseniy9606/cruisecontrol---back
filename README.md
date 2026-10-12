# Лабораторная работа 1 — Расход топлива автомобилей

Транспорт, вариант 29: расчёт экономии топлива при использовании круиз-контроля. ЛР-1 показывает расход и объём бака, без расчёта экономии. Методичка: https://github.com/iu5git/Web/tree/main/tutorials/lab1-go .

## Репозиторий и ветка

Код подготовлен для https://github.com/arseniy9606/cruisecontrol---back/tree/lab1-fuel_consumptions — отдельная ветка по теме, как требует курс. ЛР-1 находится в репозитории backend: серверное приложение Go/Gin с HTML/CSS. React начинается в следующих лабораторных; репозиторий frontend оставлен для этих этапов.

## Запуск на macOS

1. Открыть Docker Desktop и дождаться **Engine running**.
2. В терминале из папки проекта: `./scripts/start-fuel_consumptions-media.sh`.
3. Запустить Go: `./scripts/run-fuel_consumptions.sh` (если Go отсутствует: `./scripts/install-fuel_consumptions-go.sh`).
4. Открыть http://localhost:8080/fuel_consumptions . Лента: /fuel_consumptions/feed/ ; добавление: /fuel_consumptions/draft .

Первый запуск собирает MinIO и mc из официальных исходников и занимает несколько минут. Версии закреплены в docker/minio.Dockerfile. Готовые minio/minio и quay.io/minio/minio с нужным тегом при проверке недоступны. Dockerfile поддерживает текущую архитектуру хоста, включая Apple Silicon.

Консоль MinIO: http://localhost:9001 ; учебный вход **root / rootpassword**. API: localhost:9000, бакет `fuel-consumptions`. Имена объектов fuel_consumptions_*.jpg/mp4. Данные находятся в постоянном томе Docker `fuel_consumptions-media`; ./media — файлы для загрузки. Инициализация создаёт бакет, разрешает чтение и загружает 8 файлов. Порты доступны только на этом Mac.

Остановка MinIO: `docker compose stop`. Повторный запуск: `./scripts/start-fuel_consumptions-media.sh`. Проверка: `docker compose ps`; логи: `docker compose logs --tail=30 fuel_consumptions-minio`. Старая папка MinIO на рабочем столе содержит отдельный запуск без Docker; одновременно оба запуска на портах 9000/9001 не работают.

## Архитектура и данные

Go + Gin + html/template. Frontend — HTML/CSS без JavaScript. Одна коллекция FuelConsumptionRepository: 4 опубликованные записи, 1 черновик, 1 удалённая. Седан, внедорожник, микроавтобус и фура. Предметные поля FuelConsumptionRate и FuelTankVolume. Внедорожник: 10,7 л/100 км и 80 л, учебный пример Toyota Land Cruiser 2.8 D-4D, смешанный WLTP (26,4 UK mpg). Источники: docs/fuel_consumptions_data_sources.md. Съёмка иллюстрирует кузов, не подтверждает комплектацию.

Три GET контроллера: плитка с включительным диапазоном fuel_consumptions_fuel_min/max (0–60), лента с ID и next=true, draft без входного ID. Плитка показывает только расход. Верхние логотипы/подписи удалены. Описание в две строки; «Ещё» через CSS. Добавление — выбор фото/видео, ввод без сохранения, «Далее» отключена. Медиа отображаются с исходными пропорциями: object-fit:contain; фото/видео в добавлении имеют height:auto с ограничением максимальной высоты.

## Лайки и строгое задание ЛР-1

По запросу пользователя доступен POST /fuel_consumptions/:id/like. Повторный клик снимает лайк; cookie обозначает анонимный браузер. ID лежат в FuelConsumptionLikedBy, счётчик len вычисляет контроллер. Мьютекс защищает коллекцию, JSON сохраняется через tempfile/rename в .runtime/fuel_consumptions_likes.json. Ответ 303 перенаправляет на GET. Это учебное голосование, без учётных записей.

Методичка ЛР-1 запрещает изменения данных. Для строгой защиты: `FUEL_CONSUMPTIONS_STRICT_LAB1=true ./scripts/run-fuel_consumptions.sh` — POST отключён, остаются три GET и статические счётчики. Создание, публикация и удаление автомобилей не реализованы в обоих режимах.

## Проверки и сдача

`./scripts/check-fuel_consumptions.sh` — go test ./... и go vet ./... . Проверяются фильтрация/статусы/следующий автомобиль, лайки/сохранение/конкурентность/откат, Origin, строгий режим, отсутствие JS.

Отчёты, скриншоты, локальные редактируемые SVG и схема — docs/. Порядок по методичке: docs/fuel_consumptions_lab1_checklist.md. Облачная Figma не изменена: импортировать SVG самостоятельно. Реальные снимки DevTools Network и ФИО/группу добавить перед защитой. Лицензии: media/FUEL_CONSUMPTIONS_MEDIA_CREDITS.md .
