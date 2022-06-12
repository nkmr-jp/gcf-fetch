# gcf-fetch

Cloud Functions that fetch data from public APIs and store it in Google Cloud Storage.

- Runtime: Go 1.16
- Gen: [Cloud Functions (2nd gen)](https://cloud.google.com/functions/docs/2nd-gen/overview)


# Features

- Fetch data is stored in a path based on the API's URL. ( like [ghq](https://github.com/x-motemen/ghq) )
  - Therefore, even if you get various API data, you do not need to manage the bucket's path.
- Execute Cloud Pub/Sub as a trigger.
  - Therefore, periodic execution is also possible with [Cloud Scheduler](https://cloud.google.com/scheduler).
- Fetch data is managed by [Object Versioning](https://cloud.google.com/storage/docs/object-versioning).
- GCS [price](https://cloud.google.com/storage/pricing) are optimized by [Object Lifecycle Management](https://cloud.google.com/storage/docs/lifecycle).
  - The storage class is set to change from Standard to Coldline after 7 days from object creation and from Coldline to Archive after 30 days (easy to change).
- [zl](https://github.com/nkmr-jp/zl) (zap based logger) for logging by severity level.
  - Logs in JSON format so it can check the element contents in detail with [Cloud Logging](https://cloud.google.com/logging).
  - You can also check the contents of CloudEvents triggered by Functions in the log.
  
<details>
<summary>Cloud Logging's log example</summary>

```json

{
  "insertId": "xxxxxxxxxxxxxxxxxxxxxx",
  "jsonPayload": {
    "timestamp": "2022-06-12T00:45:17.427119741Z",
    "function": "github.com/nkmr-jp/gcf-fetch.parseEvent",
    "cloudEventContext": "Context Attributes,\n  specversion: 1.0\n  type: google.cloud.pubsub.topic.v1.messagePublished\n  source: //pubsub.googleapis.com/projects/[your project id]/topics/fetch-topic\n  id: xxxxxxxxxxx\n  time: 2022-06-12T00:45:14.378Z\n  datacontenttype: application/json\n",
    "cloudEventData": {
      "subscription": "projects/[your project id]/subscriptions/eventarc-asia-northeast1-fetch-xxxxx-sub-xxx",
      "message": {
        "publishTime": "2022-06-12T00:45:14.378Z",
        "messageId": "4863463195745766",
        "data": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
      }
    },
    "caller": "https://github.com/nkmr-jp/gcf-fetch/blob/v1.0.0/fetch.go#L89",
    "version": "v1.0.0",
    "message": "CLOUD_EVENT_RECEIVED"
  },
  "resource": {
    "type": "cloud_run_revision",
    "labels": {
      "service_name": "fetch",
      "project_id": "[your project id]",
      "configuration_name": "fetch",
      "revision_name": "fetch-xxxx-xiv",
      "location": "asia-northeast1"
    }
  },
  "timestamp": "2022-06-12T00:45:17.427272Z",
  "severity": "INFO",
  "labels": {
    "goog-managed-by": "cloudfunctions",
    "instanceId": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  },
  "logName": "projects/[your project id]/logs/run.googleapis.com%2Fstderr",
  "receiveTimestamp": "2022-06-12T00:45:17.670946465Z"
}
```

</details>

# Prepare

If you haven't already, install and set up the [Cloud SDK](https://cloud.google.com/sdk/docs/install-sdk).

# Usage

Create GCP resources
```sh
make init
```

Run test
```sh
make test
```

Deploy to google cloud functions
```sh
make deploy
```

Send pub/sub event
```sh
make send URL="https://api.github.com/users/github"
```

Open resources in console
```sh
make open
```

# Use Case

The api data saved in GCS can be used for various purposes such as data analysis and machine learning by loading it into BigQuery.<br>
Of course, it can also be used in applications.

List of public APIs.<br>
[GitHub - public-apis/public-apis: A collective list of free APIs](https://github.com/public-apis/public-apis)

# See
- https://github.com/GoogleCloudPlatform/golang-samples/tree/main/functions/functionsv2
- https://cloud.google.com/functions/docs/2nd-gen/getting-started#pubsub
