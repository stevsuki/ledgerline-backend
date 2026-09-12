# Category refactor — done, pending your check

`master_categories` is now a live shared vocabulary instead of a seed template.
Code and migrations are in place; **nothing is committed.**

**Delete this file once you have run through the checks at the bottom.**

---

## What the model is now

1. `GET /categories/options` returns **master rows ∪ the account's own
   categories**. A master row drops off the list once the account holds one of
   the same `(name, type)`. It takes `?type=income|expense` to narrow, or
   nothing for both — the `slug` mechanism is gone.
1b. **`GET /categories` returns the same union**, each row flagged `is_own`.
   A picker offering a bucket the list never mentions is a screen disagreeing
   with itself. Shared rows cannot be deleted (nothing to delete); editing one
   adopts it first, so `PATCH /categories/:id` resolves like the writers do.
2. **No seeding at registration.** Registration creates a user and nothing else.
3. `categories.master_category_id` is **gone**. After creation a category is
   simply the account's own; provenance is not tracked.
4. **"Others" is an ordinary row** — no delete protection, no type lock.
5. Transactions and budgets still send **one `category_id`**. The server
   resolves it — `categories` by id + user_id, then `master_categories`, and a
   hit there is **adopted**: name, type, icon and colour are copied into a row
   the account owns. Idempotent, so a retry or a second tab yields the same row.
6. `transactions.category_id` and `budgets.category_id` keep their FK.
   **A master id is never stored** — it is only ever swapped in-request.

Accepted trade-off: renaming an adopted category puts its master row back on
offer, because nothing records that it was adopted. That record would be the
column dropped in (3).

### Two decisions taken along the way

**`master_categories.type = 'both'` was removed** (migration 000038). It could
not survive adoption: one id meaning two directions leaves the resolver no way
to know which to create.

**"Others" is gone altogether** (migration 000039), replaced by a **New** button
beside the category select in the add-transaction sheet. Naming a category there
is one field and one click, and it **commits on its own** before the transaction
is saved — which is what separates it from the flow removed earlier, where the
category was created *during* the save and stranded whenever that save failed.
Icon and colour are left blank; the categories screen is where a tile is chosen.

**The ledger's category filter keeps the shared rows out.** Both endpoints now
return the union, so `getTransactionFilterOptions` drops everything with
`is_own: false`: narrowing to a category nothing was ever filed under can only
return nothing. You narrow by what you have, and pick from what you could have.

---

## Migrations — all applied

| # | What |
|---|---|
| 000033 | `master_categories` gains type/icon/color; 4 income rows |
| 000034 | categories unique per `(user_id, type, lower(name))` |
| 000035 | backfill income categories + icons for existing users |
| 000036 | unlink hand-made categories from the Others master |
| 000037 | drop `categories.master_category_id` |
| 000038 | drop the `both` type; Others back to expense |
| 000039 | drop the "Others" master row entirely |

Verified in a rolled-back transaction before applying: a fresh account is
offered all 7 expense master rows, adopting one twice inserts once, and the
adopted row disappears from the offer.

---

## Files changed

**Backend** — `domain/category.go`, `domain/master_category.go`,
`service/{auth,category,transaction,budget}_service.go`,
`repository/postgres/category_repository.go`,
`repository/postgres/model/category_model.go`,
`delivery/http/dto/category_dto.go`, `mocks/category_repository.go`,
`cmd/api/main.go`, and two tests. `authService` lost its now-dead
`categoryRepo` dependency.

**Frontend** — `lib/api/categories.ts`, `lib/categories/actions.ts`,
`lib/category-fields.ts`, `lib/data/category-list.ts`,
`lib/data/transactions.ts`, `components/categories/category-editor.tsx`.
Dead code removed along the way: `getCategoryPicker`, `FIXTURE_PICKER`,
`MASTER_DEFAULTS`, `FALLBACK_MASTER_ID`, `NIL_UUID`.

Green: `go build` · `go vet` · `go test ./...` · `tsc --noEmit` · `eslint` ·
`next build`.

---

## What is left for you

- [ ] **Restart the backend.** The running process predates all of this.
- [ ] Register a fresh account. Expect: `/categories` lists all 11 shared rows
      marked "not yet yours", and the transaction sheet offers the same.
- [ ] Save a transaction against one. Expect: it saves, and that category now
      appears in `/categories` wearing the master row's icon and colour.
- [ ] Reopen the sheet. Expect: that row is there once, not twice.
- [ ] Check an existing account still behaves — its categories are untouched.
- [ ] Delete this file.

## How to resume if this is picked up later

> Baca `REFACTOR_CATEGORY.md` di ledgerline-backend. Pekerjaannya sudah selesai;
> yang tersisa hanya daftar pengecekan di bagian bawah.
