# Module Inventory

| Module | Responsibility | Files | Complexity | S.U.P.E.R Score |
|:--|:--|--:|:--|:--|
| Root dispatch | Route root commands and global flags | 1 | Low | S green, U green, P partial, E green, R green |
| Completion generation | Emit static shell scripts for supported shells | 1 | Low | S green, U green, P green, E green, R green |
| Help text | User-facing usage text | 1 | Low | S partial, U green, P partial, E green, R partial |
| Schema registry | Machine-readable command contracts | 1 | Medium | S partial, U green, P green, E green, R partial |
| CLI tests | Integration-level command regression tests | 1 | Medium | S partial, U green, P green, E green, R partial |

## Notes

The completion command stays local-only. It does not read config, auth state, network, or Dida365 account data.
