# gcf-fetch

Cloud Functions that fetch data from public APIs and store it in Google Cloud Storage.<br>
[Google Cloud Functions Gen2](https://cloud.google.com/functions/docs/2nd-gen/overview) Golang environment is supported.

# Features

- Fetch data is stored in a path based on the API's URL. ( like [ghq](https://github.com/x-motemen/ghq) )
  - Therefore, even if you get various API data, you do not need to manage the bucket's path.
- Execute Cloud Pub/Sub as a trigger.
  - Therefore, periodic execution is also possible with [Cloud Scheduler](https://cloud.google.com/scheduler).
- Fetch data is managed by [Object Versioning](https://cloud.google.com/storage/docs/object-versioning).
- GCS [price](https://cloud.google.com/storage/pricing) are optimized by [Object Lifecycle Management](https://cloud.google.com/storage/docs/lifecycle).
  - The storage class is set to change from Standard to Coldline after 7 days from object creation and from Coldline to Archive after 30 days (easy to change).
- [zl](https://github.com/nkmr-jp/zl) (zap based logger) for logging by level.
  - Logs in JSON format so it can check the element contents in detail with [Cloud Logging](https://cloud.google.com/logging).
  - You can also check the contents of CloudEvents triggered by Functions in the log.


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

Historical data can be created by periodically executing data acquisition from various APIs. <br>
Data stored in GCS can be loaded into BigQuery and other applications for data analysis and machine learning.

Of course, it can also be used in applications.

List of public APIs.<br>
[GitHub - public-apis/public-apis: A collective list of free APIs](https://github.com/public-apis/public-apis)

# See
- https://github.com/GoogleCloudPlatform/golang-samples/tree/main/functions/functionsv2
- https://cloud.google.com/functions/docs/2nd-gen/getting-started#pubsub
