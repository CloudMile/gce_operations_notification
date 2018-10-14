# GCE Operations Notification

## Abstract

It's checking GCE instance operations (__live migration__ and __host error__) and send notification mail.

## Setup GAE

Create a new project
1. Choose App Engine.
2. Choose language for development, we using `Go`.
3. Choose deployment location.
4. You can skip the tutorial.
5. Refresh the page, and you can see the GAE dashboard.

## Setup IAM

When enable GAE API, the GCP project will create a service account for GAE automatically. We have to add the service account as a memberto other project which we want to monitor the GCE instance operation events.

1. Copy the GAE service account in `IAM & admin -> IAM` page.
2. Add the GAE service account with `Compute Viewer` role in `IAM & admin -> IAM` page of the project which you want to monitor.

You can give the GAE service account with minimal permission `compute.globalOperations.list` by creating a [custom role](https://cloud.google.com/iam/docs/creating-custom-roles).

## Setup App

- `app.yaml`
  - Replace `PROJECT_IDS` value to you want to monitor event, use comma to split prject ids.
  - Replace `TO` value to who will be notified when event occur, use comma to split email addresses. 
  - Replace `DATABASE` to you want to use, you can choes GCP `datastore` or GAE `memcache`
- `cron.yaml`
  - change schedule for Period

## Deploy

```shell
$ gcloud config set project <NEW_PROJECT_ID_YOU_CREATED_FOR_THIS_CODE>
$ gcloud app deploy app.yaml cron.yaml queue.yaml
```

Replace `<NEW_PROJECT_ID_YOU_CREATED_FOR_THIS_CODE>` to the project id which you want to deploy this GAE service.
