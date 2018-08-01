# GCE Operations Notification

## Abstract
It will check and send mail after GCE instance operations (live migration and host error) happened.

## Setup GAE

Create a new project
1. Choose App Engine.
2. Choose language for development, we using `Go`.
3. Choose deploy location.
4. You can skip the tutorial.
5. Refresh the page, and you can see the GAE main page.

## Setup IAM
1. Copy the GAE service account from `GCP -> IAM`
2. Paste the GAE service account to create new member from `GCP -> IAM` on an other projects you want to follow

## Setup App
- app.yaml
  - change `PROJECT_IDS` to you want to follow
  - change `TO` to you want to send
  - choose `DATABASE` to you want to use, you can choes GCP `datastore` or GAE `memcache`
- cron.yaml
  - change schedule for Period

## Deploy
```
$ gcloud config set project <NEW_PROJECT_ID_YOU_CREATED_FOR_THIS_CODE>
$ gcloud app deploy app.yaml cron.yaml queue.yaml
```
