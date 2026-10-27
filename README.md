# Windows local reconnect: verification evidence

- OS: Windows amd64; Go toolchain: go1.27.3.
- Upstream base: e52572aa2640c59f7ba04601fe1a7bcaae377f2d.
- Proposed fix: 00e4aae76ede09822bad653a72359adc87a3461d.
- All protocol test data is synthetic. No canvas contents, vault data, remote addresses or user configurations are included.

## Confirmed causes

1. `version.Build()` identifies a local executable by modification time and size. Windows runs its daemon from a content-addressed copy. Before this fix, copying preserved bytes but changed modification time. `hostconn.connectLocal` therefore classified an identical daemon as a different build on reconnect. The copy regression fails on upstream and passes after preserving the timestamp before rename.
2. User-data sync grouped reads by 200 paths and writes by 100 changes, without considering encoded bytes. Six 3 MiB canvas revisions exceed the 16 MiB frame limit after base64 encoding. On upstream, `sync.read` closes the client with EOF; `sync.write` fails with a 25,166,237-byte payload. The read failure causes reconnects, which can then trigger the separate local-build mismatch.

## Fix

Preserve copied executable timestamps; batch reads and writes by encoded JSON frame size, including base64 and envelopes; return an RPC size error instead of closing the connection if a reply becomes too large. This leaves the wire protocol and its 16 MiB limit unchanged, supports a single 9 MiB file, preserves stale-write reporting and keeps 50 small files in one read/write batch. An individually untransferable file returns a sync error; per-file streaming is outside this change.

## Test evidence

- `before-tests.txt`: new regression tests against unchanged upstream product code. All four selected cases fail for the documented reasons.
- `after-tests.txt`: the same scenarios plus oversized-file handling pass after the patch.
- `regressions-repeated.txt`: new regressions repeated three times, including checking that small files remain batched.
- `safe-validation.txt`: final validation with USERPROFILE, APPDATA and LOCALAPPDATA redirected to a disposable profile. Full client, protocol, usersync and app-copy package tests pass; the daemon sync regressions and existing two-daemon sync test pass three times.
- `windows-suite-comparison.md`: candid comparison of broader Windows test failures on upstream and patched trees; no clean full-suite claim.
- `go vet` and Staticcheck v0.8.2 passed for client, daemon, cmd/enowx, core/usersync, proto and hostconn.

The two-daemon tests use actual Server/Client protocol implementations over net.Pipe. They verify byte-for-byte content, an unchanged second sync, and daemon.info on both still-open connections. They do not exercise SSH or Tailscale networking.

## Visual evidence and provenance

- `before-restarting-daemon.png`: original user-provided screenshot of the reported installed v0.3.0 behavior.
- `before-disconnected.png`: cropped from the user's original disconnected screenshot to show only the error card.
- `after-connected.png`: cropped native screenshot from a separate patched application with an empty ENOWX_HOME, showing a connected local host. It confirms the clean Windows startup path; the protocol tests, not this empty-profile screenshot, prove the large-transfer fix.

Images are screenshots, not reconstructed UI. Cropping excludes unrelated windows, floating controls and account-usage indicators; application content was not altered. No new recording is included: Computer Use was stopped by the user, so screen control was discontinued.

## Reproduce

In a disposable checkout of the PR branch, run:

```powershell
$env:GOTOOLCHAIN = 'go1.27.3'
go test ./cmd/enowx ./daemon -run 'TestDaemonCopyPreservesBuildIdentity|TestSyncReadLargerThanFrame|TestSyncWriteLargerThanFrame|TestUserSyncLargerThanFrame|TestOversizedSyncReplyKeepsConnection|TestSyncLargeFileAndStaleWrites|TestSyncSingleOversizedFileKeepsConnection|TestSmallSyncFilesStayBatched' -count=3 -v
```

To reproduce the original failures while keeping the regression tests in that disposable checkout, restore only the original product files, then run the four-test command from before-tests.txt:

```powershell
git restore --source=e52572aa2640c59f7ba04601fe1a7bcaae377f2d -- client/usersync.go cmd/enowx/daemoncopy.go daemon/rpc.go
go test ./cmd/enowx ./daemon -run 'TestDaemonCopyPreservesBuildIdentity|TestSyncReadLargerThanFrame|TestSyncWriteLargerThanFrame|TestUserSyncLargerThanFrame' -count=1 -v
```
