---
name: sync-upstream
description: Triggers the upstream sync workflow to merge latest changes from coder/coder
---

# Sync Upstream

Manually trigger the upstream sync GitHub Action and report the result.

## Workflow

1. **Trigger the workflow**:

   ```sh
   gh workflow run sync-upstream.yaml --repo noeljackson/coder
   ```

2. **Wait for it to start** (a few seconds), then get the run ID:

   ```sh
   gh run list --repo noeljackson/coder --workflow sync-upstream.yaml --limit 1
   ```

3. **Wait for completion** and show the result:

   ```sh
   gh run watch <RUN_ID> --repo noeljackson/coder
   ```

4. **Report the outcome** to the user:
   - If it passed: check the logs for whether it merged commits or was
     already up to date.
   - If it failed: show the failed step logs with
     `gh run view <RUN_ID> --repo noeljackson/coder --log-failed`.

5. **If it merged new commits**, pull locally:

   ```sh
   git pull origin main
   ```
