# Pilot Demo Covers — Draft v1

This historical draft contained five candidates and had not been approved by the
user. These are original AI-generated Book Social demo illustrations, not covers
of specific editions or evidence of exclusive rights. The draft did not modify
the working database or public storage.

The [gallery](review.html) shows 240×360 thumbnails linking to 600×900 main
WebPs. Source PNGs are under `sources/`. The five concepts are a snowy London
street for *A Christmas Carol*, a Regency manor for *Pride and Prejudice*, an
Alpine landscape for *Frankenstein*, a rabbit and woodland for *Alice's Adventures
in Wonderland*, and an underwater submarine and reef for *Twenty Thousand Leagues
Under the Seas*. The collection uses painted gouache and engraved textures without
embedded text. No external reference images were used; this is not a guarantee
against resemblance.

The draft profile used PNG 1024×1536 sources and WebP 600×900 main images at
quality 84 and 240×360 thumbnails at quality 82, with 2:3 proportions and no
cropping or stretching. ImageMagick 6.9.12-98 produced derivatives:

```bash
convert sources/<slug>.png -resize 600x900 -strip -quality 84 bundle/<slug>-main.webp
convert sources/<slug>.png -resize 240x360 -strip -quality 82 bundle/<slug>-thumb.webp
```

Generation used the built-in image generation tool, not CLI/API fallback. The model
identifier was not exposed. [prompts.json](prompts.json) retains the prompts,
including the rejected first attempt; [artwork-provenance.json](artwork-provenance.json)
records slugs, sources, derivatives, checksums, sizes, and review status.

The [manifest](bundle/manifest.json) marks all five records `reviewed: false`;
the installer must reject this draft. Its historical SHA-256 is
`b31e1837d8c2ceff092cb8a7298b3ba9ab4dd2f848c7521d7b3698fa88335e3f`.
The ignored local archive `../pilot-v1-draft.tar.gz` and its digest preserve
the original package with normalized names, timestamps, UID and GID.

This README and gallery were translated after the original package was frozen.
See [storage instructions](../STORAGE.md) for assembling and verifying the
historical package; its recorded checksums describe the original files. Draft
artwork and profile were not approved at this historical point; approval required
a separate decision and did not replace application, HTMX, installation, or
backup/restore checks described in the [acceptance checklist](../../../docs/cover_acceptance_v0_3_2.md).
