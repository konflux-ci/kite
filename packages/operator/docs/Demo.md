# Demo with Sample Resources
This demo shows how the Kite Bridge Operator can monitor, create and resolve issues as they occur in a cluster.

You'll run a sample `PipelineRun` that fails and the included Pipeline Run controller will report the failure to the [Kite backend](../backend/).

You'll then modify the Pipeline manifest so it succeeds and see the operator resolve the issue.

## Requirements
- [Tekton](https://tekton.dev/docs/installation/) and [Tekton CLI](https://tekton.dev/docs/cli/) installed
- A local instance of the [Kite backend service](../backend/README.md) running on `localhost:8080`

## Run the Demo
1. **Set the local development environment variables:**
```sh
export KITE_API_URL="http://localhost:8080"
export ENABLE_HTTP2=false
```

2. **Run the Operator locally (without deploying):**
```sh
make run
```

3. **Apply the sample manifests:**
```sh
kubectl apply -k config/samples/
```

4. **Verify PipelineRun finished and failed:**
```sh
kubectl get prs
```

You should see something like this:
```bash
NAME                  SUCCEEDED   REASON   STARTTIME   COMPLETIONTIME
simple-pipeline-run   False       Failed   9s          3s
```

5. **Observe the Operator logs:**
You should see the operator detect the failure and report it to Kite.
```sh
{"level":"info","msg":"Processing failed PipelineRun","namespace":"default","pipeline_run":"simple-pipeline-run","status":"failed","time":"2025-08-26T09:51:40-04:00"}
{"level":"info","msg":"Successfully sent request to KITE","operation":"pipeline-failure","status_code":201,"time":"2025-08-26T09:51:40-04:00"}
{"id":"c7686f52-afae-4180-b719-ae5d66b50379","level":"info","msg":"Successfully reported pipeline failure to KITE","operation":"pipeline-failure","pipeline_run":"simple-pipeline-run","time":"2025-08-26T09:51:40-04:00"}
```

6. **Update the [sample manifest](../config/samples/tekton_v1_pipelinerun.yaml) so the PipelineRun succeeds:**
```yaml
apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: simple-pipeline
  namespace: default
spec:
  tasks:
  - name: echo-task
    taskSpec:
      stepTemplate:
        securityContext:
          runAsNonRoot: true
          allowPrivilegeEscalation: false
          runAsUser: 1000
          capabilities:
            drop: ["ALL"]
          seccompProfile:
            type: RuntimeDefault
      steps:
      - name: echo-message
        image: busybox:1.36
        script: |
          #!/bin/sh
          echo "Hello, Tekton!"
          exit 0 # <- Update here so it passes
```

7. **Apply the update:**
```sh
kubectl apply -k config/samples
```

8. **Run the Pipeline again:**
```sh
tkn p start simple-pipeline
```

9. **Confirm it passed:**
```bash
kubectl get prs
```

You should see an output like this:
```bash
NAME                        SUCCEEDED   REASON      STARTTIME   COMPLETIONTIME
simple-pipeline-run         False       Failed      2m9s        2m3s
simple-pipeline-run-pbwxj   True        Succeeded   32s         28s
```

10. **Confirm success in the operator logs:**
```sh
{"level":"info","msg":"Processing successful PipelineRun","namespace":"default","pipeline_run":"simple-pipeline-run-bzskm","status":"succeeded","time":"2025-08-26T09:52:56-04:00"}
{"level":"info","msg":"Successfully sent request to KITE","operation":"pipeline-success","status_code":200,"time":"2025-08-26T09:52:56-04:00"}
{"id":"6d3ba54e-7293-40b0-8db1-ce5ee3b0a7a5","level":"info","msg":"Successfully reported pipeline success to KITE","operation":"pipeline-success","pipeline_run":"simple-pipeline-run-bzskm","time":"2025-08-26T09:52:56-04:00"}
```

11. **Stop the operator:**
You can stop it with `Ctrl-C`.

12. **Delete the sample resources:**
```sh
kubectl delete -k config/samples/
```
