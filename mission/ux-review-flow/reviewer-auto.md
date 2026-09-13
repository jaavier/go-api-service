# reviewer-auto — Review de PR #6

- **Repo:** jaavier/go-api-service
- **PR:** #6 — `feat: proper 404 + context propagation for GET /users/{id}`
- **Head:** `feat/get-user-404-context`  →  **Base:** `main`
- **Fecha:** 2026-09-13
- **Veredicto:** ✅ APPROVE (sin hallazgos alto/medio)

## Alcance del diff
| Archivo | Cambio |
|---------|--------|
| `internal/store/user_store.go` | `ErrNotFound` centinela; `GetByID(ctx, id)` con `QueryRowContext`, distingue `sql.ErrNoRows`, wrap `%w`. |
| `internal/handler/user_handler.go` | Interfaz `UserGetter`; `Get` propaga `r.Context()`, mapea 400/404/500 vía `writeJSON*`. |
| `internal/handler/user_get_test.go` | Tests table-driven con mock: 200/404/500/400 + Content-Type + ctx. |
| `README.md` | Contrato documentado de `GET /users/{id}`. |

## Análisis por categoría

### Bugs / lógica — OK
- Se corrige el bug real declarado: `sql.ErrNoRows` ahora → **404** (antes **500**). Mapeo con `errors.Is(err, store.ErrNotFound)`: correcto.
- Parseo de id inválido → **400** `{"error":"invalid id"}` antes de tocar el store. Correcto.

### Seguridad — OK
- `GetByID` usa query parametrizada (`$1`), sin concatenación. Sin inyección.
- No hay secretos ni paths.

### Manejo de errores — OK
- No se tragan errores: not-found vs genérico bien separados; wrap con `%w` preserva la cadena para `errors.Is/As`.

### Recursos / concurrencia — OK
- `QueryRowContext` no requiere `Close` explícito (Scan libera). Sin goroutines nuevas ni locks.

### Semántica HTTP — OK
- 200/400/404/500 coherentes con README. `writeJSON`/`writeJSONError` setean `Content-Type: application/json` y status. Unifica el shape JSON de error.

### Tests — OK
- Cubre los 4 caminos del contrato, verifica status, Content-Type y propagación de contexto. Usa `UserHandler{getter: g}` (campo no exportado, mismo paquete) — compila.

### Compatibilidad — OK
- `NewUserHandler(*store.UserStore)` sin cambios. Único cambio de firma: `UserStore.GetByID` (ahora requiere `ctx`); único caller interno (handler) actualizado.

## Nits (severidad BAJA — no bloquean)
- Quedan `// BUG:` preexistentes en `Db` exportado, `List`, `Create`, `Delete`. Están **fuera de scope** declarado del PR; no son regresiones de este diff.

## Conclusión
El read-path de `GetByID` queda correcto de punta a punta. Sin hallazgos de severidad alta/media. **Listo para merge.**
