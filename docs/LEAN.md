# Building Leaner Go Service Containers

Google documents how to do this with in a Cloud Build pipeline here: [Building leaner containers](https://cloud.google.com/build/docs/optimize-builds/building-leaner-containers).
Unfortunately, the example they give is for a Java service build using Gradle leaving some adaption work as an 
exercise for Go developers.

## First: Define a lean configuration Dockerfile

```dockerfile
# Use the official Debian slim image for a lean production container.
# https://hub.docker.com/_/debian
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM debian:buster-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY ./service /service

# Run the web service on container startup.
CMD ["/service"]
```

## Second: Amend your `cloudbuild.yaml` to reference the Dockerfile.

It's just a one line addition to the Docker build step to reference the Dockerfile after having compiled
the Go binary at `./service` (seen the green `+` line below:

```diff
  steps:
  # Copy the GitHub repository deploy SSH key from the secrets manager so that
  # the build will be able to retrieve sibling modules from our private repo.
  # Also, copy the GitHub domain name as a known host for SSH.
  - name: 'gcr.io/cloud-builders/git'
    secretEnv: ['SSH_KEY']
    entrypoint: 'bash'
    args:
      - -c
     - |
        echo "$$SSH_KEY" >> /root/.ssh/id_ed25519
        chmod 400 /root/.ssh/id_ed25519
        cp known_hosts.github /root/.ssh/known_hosts
        git config --global url."git@github.com:mikebway".insteadOf "https://github.com/mikebway"
    volumes:
      - name: 'ssh'
        path: /root/.ssh
      - name: 'git'
        path: /root/.gitconfig

  # Build the service binary
  - name: 'golang:1.19-buster'
    args: ['go', 'build', '-v', '-o', 'service']
    volumes:
      - name: 'ssh'
        path: /root/.ssh
      - name: 'git'
        path: /root/.gitconfig

  # Build the docker image
  - name: 'gcr.io/cloud-builders/docker'
    args: [
        'build',
        '-t', 'gcr.io/$PROJECT_ID/cart-service',
+       '-f', 'Dockerfile',
        '.']

  # Push our generated image to teh container registry
  images:
    - 'gcr.io/$PROJECT_ID/cart-service'

  # Fetch the SSH private key that allows "deploy" read access to sibling
  # modules in our monorepo.
  availableSecrets:
    secretManager:
      - versionName: projects/$PROJECT_ID/secrets/${PROJECT_ID}_deploy/versions/latest
        env: 'SSH_KEY'

```
