# Allow Cloud Functions to Invoke Cloud Functions

In this project, the [Fulfillment Task Distribution Function](taskdistrib/README.md) directly invokes the 
[Fulfillment Task Email Function](taskdistrib/README.md) function. Actually, it invokes one of four
different deployments of the [Fulfillment Task Email Function](taskdistrib/README.md) with different environment
variable configurations. For this to be possible, a couple of conditions have to be met:

1. [Ingress settings](https://cloud.google.com/functions/docs/networking/network-settings#ingress_settings) for the  
   target function(s) must be configured as `all`. 
   ```shell
   gcloud functions deploy task-gy-man --set-env-vars FULFILL_OPERATION=manufacture-gold-yoyo \
       --gen2 --region us-central1 --runtime go119 \
       --entry-point=FulfillTask --trigger-http --ingress-settings=all
   ```

2. The target function(s) must grant a role for the service account of the  [Fulfillment Task Distribution Function](taskdistrib/README.md) 
   function to invoke it. For 1st gen functions, the invoker role is Cloud Functions Invoker (roles/cloudfunctions.invoker). 
   For 2nd gen functions, the invoker role is Cloud Run Invoker (roles/run.invoker). 
   
   Command to grant access to 2nd gen calling functions:
   ```shell
   gcloud run services add-iam-policy-binding task-gy-man --region=us-central1 \
      --member='serviceAccount:[account-number]-compute@developer.gserviceaccount.com' \
      --role='roles/run.invoker'
   ```
   Where `--member` is the service account email address for the calling Cloud Function in the form
   `[account-number]-compute@developer.gserviceaccount.com` prefixed by `serviceAccount:`.
   
   Command to grant access to 1st gen calling functions:
   ```shell
   gcloud functions add-iam-policy-binding task-gy-man --region=us-central1 \
      --member='serviceAccount:[account-number]-compute@developer.gserviceaccount.com' \
      --role='roles/cloudfunctions.invoker'
   ```

For more background see [Cloud Functions: Authenticating for invocation](https://cloud.google.com/functions/docs/securing/authenticating)
and [Troubleshooting Cloud Functions](https://cloud.google.com/functions/docs/troubleshooting).

## Unit Testing The Invoking Function

Since the [Fulfillment Task Distribution Function](taskdistrib/README.md) includes code to obtain a OAuth token,
and CloudEvents requires that this token must be either for a service account or a service account impersonator,
a couple of hoops have to be jumped in order for unit tests to be run successfully.

First, create a service account key (for a service account with the right to invoke the target Cloud functions)
as a JSON file on the local machine where the tests are to be run:

```shell
gcloud iam service-accounts keys create ~/temp/serviceAccount.json --iam-account=SERVICE-ACCOUNT-EMAIL-ADDRESS
```

The service account that you specify can be anything, even a service account with no access rights whatsoever.
All that the unit tests require is credentials for something that is a service account or service account impersonator
since the invocations are all handled by an httptest mock server.

Next, the `GOOGLE_APPLICATION_CREDENTIALS` environment variable must be set to point to that JSON file:

```shell
export GOOGLE_APPLICATION_CREDENTIALS=~/temp/serviceAccount.json
```

Now, you can run the unit tests.