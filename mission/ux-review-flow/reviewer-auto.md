# reviewer-auto — Review de PR #7

**Repo:** jaavier/go-api-service
**PR:** #7 — `feat: correct write-path for POST /users and DELETE /users/{id}`
**Head:** `feat/users-write-path` → **Base:** `main`
**Fecha:** 2026-09-13
**Decisión:** ✅ APPROVE (sin hallazgos alto/medio)

## Alcance revisado
- `internal/store/user_store.go` (Create / DeleteByID)
- `internal/handler/user_handler.go` (Create / Delete + interfaces)
- `internal/handler/user_create_test.go` (nuevo)
- `internal/handler/user_delete_test.go` (nuevo)
- `README.md` (docs de endpoints)

## Análisis por categoría

### Bugs / lógica
Sin hallazgos. Los `// BUG:` originales del write-path fueron resueltos:
- `Create` ahora captura el id generado con `QueryRowContext(... RETURNING id).Scan(&u.ID)`.
- `DeleteByID` inspecciona `RowsAffected()` y devuelve `ErrNotFound` cuando es 0 (antes era no-op silencioso).

### Seguridad
Sin hallazgos. Todas las queries del write-path usan parámetros vinculados (`$1`, `$2`); sin concatenación de input. No hay secretos ni path traversal.

### Manejo de errores
Correcto. Errores envueltos con `%w` (`store: create user: %w`, `store: delete user %d: %w`). El handler distingue 404 (`errors.Is(err, store.ErrNotFound)`) de 500. Envelopes JSON consistentes vía `writeJSONError`.

### Recursos
Correcto. `defer r.Body.Close()` en `Create`. `QueryRowContext`/`ExecContext` no dejan `*sql.Rows` abiertos.

### Concurrencia
Sin código concurrente nuevo. Sin races.

### Semántica HTTP / validación
Correcta:
- `POST /users`: 201 (con id), 400 invalid body, 400 required (validación con `TrimSpace`), 500.
- `DELETE /users/{id}`: 204 (sin body), 400 invalid id, 404 not found, 500.
- Context propagado end-to-end (`r.Context()` → store).

### Tests
Sólidos y table-driven, sin DB (mocks `mockCreator`/`mockDeleter`). Cubren happy path, validación (store no llamado), 404, 500 y propagación de context. Coherentes con `NewUserHandler` (firma pública intacta).

### Consistencia
Helpers `writeJSON`/`writeJSONError` presentes; interfaces `UserCreator`/`UserDeleter` satisfechas por `*store.UserStore`; imports correctos (`strings`, `errors`, `strconv`). README alineado con el comportamiento real.

## Notas (severidad BAJA, no bloqueantes)
- `internal/store/user_store.go`: persisten `// BUG:` en el read-path (`List` sin context, `rows.Close()` no diferido, `%v` en vez de `%w`, `Db` exportado). Están **fuera del alcance** de este PR (documentado como write-path only, para no colisionar con PRs #1/#2). No bloquean.
- Podría validarse formato de email en `Create` a futuro; hoy solo se exige no-blank, lo cual es aceptable para el contrato declarado.

## Conclusión
Ningún hallazgo de severidad alta o media. El PR resuelve correctamente el write-path con semántica HTTP adecuada, propagación de context, manejo de errores idiomático y tests representativos. **Listo para merge → approve.**
