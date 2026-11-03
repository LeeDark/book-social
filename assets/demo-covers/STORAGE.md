# Local Artwork Storage

Images are kept outside Git in `data/demo-covers/`. The `pilot-v1`,
`pilot-v1-approved`, and `batch-01` directories preserve the `sources/`,
`bundle/`, and `revisions/` structure. Original PNGs and prepared WebPs are
kept separately. Text manifests, prompts, review decisions, and checksums remain
in `assets/demo-covers/`.

The local directory is ignored by Git and is not itself a backup. On 2026-11-03,
the user reported choosing independent durable storage and making a backup; its
location and verification were not independently inspected. Restore it on a new machine.
The two existing pilot archives remain local and ignored in `assets/demo-covers/`.

The pilot README and gallery were translated after the original packages were
frozen. Their historical SHA256SUMS therefore apply to the original archive
contents, not the translated copies in `assets/`. Verify the preserved archive
digest against the independently recorded value in the installation protocol,
then extract and verify the original package from the project root:

```bash
pilot_work_dir=$(mktemp -d)
tar -xzf assets/demo-covers/pilot-v1-approved.tar.gz -C "$pilot_work_dir"
(cd "$pilot_work_dir/pilot-v1-approved" && sha256sum -c SHA256SUMS)
```

For installation, use the verified original `bundle/` and the previously recorded
manifest hash. Extract `pilot-v1` in the same way. To use the translated gallery,
assemble a separate working copy of the current text files and corresponding
`data/demo-covers/` images; do not treat it as the frozen archive. The
`batch-01/review.html` gallery refers directly to local `data/demo-covers/`.
