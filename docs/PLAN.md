# PLAN — `tinywasm/update`: completar la librería de self-update

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Zero-context agent: todo lo necesario está aquí.

**La mayor parte del código ya fue MOVIDO (no inventes implementaciones; ajusta lo que hay).**
Estas funciones, portadas de código probado, ya existen y compilan (con tests verdes):

| Archivo | Origen | Función |
|---|---|---|
| `checksum.go` | installer `verifyChecksum` | `VerifyChecksum(asset string, data, sums []byte) error` |
| `version.go` | installer `resolveLatestVersion` | `ResolveLatestVersion(source string, download func(string)([]byte,error)) (string, error)` |
| `download.go` | installer `DefaultDownload` | `DefaultDownload(url string) ([]byte, error)` |
| `replace.go` | deploy `handler.go` (backup/move/restore) | `Swap(targetPath, newFilePath string) (backup string, err error)`, `Rollback(targetPath, backupPath string) error` |
| `checksum_test.go` | installer test | `TestVerifyChecksum` |

`Swap` respeta la lógica `os.Rename` de deploy: respalda `targetPath` a `targetPath+".old"`,
mueve `newFilePath` en su lugar y restaura en caso de fallo.

Tu trabajo en este módulo es **pequeño**: añadir UNA función nueva (`IsOutdated`) con su test,
aplicar dos mejoras puntuales a funciones movidas, y escribir la documentación.

---

## Development Rules (MANDATORY)

- Comentarios y docs en **inglés**.
- **stdlib-only**: sin dependencias externas (sin `require` salvo el toolchain).
- **Sin `cmd/`**: es una librería.
- **No reimplementar lo movido.** Las 5 funciones de la tabla se quedan; solo se ajustan en las
  mejoras I-1/I-4 de abajo.
- **TDD**: el test nuevo va antes que la función nueva.

---

## Stage 1 — Añadir `IsOutdated` (única función nueva)

Comparación semver mínima, usada por el cliente para decidir si avisar/actualizar.

Test primero (`version_test.go`):
```go
package update

import "testing"

func TestIsOutdated(t *testing.T) {
    cases := []struct {
        cur, lat string
        want     bool
    }{
        {"v0.1.0", "v0.2.0", true},
        {"0.2.0", "v0.2.0", false},
        {"v1.2.3", "v1.2.2", false},
        {"dev", "v9.9.9", false},   // nunca molestar a builds "dev"
        {"v0.2.0", "garbage", false},
    }
    for _, c := range cases {
        if got := IsOutdated(c.cur, c.lat); got != c.want {
            t.Errorf("IsOutdated(%q,%q)=%v want %v", c.cur, c.lat, got, c.want)
        }
    }
}
```

Implementación (en `version.go`, junto a `ResolveLatestVersion`):
```go
// IsOutdated reports whether current is strictly older than latest (MAJOR.MINOR.PATCH,
// leading "v" optional, pre-release/build suffix ignored). Any parse failure -> false,
// so "dev"/empty builds are never reported as outdated.
func IsOutdated(current, latest string) bool {
    c, ok1 := parseSemver(current)
    l, ok2 := parseSemver(latest)
    if !ok1 || !ok2 {
        return false
    }
    for i := 0; i < 3; i++ {
        if c[i] != l[i] {
            return c[i] < l[i]
        }
    }
    return false
}

func parseSemver(v string) ([3]int, bool) {
    var out [3]int
    v = strings.TrimPrefix(strings.TrimSpace(v), "v")
    if i := strings.IndexAny(v, "-+"); i >= 0 {
        v = v[:i]
    }
    parts := strings.Split(v, ".")
    if len(parts) != 3 {
        return out, false
    }
    for i := 0; i < 3; i++ {
        n, err := strconv.Atoi(parts[i])
        if err != nil {
            return out, false
        }
        out[i] = n
    }
    return out, true
}
```
Añade `strconv` al import de `version.go`.

## Stage 2 — Mejora I-1: `Swap` robusto a cross-filesystem (EXDEV)

`Swap`/`Rollback` usan `os.Rename`, que falla si origen y destino están en filesystems
distintos (en deploy el binario nuevo puede venir de otro mount; en clientes `/tmp` suele ser
tmpfs). Introduce un helper `moveFile` y úsalo en `Swap` y `Rollback` en vez de `os.Rename`:

