# Doctor Upgrade Check

- **Description**: Added `dida doctor --check-upgrade` as an explicit, advisory
  GitHub Releases update check inside the normal doctor diagnostic.
- **Date**: 2026-07-07
- **Mode**: LOCAL_ONLY
- **Result**: `doctor` keeps local-only default behavior, reports
  `upgrade_check` when requested, and treats update lookup failures as
  diagnostic data instead of command failure.
- **Details**: [MASTER.md](progress/MASTER.md)
