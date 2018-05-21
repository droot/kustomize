## Steps to run build locally

Install container-builder-local from [github](https://github.com/GoogleCloudPlatform/container-builder-local). 

```sh
container-builder-local --config=build/cloudbuild_local.yaml --dryrun=false --write-workspace=/tmp/w .
```

You will find the build artifacts under `/tmp/w/workspace/dist` directory
