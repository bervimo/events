#!/bin/sh

SERVICE_ACCOUNT=`bash cmd/get_service_account.sh ${SERVICE_NAME}`
APP_VERSION=`bash cmd/get_last_version.sh`
ENV_VARS=`bash cmd/get_env_vars.sh`
DELIMITER="^${GOOGLE_DELIMITER:-;}^"

# Build and push image
gcloud builds submit --tag gcr.io/${GOOGLE_PROJECT_ID}/${SERVICE_NAME}:${APP_VERSION}

# Deploy image
gcloud run deploy ${SERVICE_NAME} --image gcr.io/${GOOGLE_PROJECT_ID}/${SERVICE_NAME}:${APP_VERSION} --set-env-vars ${DELIMITER}${ENV_VARS} --platform managed --region ${GOOGLE_REGION} --service-account ${SERVICE_ACCOUNT} --quiet

# Update traffic
gcloud run services update-traffic ${SERVICE_NAME} --to-latest --platform managed --region ${GOOGLE_REGION} --quiet
