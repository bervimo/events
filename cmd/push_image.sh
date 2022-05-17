#!/bin/sh

# Parameters
GCR_IMAGE=gcr.io/${GOOGLE_PROJECT_ID}/k8s/${SERVICE_NAME}

# Build and push image
gcloud builds submit --tag ${GCR_IMAGE}:${APP_VERSION}
gcloud container images add-tag ${GCR_IMAGE}:${APP_VERSION} ${GCR_IMAGE}:latest --quiet
