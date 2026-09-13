# reviewer-auto — PR #4 (feat: add pagination to GET /users)

Última revisión: 2026-09-13 (pass 19 reviewer-auto)
Veredicto: LISTO -> rutea a pr-merger (approve).
Estado del fuente en feat/users-pagination @ HEAD 4df06a7: re-verificado archivo por
archivo contra el diff real del PR (handler/pagination.go + _test.go,
model/pagination.go + _test.go, store/user_store.go + _test.go,
user_handler.go + _test.go, README.md).

## Cierre de hallazgos previos (pass 1-18)

Los 5 hallazgos históricos fueron escritos contra una versión ANTERIOR del
branch. Re-verificados uno por uno contra el HEAD actual del PR: TODOS resueltos.

1. **[RESUELTO] API de parsing consistente entre test e impl.**
   `pagination.go` expone `parsePageParams(r) pageParams` con `pageParams{Page,PageSize}`
   y métodos `.limit()`/`.offset()`/`.toModelPage()`, más `newPaginatedResponse(data,p,total)`.
   `pagination_test.go` usa EXACTAMENTE esa API (`parsePageParams`, `.Page`, `.PageSize`,
   `.limit()`, `.offset()`, `newPaginatedResponse`, `.TotalPages`). El paquete `handler` compila.
2. **[RESUELTO] `total_pages` existe y se computa.** `model.PagedUsers` tiene el campo
   `TotalPages int \`json:"total_pages"\``; `newPaginatedResponse` lo calcula como
   `ceil(total/PageSize)` = `(total+size-1)/size`, y `0` cuando `total==0`.
   Consistente con README, PR body y `TestNewPaginatedResponse`.
3. **[RESUELTO] Firma del store alineada con su test.** La firma real es
   `ListPaginated(ctx context.Context, page model.Page) ([]*model.User, int, error)`
   y `user_store_pagination_test.go` la invoca con `(context.Background(), model.Page{...})`
   (2 args, value Page), verificando `errors.Is(err, ErrInvalidPage)`. La validación
   `page.Size < 1 || page.Offset() < 0` corre ANTES de tocar la DB, por lo que `Db:nil`
   es seguro en el test. El paquete `store` compila.
4. **[RESUELTO] `Page.Offset()` con cota inferior.** `model/pagination.go` guarda
   `if p.Number < 1 || p.Size < 1 { return 0 }` antes de `(Number-1)*Size`; nunca negativo.
5. **[RESUELTO] Sin campo muerto en `UserHandler`.** El struct es `{lister UserLister; crud *store.UserStore}`;
   `lister` sirve `List` y `crud` sirve `Get/Create/Delete`. Ambos se usan; no hay campo huérfano.

## Cosas bien hechas (confirmadas en el diff actual)
- context propagado (QueryContext/QueryRowContext) en `ListPaginated`.
- `defer rows.Close()` + `rows.Err()` chequeado tras la iteración.
- errores envueltos con `%w` (count/query/scan/iterate) — inspeccionables con errors.Is/As.
- `ErrInvalidPage` como sentinel + guard-clause fail-fast antes de la DB.
- interfaz `UserLister` para testear el handler sin DB real (mockLister).
- prealoc de slice con `cap = page.Size`.
- `writeJSON`/`writeJSONError` centralizan headers/status y ya NO ignoran el encode
  de forma silenciosa problemática (encode error explícitamente descartado en helper).
- `page_size` capado a 100 (protege el backend); defaults sanos (page=1, size=20).
- Cobertura de tests sólida y table-driven: parsing (defaults/explícito/negativo/no-numérico/cap),
  total_pages (exacto/resto/vacío/parcial), handler (status, envelope, propagación de page al store,
  error 500), store guard-clauses, y `Page.Offset()`.

## Nits (bajo, NO bloquean)
- El path CRUD (`Get/Create/Delete`) mantiene los `// BUG:` pre-existentes (sin context,
  errores sin envolver, 404 devuelto como 500). Están FUERA del scope de este PR de
  paginación y el autor lo documentó explícitamente. No bloquean el merge de esta feature.
- `pageParams.toModelPage()` y el par `pageParams`/`model.Page` conviven; podría unificarse
  a futuro, pero hoy es coherente y está testeado.

## Veredicto
Sin hallazgos ALTO/MEDIO en el fuente actual. Solo nits de severidad baja
(pre-existentes y fuera de scope). El PR cumple el contrato documentado
(params, defaults, cap, envelope con total_pages) y trae tests deterministas.
`go build ./... && go test ./...` esperado en verde.
-> APPROVE. Ruteo a pr-merger.

## Historial de passes
- pass 1-17: 5 hallazgos (4 ALTO + 1 MEDIO) contra una versión previa del branch.
- pass 18 (2026-09-13): re-reporte de los mismos 5 (artefacto describía HEAD antiguo b420157).
- pass 19 (2026-09-13): re-verificación archivo por archivo contra el diff REAL del PR
  @ HEAD 4df06a7. Los 5 hallazgos están resueltos en el fuente actual. Sin ALTO/MEDIO
  pendientes -> APPROVE / ruteo a pr-merger.
