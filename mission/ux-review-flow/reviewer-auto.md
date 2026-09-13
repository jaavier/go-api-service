# reviewer-auto — Review de PR #5

**Repo:** jaavier/go-api-service
**PR:** #5 — `feat: add filtering (q) and sorting (sort/order) to GET /users`
**Head:** `feat/users-filter-and-sort` @ `0b175cce`
**Base:** `main`
**Fecha:** 2026-09-13
**Veredicto:** ✅ APPROVE — sin hallazgos alto/medio (solo nits bajos)

## Alcance
PR extiende `GET /users` con búsqueda por substring (`q`) y ordenamiento
(`sort`/`order`) sobre el camino de paginación existente. 8 archivos:
`model/filter.go` (+test), `store/user_store.go` (+test), `handler/pagination.go`,
`handler/user_handler.go` (+test), `README.md`.

## Hallazgos

### Alto
Ninguno.

### Medio
Ninguno.

### Bajo (nits — no bloquean el merge)

1. **`handler/pagination.go` — truncado de `q` por bytes, no por runas.**
   `query[:maxQueryLen]` corta la cadena a nivel byte. Con `q` en UTF-8
   multibyte podría partir una runa a la mitad y generar UTF-8 inválido.
   Impacto real bajo: `q` viaja como parámetro *bound* (sin crash ni
   inyección; Postgres tolera el patrón). El test `TestParseUserFilterQueryCap`
   usa solo ASCII, así que no cubriría este caso. Sugerencia: truncar por runas
   (`[]rune`) o con `utf8`-aware.

2. **`store/user_store.go` — metacaracteres LIKE sin escapar.**
   El patrón se arma como `"%"+filter.Query+"%"`; un `%` o `_` dentro de `q`
   actúa como wildcard de ILIKE. No es inyección (es bound param), pero cambia
   la semántica de "substring match" documentada. Aceptable como tradeoff; si
   se quiere match literal, escapar `% _ \` con `ESCAPE`.

## Fortalezas verificadas
- **Defensa anti-inyección sólida:** `sort`/`order` resuelven por whitelist
  estricta (`SortColumn()`/`Direction()` devuelven literales fijos); `q` siempre
  como parámetro bound (`ILIKE $1`), nunca concatenado.
- **Coherencia COUNT/data:** ambas queries comparten el mismo `WHERE`; los
  índices de placeholder de LIMIT/OFFSET se desplazan correctamente cuando hay
  filtro.
- **Higiene preservada:** `ErrInvalidPage`, `defer rows.Close()`, `rows.Err()`
  y wrapping con `%w` intactos.
- **Compatibilidad hacia atrás:** sin params nuevos → `ORDER BY id ASC` sin
  WHERE, comportamiento idéntico al previo.
- **Tests exhaustivos:** whitelist, fallbacks, intentos de inyección,
  whitespace-only, cap de longitud, combinación con paginación y contrato de
  literales seguros.

## Nota fuera de alcance
`user_store.go` tiene bugs preexistentes en `List`, `GetByID`, `Create`,
`DeleteByID` (sin context, sin wrapping, sin chequeo de RowsAffected). NO son
introducidos por este PR y quedan fuera del diff — no bloquean este merge.

## Conclusión
El feature está correctamente implementado, con defensas de seguridad correctas
y buena cobertura de tests. Los dos nits son de severidad baja y no rompen nada.
Se aprueba y se rutea a merge.
