# Shell (Go)

POSIX-подобная оболочка на Go — pet-проект по мотивам
[CodeCrafters «Build Your Own Shell»](https://app.codecrafters.io/courses/shell/overview).
Интерактивный REPL: разбор командной строки, builtins, внешние программы, пайпы, jobs, history и переменные.

## Требования

- Go 1.27+ (как в `go.mod`)

## Сборка и запуск

Локально (сборка во временный бинарник и сразу запуск):

```sh
./your_program.sh
```

Вручную:

```sh
go build -o myshell ./app
./myshell
```

Тесты:

```sh
go test ./...
```

## Возможности

**Builtins:** `echo`, `pwd`, `cd` (в т.ч. `~`), `type`, `exit`, `jobs`, `history`, `complete`, `declare`

**Внешние команды:** поиск в `PATH`, коды выхода без лишнего шума

**Парсер:** кавычки (`'` / `"`), экранирование `\`, редиректы (`>`, `>>`, `<`, `2>`, `2>>`), фон (`&`), пайпы (`|`)

**Пайплайны:** несколько команд через `|`, builtins в pipe, редиректы на краях

**Фоновые jobs:** `cmd &`, builtin `jobs`, уведомления о завершении (`Done`)

**Автодополнение (Tab):** команды и пути; programmable completion через `complete -C` / `-p` / `-r`

**History:** in-memory список, `history` / `history N`, стрелки ↑/↓, `history -r` / `-w` / `-a`, `HISTFILE` (загрузка при старте, append при выходе)

**Переменные:** `declare name=value`, `declare -p name`, раскрытие `$VAR` / `${VAR}` в аргументах (вне одинарных кавычек)

## Структура

```
app/
  main.go          # точка входа
  shell/           # REPL, readline, исполнение pipeline
  parser/          # токенизация и разбор
  commands/        # builtins и внешние команды
  completion/      # Tab-completion
  history/         # история команд
  jobs/            # фоновые задачи
  vars/            # переменные оболочки
```
