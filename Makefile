#See: https://cloud.google.com/functions/docs/2nd-gen/getting-started#pubsub

REGION=asia-northeast1
PROJECT_ID=$(shell gcloud config get-value project)
PROJECT_NUMBER=$(shell gcloud projects list --filter="project_id:$(PROJECT_ID)" --format='value(project_number)')

FUNC_NAME=fetch
ENTRY_POINT=Fetch
TOPIC_NAME=$(FUNC_NAME)-topic
BUCKET_NAME=$(PROJECT_ID)-fetch

start:
	export FUNCTION_TARGET=$(ENTRY_POINT) && \
	go run cmd/main.go

init:
	gcloud projects add-iam-policy-binding $(PROJECT_ID) \
    --member=serviceAccount:service-$(PROJECT_NUMBER)@gcp-sa-pubsub.iam.gserviceaccount.com \
    --role=roles/iam.serviceAccountTokenCreator
	gcloud pubsub topics create $(TOPIC_NAME)
	gsutil mb -c regional -l $(REGION) gs://$(BUCKET_NAME)

deploy:
	gcloud beta functions deploy $(FUNC_NAME) \
    --gen2 \
    --runtime go116 \
    --trigger-topic $(FUNC_NAME)-topic \
    --entry-point $(ENTRY_POINT) \
    --source .

show:
	gcloud beta functions describe $(FUNC_NAME) --gen2

URL=""
send:
ifeq ($(URL),)
	$(error "Please specify URL")
endif
	gcloud pubsub topics publish $(FUNC_NAME)-topic \
	--attribute=url=$(URL),bucket=$(BUCKET_NAME)

log:
	gcloud beta functions logs read $(FUNC_NAME) --gen2 --limit=100

open:
	open https://console.cloud.google.com/functions/details/asia-northeast1/$(FUNC_NAME)?env=gen2
	open https://console.cloud.google.com/cloudpubsub/topic/detail/$(FUNC_NAME)-topic
