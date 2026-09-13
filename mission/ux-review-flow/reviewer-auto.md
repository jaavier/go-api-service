# reviewer-auto — PR #4 (feat: add pagination to GET /users)

Veredicto: CAMBIOS REQUERIDOS (request-changes efectivo) — rutea a fixer-auto.
Nota: GitHub rechaza REQUEST_CHANGES en PR propio; review posteado como COMMENT.

## Hallazgos

### ALTO (rompen compilacion / runtime)
1. handler/pagination_test.go usa API inexistente: `parsePageParams`, tipo `pageParams`,
   metodos `.limit()`/`.offset()`, `newPaginatedResponse`, campo `.TotalPages`.
   Impl real: `parsePage(r) (model.Page, error)` + `model.Page{Number,Size}.Offset()`.
   -> paquete handler no compila.
2. store/user_store_pagination_test.go llama `ListPaginated(ctx, limit, offset)` (3 args)
   y espera guard-clauses. Firma real: `ListPaginated(ctx, page model.Page)` (2 args),
   sin validacion. -> paquete store no compila + comportamiento ausente.
3. `total_pages` documentado en README + PR body + asserted en test, pero
   `model.PagedUsers` no tiene `TotalPages` ni nada lo calcula.
4. `Page.Offset() = (Number-1)*Size` sin cota inferior -> offset negativo si Number==0.
   Store tampoco valida.

### MEDIO
5. `UserHandler.store` queda como campo muerto (List usa `h.lister`).

## Cosas bien hechas
- context propagado (QueryContext/QueryRowContext)
- defer rows.Close() + rows.Err() chequeado
- errores envueltos con %w
- interfaz UserLister para testear sin DB

## Para fixer-auto
Unificar nombres/firmas entre tests e impl, implementar total_pages, agregar
guard-clauses en store y cota en Offset(), limpiar campo store muerto.
