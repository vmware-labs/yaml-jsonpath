# YAML JSONPath Evaluator Web Application

To run this locally, issue the following from the root of this repository:
```
GAE_VERSION=$(git rev-parse --short HEAD) go run ./web
```

Then navigate to [localhost:8080](http://localhost:8080).

## Deploy to the Google Cloud

To deploy to Google Application Engine, change to the `web` directory and issue:
```
gcloud app deploy --version=$(git rev-parse --short HEAD)
```

Alternatively to deploy without prompts, run `scripts/gcloud-deploy.sh`.

For more information, see [Building a Go App on App Engine](https://cloud.google.com/appengine/docs/standard/go/building-app).