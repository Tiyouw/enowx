# Broader Windows suite comparison

Base: e52572aa2640c59f7ba04601fe1a7bcaae377f2d
Proposed change: 00e4aae76ede09822bad653a72359adc87a3461d

The broader command was `go test ./client ./proto ./core/usersync ./hostconn ./cmd/enowx ./daemon -count=1`. It is **not green** on this Windows environment, even without this patch.

Every failure observed on the patched tree also occurred on the unchanged upstream tree. These suites include existing Windows path, Unix-command and HOME/USERPROFILE assumptions. This comparison does not claim that the failures were fixed. The final safe validation uses an isolated Windows profile and the affected sync regressions. Raw broad-suite logs are not published because some upstream tests scan machine settings.

| Test | Unchanged upstream | Patched |
| --- | --- | --- |
| TestAddFilesIntoAFolder | FAIL | FAIL |
| TestAgentAPIThroughTheDaemon | FAIL | not failed |
| TestBrowseFoldersThroughTheProtocol | FAIL | FAIL |
| TestBrowserActionsGoThroughTheApp | FAIL | FAIL |
| TestCanvasAgentStartsOnTheCanvas | FAIL | FAIL |
| TestChangeFilesOfTheFolder | FAIL | FAIL |
| TestCloneMakesAProject | FAIL | FAIL |
| TestClonePrefsAreKept | FAIL | FAIL |
| TestConnectThroughOwnTailnet | FAIL | FAIL |
| TestEnvSyncThroughTheProtocol | FAIL | FAIL |
| TestMoveOldHomeTakesOnlyEnowXsFiles | FAIL | FAIL |
| TestNotesAgentStartsOnNotesOnly | FAIL | FAIL |
| TestTrafficMCPForAgents | FAIL | FAIL |
| TestTranscriptFollowsTheSessionsFile | FAIL | FAIL |
| TestUnshownTranscriptIsDroppedAndReadAnew | FAIL | FAIL |

New failing test names on patched tree: 0. A test failing only on baseline is not claimed as fixed by this change.
