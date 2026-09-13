# reviewer-auto — PR #4 (feat: add pagination to GET /users)

Ultima revision: 2026-09-13 (pass 10 reviewer-auto)
Veredicto: CAMBIOS REQUERIDOS (request-changes efectivo) -> rutea a fixer-auto.
Nota: GitHub rechaza REQUEST_CHANGES en PR propio; review posteado como COMMENT
(review id 5188885078 @ commit a038002).
Estado del fuente en feat/users-pagination: re-verificado archivo por archivo
(handler/pagination.go, model/pagination.go, store/user_store.go). Los hallazgos
ALTO/MEDIO SIGUEN sin corregir -> loop continua.

## Hallazgos

### ALTO (rompen compilacion / runtime)
1. handler/pagination_test.go usa API inexistente: `parsePageParams`, tipo `pageParams`,
   campos `.Page`/`.PageSize`, metodos `.limit()`/`.offset()`, `newPaginatedResponse`,
   campo `.TotalPages`. Impl real: `parsePage(r) (model.Page, error)` +
   `model.Page{Number,Size}.Offset()`. -> paquete handler NO compila.
2. store/user_store_pagination_test.go llama `ListPaginated(ctx, limit, offset)` (3 args int)
   y espera guard-clauses. Firma real: `ListPaginated(ctx, page model.Page)` (2 args),
   sin validacion (con Db:nil haria nil-deref). -> paquete store NO compila + comportamiento ausente.
3. `total_pages` documentado en README + PR body + asserted en test, pero
   `model.PagedUsers` no tiene `TotalPages` ni nada lo calcula.
4. `Page.Offset() = (Number-1)*Size` sin cota inferior -> offset negativo si Number==0.
   Store tampoco valida Size/Offset.

### MEDIO
5. `UserHandler.store` queda como campo muerto (List usa `h.lister`).

## Cosas bien hechas
- context propagado (QueryContext/QueryRowContext)
- defer rows.Close() + rows.Err() chequeado
- errores envueltos con %w
- interfaz UserLister para testear sin DB

## Para fixer-auto
Unificar nombres/firmas entre tests e impl (elegir UNA API), implementar total_pages
en model.PagedUsers (ceil(total/size), 0 si total==0), agregar guard-clauses en
ListPaginated (Size>0, offset>=0) o alinear el test a la firma real, cota inferior
en Offset() (no negativo), limpiar campo store muerto.
Objetivo: `go build ./... && go test ./...` verde.

## Historial de passes
- pass 1-6: mismos hallazgos, sin correccion. Commits de doc no tocan el fuente.
- pass 7 (2026-09-13): sin cambios en fuente respecto a pass 6 -> ruteo a fixer-auto.
- pass 8 (2026-09-13): fuente re-verificado (pagination.go, user_store.go); mismatch tests/impl
  intacto, total_pages ausente, sin guard-clauses. Se mantiene ruteo a fixer-auto.
- pass 9 (2026-09-13): re-verificado pagination.go + user_store.go + model/pagination.go @ commit
  57985e5; los 5 hallazgos (4 ALTO + 1 MEDIO) intactos. Review COMMENT id 5188882872. Ruteo a fixer-auto.
- pass 10 (2026-09-13): re-verificado handler/pagination.go + model/pagination.go + store/user_store.go
  @ commit a038002; los 5 hallazgos (4 ALTO + 1 MEDIO) intactos, fuente sin cambios. Review COMMENT
  id 5188885078. Ruteo a fixer-auto.
