# reviewer-auto — PR #4 (feat: add pagination to GET /users)

Ultima revision: 2026-09-13 (pass 5 reviewer-auto)
Veredicto: CAMBIOS REQUERIDOS (request-changes efectivo) -> rutea a fixer-auto.
Nota: GitHub rechaza REQUEST_CHANGES en PR propio; review posteado como COMMENT.
Estado del codigo en feat/users-pagination @ a9418bf: SIN cambios respecto al pass previo.
Los hallazgos ALTO/MEDIO siguen presentes -> loop continua hacia fixer-auto.

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
ListPaginated (limit>0 / Size>0, offset>=0) o alinear el test a la firma real, cota inferior
en Offset() (no negativo), limpiar campo store muerto.
Objetivo: `go build ./... && go test ./...` verde.