```go
// moveFile renames src to dst, falling back to copy+remove across filesystems (EXDEV).
func moveFile(src, dst string) error {
    if err := os.Rename(src, dst); err == nil {
        return nil
    }
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()
    info, err := in.Stat()
    if err != nil {
        return err
    }
    out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
    if err != nil {
        return err
    }
    if _, err := io.Copy(out, in); err != nil {
        out.Close()
        return err
    }
    if err := out.Close(); err != nil {
        return err
    }
    return os.Remove(src)
}
```
Añade `io` al import de `replace.go`. Reemplaza los `os.Rename(...)` de `Swap` y `Rollback`
por `moveFile(...)`. Mantén el resto de la lógica igual.

Test (`replace_test.go`): cubre Swap (con backup), Rollback y fresh-install (target inexistente
→ backup ""):
```go
package update

import (
    "os"
    "path/filepath"
    "testing"
)

func TestSwapAndRollback(t *testing.T) {
    dir := t.TempDir()
    target := filepath.Join(dir, "tool")
    src := filepath.Join(dir, "tool.new")
    os.WriteFile(target, []byte("old"), 0755)
    os.WriteFile(src, []byte("new"), 0755)

    backup, err := Swap(target, src)
    if err != nil || backup == "" {
        t.Fatalf("swap: backup=%q err=%v", backup, err)
    }
    if got, _ := os.ReadFile(target); string(got) != "new" {
        t.Errorf("target=%q want new", got)
    }
    if err := Rollback(target, backup); err != nil {
        t.Fatalf("rollback: %v", err)
    }
    if got, _ := os.ReadFile(target); string(got) != "old" {
        t.Errorf("after rollback target=%q want old", got)
    }
}

func TestSwapFreshInstall(t *testing.T) {
    dir := t.TempDir()
    target := filepath.Join(dir, "tool") // inexistente
    src := filepath.Join(dir, "tool.new")
    os.WriteFile(src, []byte("new"), 0755)
    backup, err := Swap(target, src)
    if err != nil || backup != "" {
        t.Fatalf("fresh: backup=%q err=%v", backup, err)
    }
    if got, _ := os.ReadFile(target); string(got) != "new" {
        t.Errorf("target=%q want new", got)
    }
}
```

## Stage 3 — Mejora I-4: headers de GitHub en `DefaultDownload`

Evita el rate-limit (60 req/h sin auth) y soporta token opcional. En `download.go`:
- Construye la request con `http.NewRequest`, fija `Accept: application/vnd.github+json`.
- Si hay token en `GITHUB_TOKEN` o `GH_TOKEN` (primero no vacío), añade
  `Authorization: Bearer <token>`.
- Mantén el `Timeout: 60s` y el chequeo de status 200.

Define constantes (en `download.go`): `EnvGitHubToken = "GITHUB_TOKEN"`,
`EnvGHToken = "GH_TOKEN"`. (No es testeable sin red; no añadas test de red.)

## Stage 4 — Documentación

- **`README.md`**: reemplaza el stub. Documenta las 6 funciones públicas con un ejemplo de cada
  caso: (a) resolver+verificar descarga (`ResolveLatestVersion`+`DefaultDownload`+`VerifyChecksum`),
  (b) comparar versión (`IsOutdated`), (c) swap con rollback (`Swap`/`Rollback`). Linkea `docs/API.md`.
- **`docs/API.md`**: una línea de contrato por símbolo exportado.
- **`README.md`** debe linkear todo lo que haya en `docs/`.

---

## Stages table

| Stage | Output | Done when |
|------|--------|-----------|
| 1 | `version_test.go` + `IsOutdated`/`parseSemver` en `version.go` | test verde |
| 2 | `moveFile` en `replace.go` + `Swap`/`Rollback` usan `moveFile`; `replace_test.go` | tests verdes; cross-fs cubierto |
| 3 | `DefaultDownload` con headers + token; constantes env | compila; vet limpio |
| 4 | `README.md`, `docs/API.md` | cada símbolo documentado; README linkea docs |

## Acceptance criteria

- `go test ./...` verde; `go vet ./...` limpio; sin deps externas; sin `cmd/`.
- Las 5 funciones movidas siguen existiendo con sus firmas; solo `Swap`/`Rollback`/`DefaultDownload`
  cambian internamente por I-1/I-4.
- API pública total: `VerifyChecksum`, `ResolveLatestVersion`, `IsOutdated`, `DefaultDownload`,
  `Swap`, `Rollback` (+ constantes env).
