# "SQL-мигратор"
## Технологии

go ~1.23  
Поддерживаемый драйвер БД: `PSQL ^14`

## Общее описание

Аналог инструментов, приведенных в секции "Database schema migration" [awesome-go](https://github.com/avelino/awesome-go).

Утилита работает с миграциями, представленными в виде SQL-файлов.  
Позволяет:  
- генерировать шаблон миграции
- применять миграции
- откатывать миграции

## Конфигурация

### 1)   
Задаётся в yml-файле `configs/config.yml`:
- dsn - строка подключения к БД
- dir - директория для хранения файлов
- table_name - название таблицы в БД

В конфигурации можно использовать переменные окружения, тогда в качестве значения нужно использовать   
специальную нотацию: `${ENV_VAR}` или `$ENV_VAR`.  

Пример:  
`configs/config.yml`

```yml
migrator:
  dsn: $DB_DSN
  dir: ./migrations
  table_name: migrations

logger:
  level: INFO
```

`DB_DSN=SOME_ENV_HERE ./bin/gomigrator -config="configs/config.yml"`

### 2)  
Либо через флаги приложения (см. ниже).

## Использование  
### Как CLI-утилита

Устанавливаем  
```bash
go install github.com/EvgenyRomanov/sql-migrator/cmd/gomigrator@latest
```

#### Выбираем способ конфигурирования

**С помощью файла конфигурации**  

Создаем файл конфигурации в нужном месте, следующего содержания:  
```yml
migrator:
  dsn: postgresql://postgres:postgres@localhost:5432/gomigrator?sslmode=disable
  dir: ./migrations
  table_name: migrations

logger:
  level: INFO
```

и далее для запуска мигратора с нужным файлом конфигурации используем:  
```bash
gomigrator -config="configs/config.yml"
```

По умолчанию файл конфигурации не используется!

**С помощью флагов**

Задаем нужные параметры через флаги конфигурации прилоежния, а именно:
- `dsn` — является обязательным параметром
- `dir` — "./migrations" по умолчанию
- `tableName` — "migrations" по умолчанию

#### Помощь

Чтобы посмотреть как работает приложение можно вызвать команду `gomigrator help`  
```bash
gomigrator help

Usage: gomigrator [OPTIONS] COMMAND [arg...]
  
  You can override varuables from config file by ENV, just use something like "${DB_DSN}"

  OPTIONS:
    -config         Path to configuration file (no default value)
    -dsn            DSN string to database
    -dir            Folder for migrations files ("./migrations" by default)
    -tableName      Name of migrations table ("migrations" by default)  
                
  COMMAND:
    create [name]   Create migration with 'name'
    up              Migrate the DB to the most recent version available
    down            Roll back the version by 1
    redo            Re-run the latest migration
    status          Print all migrations status
    dbversion       Print migrations status (last applied migration)
    help            Print usage
    version         Application version

  Examples:
    gomigrator -config="../configs/config-test.yml" create "create_user_table"
    DB_DSN="postgresql://app:test@pgsql:5432/app" gomigrator up
```

**Создание миграции**

```bash
gomigrator -config="./configs/config.yml" create test_migration

2025-03-17 19:53:44 [INFO] Success create new migration 1742241224843_test_migration.sql
```

Миграция будет создана в директории, указанной в файле конфигурации.  
Шаблон SQL-миграции:  
```sql
-- +gomigrator Up
CREATE TABLE IF NOT EXISTS test (
	id serial NOT NULL,
	test text
);
SELECT * FROM test;

-- +gomigrator Down
DROP TABLE test;
```

Согласно шаблону, инструкции `-- +gomigrator Up` и `-- +gomigrator Down` должны присутствовать в **обязательном** порядке!

**Запуск всех миграций**

```bash
gomigrator -config="./configs/config.yml" up

2025-03-17 19:36:28 [INFO] Migration 20250318000001 successfully applied!
2025-03-17 19:36:28 [INFO] Migration 20250318000002 successfully applied!
```

**Откат последней выполненной миграции**

```bash
gomigrator -config="./configs/config.yml" down

2025-03-17 19:36:28 [INFO] Migration 20250318000002 successfully rollback!
```

**Повтор последней миграции**

```bash
gomigrator -config="./configs/config.yml" redo

2025-03-17 19:36:29 [INFO] Migration 20250318000001 successfully rollback!
2025-03-17 19:36:29 [INFO] Migration 20250318000001 successfully applied!
```

**Вывод статуса миграций**

```bash
gomigrator -config="./configs/config.yml" status

+---+----------------+-----------------------------------+---------------------+
| # |        VERSION | NAME                              | APPLIED AT          |
+---+----------------+-----------------------------------+---------------------+
| 1 | 20250318000001 | 20250318000001_test_migration.sql | 2025-03-17 19:36:29 |
| 2 | 20250318000002 | 20250318000002_test_migration.sql | 2025-03-17 19:36:29 |
+---+----------------+-----------------------------------+---------------------+
|   |          TOTAL | 2                                 |                     |
+---+----------------+-----------------------------------+---------------------+
```

**Вывод версии базы**

```bash
gomigrator -config="./configs/config.yml" dbversion

2025-03-17 19:36:28 [INFO] Current migration version: 20250318000002
```

## Демо-режим  
Для демонстрации работы приложения можно использовать команду из make-файла:

```bash
make run-compose-demo
```

Команда поднимает контейнеры с приложением и БД и выполняет основные команды.  
При этом будут использованы тестовые миграции в директории `build/migrations`,   
а также конфиг по умолчанию из директории `configs/config.yml`.

Задача: [MISSION.md](docs/MISSION.md)
