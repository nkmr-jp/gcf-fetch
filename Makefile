# See:
# https://cloud.google.com/functions/docs/tutorials/pubsub
# https://cloud.google.com/storage/docs/lifecycle-configurations#delete-objects-json

REGION=asia-northeast1
PROJECT_ID=$(shell gcloud config get-value project)
PROJECT_NUMBER=$(shell gcloud projects list --filter="project_id:$(PROJECT_ID)" --format='value(project_number)')

FUNC_NAME=fetch
ENTRY_POINT=Fetch
TOPIC_NAME=$(FUNC_NAME)-topic
BUCKET_NAME=$(PROJECT_ID)-$(FUNC_NAME)
# VERSION=$(shell git rev-parse --short HEAD)
VERSION=$(shell git describe --abbrev=0 --tags)

init:
	@echo
	@echo "---- create pubusub topic. ----"
	-gcloud pubsub topics create $(TOPIC_NAME)
	@echo
	@echo "---- create bucket and set versioning. ----"
	-gsutil mb -c regional -l $(REGION) gs://$(BUCKET_NAME)
	-gsutil versioning set on gs://$(BUCKET_NAME)
	-gsutil lifecycle set ./lifecycle.json gs://$(BUCKET_NAME)
	@echo
	@echo "---- create bucket and set versioning, for test. ----"
	-gsutil mb -c regional -l $(REGION) gs://$(BUCKET_NAME)-test
	-gsutil versioning set on gs://$(BUCKET_NAME)-test
	-gsutil lifecycle set ./lifecycle.json gs://$(BUCKET_NAME)-test
	@echo
	@echo "---- check resources in google cloud console. ----"
	open https://console.cloud.google.com/cloudpubsub/topic/detail/$(FUNC_NAME)-topic
	open https://console.cloud.google.com/storage/browser?project=$(PROJECT_ID)


# See: https://cloud.google.com/pubsub/docs/push
# > If your project was created on or before April 8, 2021,
# > you must grant the roles/iam.serviceAccountTokenCreator role to
# > the Google-managed service account service-{PROJECT_NUMBER}@gcp-sa-pubsub.iam.gserviceaccount.com
# > on the project in order to allow Pub/Sub to create tokens.
add-iam-policy-binding:
	@echo
	@echo "---- add iam policy binding. ----"
	-gcloud projects add-iam-policy-binding $(PROJECT_ID) \
	--member=serviceAccount:service-$(PROJECT_NUMBER)@gcp-sa-pubsub.iam.gserviceaccount.com \
	--role=roles/iam.serviceAccountTokenCreator

test:
	export BUCKET_NAME=$(BUCKET_NAME)-test && go test -v

deploy:
	gcloud functions deploy $(FUNC_NAME) \
	--gen2 \
	--runtime=go116 \
	--region=$(REGION) \
	--trigger-topic=$(FUNC_NAME)-topic \
	--entry-point=$(ENTRY_POINT) \
	--set-env-vars BUCKET_NAME=$(BUCKET_NAME),VERSION=$(VERSION) \
	--source=.

show:
	gcloud functions describe $(FUNC_NAME) --gen2

URL=""
send:
ifeq ($(URL),)
	$(error "Please specify URL")
endif
	gcloud pubsub topics publish \
	$(FUNC_NAME)-topic \
	--project $(PROJECT_ID) \
	--message=$(URL)

log:
	gcloud functions logs read $(FUNC_NAME) --gen2 --limit=100

open:
	open https://console.cloud.google.com/storage/browser?project=$(PROJECT_ID)
	open https://console.cloud.google.com/functions/details/$(REGION)/$(FUNC_NAME)?env=gen2
	open "https://console.cloud.google.com/logs/query;query=%2528resource.type%20%3D%20%22cloud_function%22%0Aresource.labels.function_name%20%3D%20%22$(FUNC_NAME)%22%0Aresource.labels.region%20%3D%20%22$(REGION)%22%2529%0A%20OR%20%0A%2528resource.type%20%3D%20%22cloud_run_revision%22%0Aresource.labels.service_name%20%3D%20%22$(FUNC_NAME)%22%0Aresource.labels.location%20%3D%20%22$(REGION)%22%2529%0A%20severity%3E%3DDEFAULT;?project=$(PROJECT_ID)"

lint:
	golangci-lint run --fix