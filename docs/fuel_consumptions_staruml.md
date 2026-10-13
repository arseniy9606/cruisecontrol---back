# ER диаграмма в StarUML

Готовый редактируемый проект: `fuel_consumptions_lab2.mdj`. Он открыт и проверен в установленном StarUML 7.1.1. Для запуска используй `ОТКРЫТЬ ДИАГРАММУ.command` в папке «лаб2-РИП» либо File → Open в StarUML. Если показано окно trial, нажми Continue. View → Fit To Window показывает схему целиком; для чтения отдельных полей увеличь масштаб.

## Как создать такую диаграмму самостоятельно

1. Выбери File → New From Template → Data Model. Альтернативно выбери модель в Model Explorer, затем Model → Add Diagram → ER Diagram.
2. В Toolbox → Entity-Relationship выбери Entity и нарисуй три таблицы. Назови их `fuel_consumption_drivers`, `fuel_consumptions`, `fuel_consumption_likes`.
3. Выбери таблицу. Model → Add → Column добавляет столбец; двойной клик открывает QuickEdit. Вводи, например, `fuel_consumption_name: VARCHAR(120)`. Свойства Type и Length можно редактировать в панели Properties.
4. Перенеси все поля из таблиц ниже. Для ID включи Primary Key. Для внешних ключей включи Foreign Key и укажи Reference To соответствующий первичный ключ. Для `fuel_consumption_formed_at` включи Nullable. Для логина — Unique.
5. Добавь три связи One to Many Relationship: водители → карточки; водители → лайки; карточки → лайки. На стороне родителя Cardinality = 1, на стороне дочерней таблицы = 0..*. Identifying выключен: FK не входят в PK дочерней таблицы.
6. Для каждой FK проверь соответствие: `fuel_consumption_creator_id` → driver_id; likes.driver_id → driver_id; likes.fuel_consumption_id → fuel_consumptions.id. Все ID имеют тип BIGINT. BIGSERIAL в PostgreSQL означает BIGINT с последовательностью; поэтому на ER-диаграмме показан фактический BIGINT.
7. Сохрани через File → Save As с расширением `.mdj`. Все диаграммы курса можно хранить в этом одном проекте.
8. Для картинки: File → Export Diagram As → PNG или SVG. Пробная версия StarUML добавляет водяные знаки при экспорте; в комплекте сохранён реальный снимок рабочего окна со схемой, а также отдельная чистая справочная схема, построенная по тем же таблицам.

## Поля таблиц

### fuel_consumption_drivers

- `fuel_consumption_driver_id`: BIGINT, PK.
- `fuel_consumption_driver_login`: VARCHAR(60), UNIQUE, NOT NULL.
- `fuel_consumption_driver_name`: VARCHAR(120), NOT NULL.
- `fuel_consumption_driver_created_at`: TIMESTAMPTZ, NOT NULL.

### fuel_consumptions

- `fuel_consumption_id`: BIGINT, PK.
- `fuel_consumption_name`: VARCHAR(120), NOT NULL.
- `fuel_consumption_description`: VARCHAR(1000), NOT NULL.
- `fuel_consumption_status`: VARCHAR(20), NOT NULL.
- `fuel_consumption_image_url`: VARCHAR(500), NOT NULL.
- `fuel_consumption_video_url`: VARCHAR(500), NOT NULL.
- `fuel_tank_volume`: NUMERIC(7,1), NOT NULL.
- `fuel_consumption_rate`: NUMERIC(4,1), NOT NULL.
- `fuel_consumption_created_at`: TIMESTAMPTZ, NOT NULL.
- `fuel_consumption_creator_id`: BIGINT, FK, NOT NULL.
- `fuel_consumption_formed_at`: TIMESTAMPTZ, NULL допустим.

### fuel_consumption_likes

- `fuel_consumption_like_id`: BIGINT, PK.
- `fuel_consumption_id`: BIGINT, FK, NOT NULL.
- `fuel_consumption_driver_id`: BIGINT, FK, NOT NULL.

## Что объяснить преподавателю

PK однозначно определяет строку; FK связывает существующие записи. Между пользователями и карточками связь лайков M:N, реализованная отдельной таблицей с двумя FK. У водителя также есть отдельная связь авторства 1:N с карточками. UNIQUE пары предотвращает повторные лайки; частичный UNIQUE по автору для статуса «черновик» допускает только один черновик. Все FK используют ON DELETE RESTRICT: физическое удаление родителя со ссылками запрещено. Удаление в интерфейсе меняет статус, а не удаляет строки. Дата формирования NULL до публикации.

Пределы диапазонов, допустимые статусы и частичный индекс подробно показаны в SQL-миграции; одна ER-картинка не заменяет эти ограничения. Источник меню и операций: https://docs.staruml.io/working-with-additional-diagrams/entity-relationship-diagram .
