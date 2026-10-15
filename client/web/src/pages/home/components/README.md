# pages/home/components/

`CategoryBreakdown` and `HeroBalanceCard` — presentational components used only by
`pages/home/`. Both render pre-computed values only; see `pages/home/README.md` for where the
aggregation math actually lives (`../index.tsx`).

## Testing

Pure render tests — pre-computed values in (amounts, percentages, labels), assert the rendered
output. The aggregation math that produces those values is tested at the container level (see
`pages/home/README.md`), not here.
