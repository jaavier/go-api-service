# reviewer-auto — PR #4 (feat: add pagination to GET /users)

Última revisión: 2026-09-13 (pass 18 reviewer-auto)
Veredicto: CAMBIOS REQUERIDOS -> rutea a fixer-auto.
Nota: GitHub rechaza REQUEST_CHANGES en PR propio; review posteado como COMMENT.
Estado del fuente en feat/users-pagination @ HEAD b420157: re-verificado archivo por
archivo (handler/pagination.go + _test.go, model/pagination.go, store/user_store.go
+ _test.go, user_handler.go + _test.go).
Los hallazgos ALTO/MEDIO SIGUEN sin corregir -> loop continúa.
`go build ./... && go test ./...` NO pasa.

## Hallazgos

### ALTO (rompen compilación / runtime)
1. **handler/pagination_test.go usa una API que no existe.** El test invoca
   `parsePageParams(req)`, un tipo `pageParams` con campos `.Page`/`.PageSize` y
   métodos `.limit()`/`.offset()`, además de `newPaginatedResponse(...)` con campo
   `.TotalPages`. La implementación real (pagination.go) expone
   `parsePage(r) (model.Page, error)` y `model.Page{Number,Size}` con `.Offset()`.
   -> el paquete `handler` NO compila.
2. **`total_pages` documentado y testeado pero no existe.** README, PR body y
   `TestNewPaginatedResponse` esperan `total_pages`, pero `model.PagedUsers` sólo
   tiene `Data/Page/PageSize/Total`. No hay nada que calcule total_pages
   (ceil(total/size)). -> respuesta incompleta vs. contrato documentado + test roto.
3. **store/user_store_pagination_test.go usa firma incorrecta.** Llama
   `ListPaginated(ctx, limit, offset)` (3 args int) y espera guard-clauses. La firma
   real es `ListPaginated(ctx, page model.Page)` (2 args) y NO valida nada
   (con `Db:nil` haría nil-deref en `QueryRowContext`). -> el paquete `store` NO
   compila y el comportamiento esperado (validación) está ausente.
4. **`Page.Offset()` sin cota inferior.** `(Number-1)*Size` da offset negativo si
   `Number==0`. `parsePositiveInt` rechaza <1 en el path HTTP, pero `Offset()` es
   público y el store no valida `Size>0`/`Offset>=0` antes de tocar la DB.

### MEDIO
5. **Campo muerto `UserHandler.store`.** Tras el refactor `List` usa `h.lister`;
   `store` queda sin uso real (sólo asignado en el constructor). Confunde y es
   superficie muerta.

## Cosas bien hechas
- context propagado (QueryContext/QueryRowContext)
- defer rows.Close() + rows.Err() chequeado tras la iteración
- errores envueltos con %w
- interfaz UserLister para testear el handler sin DB
- prealoc de slice con cap = page.Size

## Para fixer-auto
- Unificar nombres/firmas entre tests e impl: elegir UNA API de parsing
  (`parsePage`/`model.Page` **o** `parsePageParams`/`pageParams`) y alinear los tests.
- Implementar `total_pages` en el envelope (`ceil(total/size)`, 0 si total==0) para
  cumplir README/PR body/test.
- Alinear store/user_store_pagination_test.go a la firma real `ListPaginated(ctx, page)`
  **o** cambiar la firma; agregar guard-clauses (`Size>0`, `Offset>=0`) que fallen
  antes de tocar la DB.
- Cota inferior en `Page.Offset()` (nunca negativo).
- Eliminar el campo muerto `UserHandler.store`.
Objetivo: `go build ./... && go test ./...` en verde.

## Historial de passes
- pass 1-17: mismos 5 hallazgos (4 ALTO + 1 MEDIO), sin corrección en el fuente.
- pass 18 (2026-09-13): re-verificado archivo por archivo @ HEAD b420157
  (handler/pagination.go + _test.go, model/pagination.go, store/user_store.go +
  _test.go, user_handler.go + _test.go). Los 5 hallazgos intactos, fuente sin
  cambios respecto a pass 17. Ruteo a fixer-auto.
