# Verbose mode

`opener` должен поддерживать флаги:

```bash
opener --verbose <target>
opener -v <target>
```

`--verbose` и `-v` являются эквивалентными.

Verbose mode предназначен для диагностики и должен показывать пользователю **логику принятия решения о том, какое приложение необходимо запустить**.

## Пример

При выполнении:

```bash
opener document.pdf -v
```

или:

```bash
opener -v document.pdf
```

вывод может выглядеть так:

```text
[verbose] target: document.pdf
[verbose] target type: file
[verbose] file type: pdf
[verbose] checking config: ~/.opener.yaml
[verbose] config rule found: open.files.pdf
[verbose] configured application: Google Chrome
[verbose] launch strategy: macOS application
[verbose] command: open -a "Google Chrome" document.pdf
```

После этого приложение запускается.

---

## Пример для директории

```bash
opener -v ./project
```

Вывод:

```text
[verbose] target: ./project
[verbose] target type: directory
[verbose] checking config: ~/.opener.yaml
[verbose] no custom directory rule found
[verbose] using default application: Finder
[verbose] command: open ./project
```

---

## Пример с alias

```bash
opener -v ide ./project
```

Вывод:

```text
[verbose] alias: ide
[verbose] target: ./project
[verbose] checking aliases
[verbose] alias found: ide
[verbose] alias type: application
[verbose] application: Visual Studio Code
[verbose] command: open -a "Visual Studio Code" ./project
```

---

## Пример с CLI-командой

Конфигурация:

```yaml
aliases:
  editor:
    command: "nvim"
```

Вызов:

```bash
opener -v editor README.md
```

Вывод:

```text
[verbose] alias: editor
[verbose] target: README.md
[verbose] checking aliases
[verbose] alias found: editor
[verbose] alias type: command
[verbose] executable: nvim
[verbose] command: nvim README.md
```

---

# Verbose resolution pipeline

Для automatic mode verbose output должен отражать последовательность resolution.

Общая схема:

```text
target
  ↓
target type
  ↓
file type / extension
  ↓
configuration
  ↓
matching rule
  ↓
fallback
  ↓
application / command
  ↓
launch
```

Например:

```text
target
  ↓
file
  ↓
pdf
  ↓
open.files.pdf
  ↓
app = Google Chrome
  ↓
open -a "Google Chrome" document.pdf
```

Для директории:

```text
target
  ↓
directory
  ↓
open.directory
  ↓
no custom rule
  ↓
default macOS application
  ↓
open ./project
```

---

# Verbose output должен быть диагностическим

Verbose mode не должен изменять поведение программы.

Без:

```bash
-v
```

пользователь должен видеть только существенные ошибки.

С:

```bash
-v
```

пользователь получает подробную информацию о каждом этапе resolution.

Verbose output должен направляться в `stderr`, чтобы обычный stdout можно было использовать для машинной обработки в будущем.

---

# CLI parsing

`-v` и `--verbose` должны поддерживаться независимо от положения относительно target:

```bash
opener -v document.pdf
opener --verbose document.pdf
opener document.pdf -v
opener document.pdf --verbose
```

То же должно работать с alias:

```bash
opener -v ide .
opener ide . -v
```

Cobra должна самостоятельно обрабатывать данный флаг.

---

# Архитектурное требование

Не следует добавлять `fmt.Println()` непосредственно в resolver/launcher.

Verbose mode должен быть представлен отдельной опцией контекста выполнения, например:

```go
type Options struct {
    Verbose bool
}
```

Компоненты resolution могут использовать единый logger/diagnostic interface.

Например:

```go
type Logger interface {
    Debug(format string, args ...any)
}
```

При `Verbose == false` debug-сообщения игнорируются.

При `Verbose == true` они выводятся пользователю.

Это позволит в дальнейшем добавить другие уровни логирования без изменения основной логики `opener`.

---

# Definition of Done для verbose mode

Должны работать:

```bash
opener -v document.pdf
opener --verbose document.pdf

opener -v .
opener --verbose .

opener -v ide .
opener ide . --verbose
```

Для каждого вызова пользователь должен видеть:

1. какой target был получен;
2. как был определён его тип;
3. какой file type определён, если target является файлом;
4. какой конфигурационный раздел проверяется;
5. найдено ли соответствующее правило;
6. какой application/command выбран;
7. какой конкретный macOS command будет выполнен.

Например, минимальный ожидаемый trace:

```text
[verbose] target: document.pdf
[verbose] target type: file
[verbose] file type: pdf
[verbose] config rule: open.files.pdf
[verbose] resolved application: Google Chrome
[verbose] launch: open -a "Google Chrome" document.pdf
```
