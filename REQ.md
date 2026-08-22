# opener

## 1. Цель проекта

`opener` — небольшая CLI-утилита для **macOS**, которая предоставляет единый интерфейс для открытия файлов, директорий и приложений.

Основная идея — скрыть различия между конкретными macOS-приложениями и дать простой интерфейс:

```bash
opener <target>
opener <alias> <target>
```

Примеры:

```bash
opener file.pdf
opener ./project
opener ide ./project
```

`opener` не должен самостоятельно реализовывать механизм открытия файлов. В основе необходимо использовать нативные механизмы macOS (`open`, Launch Services), добавляя поверх них удобный слой алиасов и пользовательской конфигурации.

---

## 2. Режимы работы

### 2.1. Открытие файла

```bash
opener file.pdf
```

Утилита определяет тип target и выбирает способ его открытия согласно конфигурации.

В первой версии необходимо поддержать PDF:

```text
.pdf → приложение, заданное в конфигурации
```

Если специального правила нет, используется системный механизм macOS:

```bash
open file.pdf
```

---

### 2.2. Открытие директории

```bash
opener .
opener ./project
opener ~/Downloads
```

Директория должна открываться в стандартном файловом менеджере macOS — Finder.

Для этого использовать:

```bash
open .
```

или:

```bash
open ./project
```

Не следует жестко запускать Finder через отдельный API — `open` является стандартным механизмом macOS для этой задачи.

---

### 2.3. Открытие через алиас

Форма:

```bash
opener <alias> <target>
```

Например:

```bash
opener ide .
```

Конфигурация:

```yaml
aliases:
  ide: code
  editor: nvim
  browser: Google Chrome
```

Тогда:

```bash
opener ide .
```

должно открыть текущую директорию в VS Code.

Для macOS GUI-приложений предпочтительно использовать механизм:

```bash
open -a "Visual Studio Code" .
```

Поэтому конфигурация должна позволять различать CLI-команды и macOS applications.

---

# 3. Конфигурация

Основной конфигурационный файл:

```text
~/.opener.yaml
```

Файл необязателен.

Если его нет, `opener` должен работать с разумными встроенными defaults.

---

# 4. Формат конфигурации

Использовать YAML.

Предлагаемый формат:

```yaml
aliases:
  ide:
    app: "Visual Studio Code"

  browser:
    app: "Google Chrome"

  editor:
    command: "nvim"

open:
  directory:
    command: "open"

  files:
    pdf:
      command: "open"
```

Для macOS-приложений использовать:

```yaml
app: "Application Name"
```

Для CLI-программ использовать:

```yaml
command: "executable"
```

Это позволяет корректно работать как с GUI-приложениями macOS, так и с обычными CLI-программами.

---

# 5. Алиасы

Например:

```yaml
aliases:
  ide:
    app: "Visual Studio Code"

  browser:
    app: "Google Chrome"

  editor:
    command: "nvim"
```

Использование:

```bash
opener ide .
opener browser https://github.com
opener editor README.md
```

Для `app` использовать нативный механизм:

```bash
open -a "Visual Studio Code" .
```

Для `command` использовать:

```text
exec.Command("nvim", "README.md")
```

Алиасы не должны быть ограничены GUI-приложениями.

---

# 6. Автоматическое открытие

Автоматический режим:

```bash
opener <target>
```

должен использовать конфигурацию:

```yaml
open:
  directory:
    command: "open"

  files:
    pdf:
      command: "open"
```

Например:

```bash
opener document.pdf
```

→

```bash
open document.pdf
```

А:

```bash
opener .
```

→

```bash
open .
```

---

# 7. Переопределение поведения

Пользователь должен иметь возможность настроить конкретное приложение.

Например, чтобы PDF всегда открывался в Google Chrome:

```yaml
open:
  files:
    pdf:
      app: "Google Chrome"
```

Тогда:

```bash
opener document.pdf
```

эквивалентно:

```bash
open -a "Google Chrome" document.pdf
```

Чтобы PDF открывался в Safari:

```yaml
open:
  files:
    pdf:
      app: "Safari"
```

---

# 8. Архитектура

Проект реализовать на Go.

Рекомендуемая структура:

