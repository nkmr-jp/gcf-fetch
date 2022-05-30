# See:
# https://cloud.google.com/functions/docs/2nd-gen/getting-started#pubsub
# https://cloud.google.com/storage/docs/lifecycle-configurations#delete-objects-json

REGION=asia-northeast1
PROJECT_ID=$(shell gcloud config get-value project)
PROJECT_NUMBER=$(shell gcloud projects list --filter="project_id:$(PROJECT_ID)" --format='value(project_number)')

FUNC_NAME=fetch
ENTRY_POINT=Fetch
TOPIC_NAME=$(FUNC_NAME)-topic
BUCKET_NAME=$(PROJECT_ID)-fetch
VERSION=$(shell git rev-parse --short HEAD)

init:
	@echo
	@echo "---- add iam policy binding. ----"
	-gcloud projects add-iam-policy-binding $(PROJECT_ID) \
	--member=serviceAccount:service-$(PROJECT_NUMBER)@gcp-sa-pubsub.iam.gserviceaccount.com \
	--role=roles/iam.serviceAccountTokenCreator
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
	make open

test:
	export BUCKET_NAME=$(BUCKET_NAME)-test && go test -v

deploy:
	gcloud beta functions deploy $(FUNC_NAME) \
	--gen2 \
	--runtime go116 \
	--trigger-topic $(FUNC_NAME)-topic \
	--entry-point $(ENTRY_POINT) \
	--set-env-vars BUCKET_NAME=$(BUCKET_NAME),VERSION=$(VERSION) \
	--source .

show:
	gcloud beta functions describe $(FUNC_NAME) --gen2

URL=""
send:
ifeq ($(URL),)
	$(error "Please specify URL")
endif
	gcloud pubsub topics publish $(FUNC_NAME)-topic \
	--message=$(URL)

log:
	gcloud beta functions logs read $(FUNC_NAME) --gen2 --limit=100

open:
	open https://console.cloud.google.com/iam-admin/serviceaccounts?project=$(PROJECT_ID)
	open https://console.cloud.google.com/cloudpubsub/topic/detail/$(FUNC_NAME)-topic
	open https://console.cloud.google.com/storage/browser?project=$(PROJECT_ID)
	open https://console.cloud.google.com/functions/details/$(REGION)/$(FUNC_NAME)?env=gen2
