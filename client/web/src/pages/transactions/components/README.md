# pages/transactions/components/

`TransactionFiltersDropdown`, `FilterButton`, `TransactionFiltersSummary` — presentational
components used only by `pages/transactions/`. See `pages/transactions/README.md` for why these
are three separate files and where the shared `filterLabels` map lives.

## Testing

- `TransactionFiltersDropdown`'s draft/applied state split (see `pages/transactions/README.md`)
  is the main thing worth testing here directly, independent of the page container.
- `FilterButton`/`TransactionFiltersSummary` are simple pure-render tests — props in, assert
  rendered labels/pill state.