```text
opener/
├── cmd/
│   └── opener/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── opener/
│   │   ├── opener.go
│   │   ├── resolver.go
│   │   └── launcher.go
│   └── cli/
│       └── cli.go
├── go.mod
├── go.sum
├── README.md
└── LICENSE
```

Разделить ответственность:

1. CLI parsing
2. configuration loading
3. target resolution
4. alias resolution
5. command/application resolution
6. process launching

---

# 9. Go-библиотеки

Для CLI использовать:

* `github.com/spf13/cobra` — CLI и parsing arguments
* `github.com/spf13/viper` — configuration management и YAML

Для запуска процессов использовать стандартный пакет:

```go
os/exec
```

Не писать собственный YAML parser.

---

# 10. Запуск macOS applications

Для GUI-приложений использовать macOS `open`.

Например конфигурация:

```yaml
aliases:
  ide:
    app: "Visual Studio Code"
```

Вызов:

```bash
opener ide .
```

должен приводить к запуску:

```bash
open -a "Visual Studio Code" .
```

Для PDF:

```yaml
open:
  files:
    pdf:
      app: "Google Chrome"
```

должно выполняться:

```bash
open -a "Google Chrome" document.pdf
```

---

# 11. Запуск CLI-программ

Для:

```yaml
aliases:
  editor:
    command: "nvim"
```

использовать:

```go
exec.Command("nvim", "README.md")
```

Не использовать shell:

```go
exec.Command("sh", "-c", ...)
```

Команды должны запускаться напрямую.

---

# 12. Определение target

Для существующего пути:

```go
os.Stat()
```

Если target является директорией:

```text
directory
```

Если target является файлом:

```text
file
```

В первой версии необходимо специально поддержать:

```text
.pdf
```

Для остальных файлов использовать системный механизм:

```bash
open <target>
```

Таким образом, `opener` не должен самостоятельно реализовывать всю систему ассоциаций файлов macOS.

---

# 13. Fallback

Если для файла нет специального правила:

```bash
opener image.png
```

использовать:

```bash
open image.png
```

macOS самостоятельно определит приложение через Launch Services.

То же относится к URL:

```bash
opener https://github.com
```

Если target не является локальным файлом или директорией, его можно передать системному `open`.

---

# 14. Приоритет интерпретации аргументов

Один аргумент:

```bash
opener <target>
```

означает автоматическое открытие target.

Два и более:

```bash
opener <alias> <target> [target...]
```

означает использование alias.

Например:

```bash
opener ide .
opener ide ./project
opener editor README.md
```

Если alias отсутствует:

```text
unknown alias: ide
```

и ненулевой exit code.

---

# 15. Примеры

### Finder

```bash
opener .
opener ~/Downloads
```

### PDF

```bash
opener document.pdf
```

### VS Code

```yaml
aliases:
  ide:
    app: "Visual Studio Code"
```

```bash
opener ide .
```

### Chrome

```yaml
aliases:
  browser:
    app: "Google Chrome"
```

```bash
opener browser https://github.com
```

### PDF в Chrome

```yaml
open:
  files:
    pdf:
      app: "Google Chrome"
```

```bash
opener document.pdf
```

### Neovim

```yaml
aliases:
  editor:
    command: "nvim"
```

```bash
opener editor README.md
```

---

# 16. Definition of Done

Первая версия считается готовой, если:

```bash
opener document.pdf
```

открывает PDF;

```bash
opener .
```

открывает Finder;

```bash
opener ./project
```

открывает директорию в Finder;

конфигурация:

```yaml
aliases:
  ide:
    app: "Visual Studio Code"
```

позволяет:

```bash
opener ide .
```

открыть директорию в VS Code;

конфигурация:

```yaml
open:
  files:
    pdf:
      app: "Google Chrome"
```

позволяет:

```bash
opener document.pdf
```

открыть PDF в Google Chrome.

Также должны присутствовать:

* `--help`;
* `--version`;
* понятные ошибки;
* ненулевые exit codes;
* unit-тесты config loading;
* unit-тесты alias resolution;
* unit-тесты target resolution;
* README с примерами;
* отсутствие hardcoded конкретных IDE/file managers в коде.

Главный принцип:

> **`opener` — это macOS CLI-обертка над нативным механизмом `open`, дополненная удобными алиасами и пользовательской конфигурацией.**
