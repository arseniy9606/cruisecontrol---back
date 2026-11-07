# Расход топлива автомобилей — круиз-контроль

Учебный проект РИП. Вариант 29, транспорт. Расход топлива и объём бака; фильтрация по расходу.

| Работа | Ветка |
|---|---|
| ЛР1: шаблоны и MinIO | [lab1-fuel_consumptions](../../tree/lab1-fuel_consumptions) |
| ЛР2: PostgreSQL и ORM | [lab2-fuel_consumptions](../../tree/lab2-fuel_consumptions) |
| ЛР3: REST API | [lab3-fuel_consumptions](../../tree/lab3-fuel_consumptions) |
| ЛР4: JWT, Redis blacklist, Swagger | [lab4-fuel_consumptions](../../tree/lab4-fuel_consumptions) |

## Запуск

```sh
cp .env.example .env
# Замените JWT_SECRET в .env результатом команды:
openssl rand -hex 32
docker compose up -d --build
```

[Swagger UI](http://localhost:8086/swagger/index.html), [Swagger JSON](http://localhost:8086/swagger/doc.json).
API: http://localhost:8086/api. PostgreSQL localhost:55434, MinIO http://localhost:9300, консоль http://localhost:9301.
Redis доступен внутри Compose. Все хранилища в Docker, тома ЛР4 отдельные от ЛР1–3.
Учебные пользователи новой базы: creator и moderator, пароль Lab4Demo123. Оба имеют права создателя: отдельной роли модератора в текущем задании нет.

## Авторизация

POST http://localhost:8086/api/users/login принимает login/password, проверяет bcrypt-хеш и возвращает access_token, token_type=Bearer, expires_in в секундах.
В Swagger нажмите Authorize и введите `Bearer ` и токен. В Insomnia/Postman вставьте тот же текст в заголовок Authorization. Токены не записываются в коллекции.

JWT подписан HS256. Проверяются алгоритм, подпись, exp, nbf, iat, issuer, audience, subject и jti. В subject хранится ID пользователя; создатель автомобиля и автор лайка берутся из запроса после проверки JWT, константы пользователя нет. Каждый вход выдаёт отдельный jti.

POST http://localhost:8086/api/users/logout отзывает только текущий JWT. Redis хранит `jwt:blacklist:<SHA256 токена>` со значением ID пользователя и TTL до exp. После выхода этот токен не позволяет изменять данные. Другой токен того же пользователя продолжает работать. AOF и отдельный том сохраняют blacklist после перезапуска Redis.
Публичный каталог и лента доступны с отсутствующим, неверным, просроченным или отозванным токеном; в этих случаях is_owner=0, liked=0. При отказе Redis доступ создателя не предоставляется: защищённый запрос возвращает 503, каталог остаётся доступен гостю.

Регистрация и вход доступны гостю; остальные изменения требуют JWT. Неавторизованный запрос получает 403, неверный логин/пароль — 403. Ошибки имеют пустое тело. Формат JSON карточек из ЛР3 сохранён, статус хранится только в БД.

Подробные [методы и таблицы](docs/README.md), [порядок показа и контрольные вопросы](docs/defense.md), [UML](docs/lab4.mdj).
Коллекции для клиентов находятся в collections; все URL полные, без переменных. Для файла выберите локальные media/cars_sedan.jpg и media/cars_sedan.mp4. Примеры ID замените числом из ответа создания.

## Проверки

```sh
python3 scripts/check_api.py
docker compose exec redis redis-cli --scan --pattern 'jwt:blacklist:*'
docker compose exec redis redis-cli GET <ключ из предыдущей команды>
docker compose exec redis redis-cli TTL <ключ из предыдущей команды>
```

Swagger генерируется из комментариев обработчиков: `swag init -g cmd/server/main.go` (swag v1.16.6).
Сборка: `go test ./...` и `go vet ./...`.

Структура: cmd/server — запуск; cmd/migrate — GORM миграция и учебные данные; internal/auth — JWT/Redis; internal/api — middleware, обработчики, DTO; internal/model — модели; internal/storage — PostgreSQL/MinIO; docs — Swagger и UML. HTML-шаблонов нет.

Основа: [методичка ЛР4](https://github.com/iu5git/Web/tree/main/tutorials/lab4-go), [пример оформления](https://github.com/Dovletov-Seyran/wells-risk-backend/tree/criteria-jwt-auth). ТЗ и РПЗ исключены по заданию.
