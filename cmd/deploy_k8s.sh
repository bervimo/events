#!/bin/sh

# Parameters
GCR_IMAGE=gcr.io/${GOOGLE_PROJECT_ID}/k8s/${SERVICE_NAME}

# Authentication with cluster
gcloud container clusters get-credentials ${CLUSTER_NAMESPACE_NAME} --zone ${CLUSTER_LOCATION} --project ${GOOGLE_PROJECT_ID}

# Install/Update deployment
helm upgrade --install --wait --timeout 1m ${SERVICE_NAME} ../../helm/core-service/ \
    -f k8s/values.${APP_ENV}.yaml \
    --set fullnameOverride=${SERVICE_NAME} \
    --set app.name=${SERVICE_NAME} \
    --set app.version=${APP_VERSION} \
    --set image.repository=${APP_IMAGE} \
    --set image.tag=${APP_VERSION} \
    --set env.SERVICE_NAME=${SERVICE_NAME} \
    --set env.SERVICE_VERSION=${APP_VERSION} \
    --set env.K8S_CLUSTER_NAME=${CLUSTER_NAME} \
    --set env.K8S_NAMESPACE_NAME=${CLUSTER_NAMESPACE_NAME:-default} \
    --set env.K8S_CLUSTER_LOCATION=${CLUSTER_LOCATION}
