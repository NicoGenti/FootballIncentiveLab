---
id: 1-test-bootstrap-hello-world
title: Test bootstrap: hello world
status: draft
created: 2026-10-19
updated: 2026-10-19
owner: issue-bridge
executor: orchestrator
gh_body_sha: 5090cebc8d939dbf7400ea87e3404d8d2e584c31d5d098e55d88711f878d2471
gh_issue: 1
gh_state: OPEN
gh_synced_at: 2026-10-19T16:09:02Z
---

## Goal
Verificare il ciclo completo issue -> agente -> PR -> plan-sync su un repo Go.

## Success Criteria
- Il piano viene creato in plan/draft/ e spostato in plan/qa/ dall'agente.
- I test Go passano (go test ./...).
- La PR viene aperta in bozza e mergiata.
- issue_bridge check → BRIDGE OK.

## Scope
- In: verifica del flusso agentico su repo Go.
- Out: modifiche al codice applicativo.

## Safety
- Nessun push su main dall'agente.
- Solo il job di adozione committa su master.

## Inputs Available
- Repo Go con go.mod e go.sum.
- Board plan/ con strumenti venduti.
- Workflow agent-issue, plan-adopt, plan-sync.

## Outputs Required
- PR in bozza con il piano in qa/.
- issue_bridge check → BRIDGE OK.

## Test Plan
- go test ./...
- issue_bridge check → exit 0

## Verification
- BRIDGE OK, exit 0, 0 findings.
- PR mergiata, plan-sync success.

## Notes / Edge Cases
- Repo Go: usare go mod download e go test ./... (non pnpm).
- Il modello free potrebbe essere lento su issue complesse.
