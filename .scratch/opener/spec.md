# opener — Spec

## Problem Statement

Пользователь macOS хочет единый CLI-интерфейс для открытия файлов, директорий и приложений, который скрывает различия между конкретными macOS-инструментами (Finder, IDE, браузеры), но не переизобретает файловые ассоциации macOS — они должны оставаться на стороне системы (`open`, Launch Services).

## Solution

`opener` — CLI-утилита на Go, тонкая обёртка над `open`/Launch Services, дополненная YAML-конфигурацией (`~/.opener.yaml`) с алиасами приложений/команд и правилами открытия по типу target (директория, расширение файла). Поддерживает автоматический режим (`opener <target>`) и алиас-режим (`opener <alias> <target>...`), а также verbose-режим для диагностики цепочки принятия решений.

## User Stories

1. Как пользователь macOS, я хочу выполнить `opener document.pdf` и получить PDF, открытый в приложении по умолчанию, чтобы не помнить имя конкретного приложения.
2. Как пользователь, я хочу настроить `open.files.pdf.app: "Google Chrome"` в конфиге, чтобы все PDF всегда открывались в Chrome.
3. Как пользователь, я хочу настроить `open.files.pdf.app: "Safari"`, чтобы переопределить предыдущее правило другим приложением.
4. Как пользователь, я хочу выполнить `opener .`/`opener ./project`/`opener ~/Downloads` и получить директорию, открытую в Finder, без явного упоминания Finder в конфиге или коде.
5. Как пользователь, я хочу задать алиас `ide: {app: "Visual Studio Code"}` и выполнить `opener ide .`, чтобы открыть текущую директорию в VS Code.
6. Как пользователь, я хочу задать алиас `browser: {app: "Google Chrome"}` и выполнить `opener browser https://github.com`, чтобы открыть URL в конкретном браузере.
7. Как пользователь, я хочу задать алиас `editor: {command: "nvim"}` и выполнить `opener editor README.md`, чтобы CLI-редактор запустился напрямую как процесс, без GUI-обёртки.
8. Как пользователь, я хочу передать несколько targets алиасу CLI-команды (`opener editor a.md b.md`), чтобы все файлы попали в один вызов команды.
9. Как пользователь, я хочу открыть файл без специального правила (например, `image.png`) и получить поведение обычного `open image.png`, чтобы Launch Services сама выбрала приложение.
10. Как пользователь, я хочу передать произвольный URL или несуществующий путь (`opener https://github.com`) и чтобы `opener` не падал с ошибкой валидации, а передал target системному `open`.
11. Как пользователь без `~/.opener.yaml`, я хочу, чтобы `opener` работал из коробки с разумным поведением (эквивалент системного `open`), чтобы не быть обязанным писать конфиг.
12. Как пользователь, я хочу получить понятную ошибку `unknown alias: ide` и ненулевой exit code при вызове несуществующего алиаса, чтобы сразу понять причину сбоя.
13. Как пользователь, я хочу вызвать `opener --help`, чтобы увидеть встроенную справку по использованию.
14. Как пользователь, я хочу вызвать `opener --version`, чтобы узнать версию установленной утилиты.
15. Как пользователь, я хочу включить диагностику флагом `-v`/`--verbose` в любой позиции (`opener -v document.pdf`, `opener document.pdf -v`, `opener -v ide .`, `opener ide . --verbose`), чтобы не запоминать конкретный синтаксис.
16. Как пользователь, я хочу видеть в verbose-режиме каждый шаг resolution pipeline (target → тип → file type → проверяемый конфиг-раздел → найденное правило/fallback → выбранное приложение/команда → итоговая macOS-команда), чтобы понимать, почему было выбрано именно это приложение.
17. Как пользователь, я хочу, чтобы verbose-вывод шёл в stderr, а не в stdout, чтобы иметь возможность в будущем парсить stdout машинно без помех от диагностики.
18. Как пользователь, я хочу, чтобы поведение программы было идентично с `-v` и без него (кроме объёма вывода), чтобы диагностика не была источником побочных эффектов.
19. Как разработчик, поддерживающий `opener`, я хочу, чтобы CLI-парсинг, загрузка конфигурации, target resolution, alias resolution, resolution команды/приложения и запуск процесса были разделены на отдельные модули, чтобы можно было развивать/тестировать их независимо.
20. Как разработчик, я хочу единый diagnostic/logger-интерфейс (`Debug(format string, args ...any)`), а не разбросанные `fmt.Println`, чтобы иметь возможность добавить другие уровни логирования в будущем без изменения основной логики.
21. Как разработчик, я хочу, чтобы CLI-команды запускались напрямую через `exec.Command`, без shell (`sh -c`), чтобы избежать shell-инъекций и лишнего слоя интерпретации.
22. Как разработчик, я хочу unit-тесты на config loading, alias resolution и target resolution, чтобы иметь уверенность в корректности resolution-логики без запуска реальных macOS-приложений в CI.

## Implementation Decisions

