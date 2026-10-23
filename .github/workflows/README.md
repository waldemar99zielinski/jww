# GitHub Actions

- `publish-image.yml` — manual (`workflow_dispatch`) build and push of the
  application image to DigitalOcean Container Registry. It also creates a
  matching Git tag and then calls `deploy-image.yml` with that image tag.
- `deploy-image.yml` — updates an existing DigitalOcean App Platform app so one
  of its components uses the supplied image tag. It runs as a reusable workflow
  from `publish-image.yml`, and can also be triggered manually to redeploy an
  already published tag.

## Configuration

Repository variables:

| Variable | Purpose |
|---|---|
| `DIGITALOCEAN_REGISTRY` | DOCR registry name (`registry.digitalocean.com/<name>`). |
| `DIGITALOCEAN_APP_NAME` | Name of the existing App Platform app to update. |
| `DIGITALOCEAN_APP_SERVICE` | Name of the app component whose image should change. |

Repository secret:

| Secret | Purpose |
|---|---|
| `DIGITALOCEAN_ACCESS_TOKEN` | Token that can push to DOCR and update the App Platform app. |

The App Platform component must already reference the `jww` image from DOCR; the
deploy workflow only changes its tag, leaving the rest of the app spec intact.

## Manual deployment

Run **Deploy Docker image** from the Actions tab and provide an existing image
tag, for example:

```
jww-1.2.3-abcde-23-10-2026
```
