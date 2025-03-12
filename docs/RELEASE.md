# Release Process

This document describes the release process for Pod Ashiato.

## Creating a New Release

1. Update version information in relevant files if applicable
2. Commits changes to the `main` branch
3. Create a new release tag using GitHub's release feature:
   - Go to the repository on GitHub
   - Click on "Releases" in the right sidebar
   - Click "Create a new release"
   - Enter a tag in the format `v1.0.0` (following [Semantic Versioning](https://semver.org/))
   - Add release notes describing changes
   - Click "Publish release"

## Automated Actions

When a new release is created, the following automated actions occur:

1. GitHub Actions workflow `docker-publish.yml` is triggered
2. The workflow builds the Docker image
3. The image is tagged with:
   - The exact version (e.g., `v1.0.0`)
   - The major.minor version (e.g., `1.0`)
   - `latest` tag (for non-prerelease versions)
4. The tagged images are pushed to GitHub Container Registry (ghcr.io)

## Manual Release

If needed, you can also trigger the Docker image build manually:

1. Go to the "Actions" tab in the GitHub repository
2. Select the "Build and Publish Docker Image" workflow
3. Click "Run workflow"
4. Enter the desired tag for the image
5. Click "Run workflow"

## Using Released Images

Released images can be used in Kubernetes deployments:

```yaml
containers:
- name: pod-ashiato
  image: ghcr.io/takutakahashi/pod-ashiato:v1.0.0  # Specify the version
```

Or pulled directly with Docker:

```bash
docker pull ghcr.io/takutakahashi/pod-ashiato:v1.0.0
```