- **Модуль:** `github.com/inchestnov/opener`, `go 1.26`, лицензия MIT.
- **Структура:** `cmd/opener` (main), `internal/config` (загрузка `~/.opener.yaml`), `internal/opener` (`resolver.go`, `launcher.go`, `opener.go`), `internal/cli` (cobra-команда) — как задано в REQ.md.
- **Конфигурация:** `Config{Aliases map[string]AliasRule; Open OpenConfig{Directory AliasRule; Files map[string]AliasRule}}`, `AliasRule{App, Command string}` — какое из полей задано, та стратегия и используется. Отсутствие `~/.opener.yaml` не является ошибкой — возвращается нулевая `Config{}`, что естественно приводит к fallback-поведению (`open <target>`) без необходимости хардкодить дефолтные YAML-значения.
- **Единственный основной seam для тестирования:** чистая функция `Resolve(args, cfg, logger) (Action, error)` в `internal/opener/resolver.go`. Она не выполняет никакого I/O, кроме `os.Stat` для определения типа target, и возвращает декларативное решение (`Action{Strategy: App|Command|Fallback, Name string, Args []string}`) — без побочных эффектов запуска процессов. Это единственная точка, через которую проходит вся decision-логика (alias resolution + target/extension resolution + config matching + fallback), поэтому она покрывает и alias resolution, и target resolution одним и тем же тестируемым интерфейсом.
- **Вторичный seam:** `LoadConfig` в `internal/config` — parsing `~/.opener.yaml` через viper, принимает путь к файлу (не хардкодит `~`), что позволяет тестировать через `t.TempDir()`.
- **Launcher:** `internal/opener/launcher.go` — тонкая функция, исполняющая `Action` через `os/exec` (`open -a "App" target...` / `exec.Command(command, args...)` / `open target...`). Без интерфейса-абстракции над `exec.Command` — сознательное решение, так как DoD не требует тестов запуска процессов.
- **Разбор аргументов:** 1 позиционный аргумент → автоматический режим; 2+ → алиас-режим, где первый аргумент — имя алиаса, остальные — targets (могут быть несколько). Несуществующий алиас → ошибка `unknown alias: <name>` и ненулевой exit code.
- **Target resolution:** `os.Stat` определяет directory/file; для файлов расширение приводится к нижнему регистру для сопоставления (`.pdf` — единственное специально обрабатываемое расширение в v1). Если `os.Stat` не резолвит путь (URL, несуществующий путь) — target передаётся как непрозрачная строка в `open`/алиас-команду без ошибки resolution.
- **Verbose:** `Options{Verbose bool}`, пробрасывается из cobra-флагов в `Logger` (`Debug(format string, args ...any)`); при `Verbose == false` — no-op реализация; при `true` — пишет в stderr. `resolver.go`/`launcher.go` вызывают `Logger.Debug` на каждом этапе pipeline, в порядке, заданном в verbose.md (target → type → file type → checking config → rule found/fallback → resolved app/command → launch command).
- **CLI:** cobra root command, `-v`/`--verbose` — bool-флаг, независимый от позиции; `--version` — константа `0.1.0`; `--help` — стандартный cobra. `main.go` транслирует ошибки в `os.Exit(1)`.

## Testing Decisions

- Тестируется только внешнее поведение через seam-функции (`Resolve`, `LoadConfig`), а не внутренние детали реализации.
- Только stdlib `testing`, table-driven тесты, без testify (минимизация зависимостей сверх обязательных cobra/viper).
- `internal/config`: тесты `LoadConfig` на отсутствующий/валидный/невалидный YAML через `t.TempDir()`.
- `internal/opener`: тесты `Resolve` на alias resolution (найден/не найден, app-тип/command-тип) и target resolution (директория/файл/расширение pdf/прочее расширение/несуществующий путь) — через in-memory `Config` и реальные файлы/директории в `t.TempDir()`, без запуска процессов.
- `launcher.go` не покрывается unit-тестами — реально запускает OS-процессы, вне DoD.
- Prior art отсутствует — проект greenfield, эти тесты задают паттерн для последующих компонентов.

## Out of Scope

- Специальная обработка любых типов файлов, кроме `.pdf`, в v1 (остальное — fallback на системный `open`).
- Платформы, отличные от macOS.
- Запуск команд через shell (`sh -c`) — только прямой `exec.Command`.
- Интерактивный wizard для генерации конфига.
- Публикация в issue tracker и разбивка на тикеты через `/to-tickets` — недоступны в этом окружении (см. Further Notes).

## Further Notes

- В проекте не настроен issue tracker и triage label vocabulary (`/setup-matt-pocock-skills` не запускался), поэтому спека сохранена как файл в репозитории (`SPEC.md`), а не опубликована в трекер с меткой `ready-for-agent`.
- Запрошенный `+ /to-tickets` skill отсутствует в текущем наборе доступных skills — разбивка на тикеты не выполнена. Нужно запустить `/setup-matt-pocock-skills`, чтобы разблокировать оба шага.
- Локальный git-репозиторий уже инициализирован (без remote — по решению пользователя, GitHub/remote отложены на потом).
