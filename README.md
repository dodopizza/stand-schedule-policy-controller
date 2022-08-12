# StandSchedulePolicy Controller

Controller that evicts all pods from target namespaces and optionally stops external resources.

At the moment the following external resources supported:

* `Azure`
  * ManagedMySQL databases
  * Virtual Machines

## Usage

Install in Kubernetes cluster as standalone deployment and create custom resource:

```yaml
apiVersion: automation.dodois.io/v1
kind: StandSchedulePolicy
metadata:
  name: test-policy-name
spec:
  resources: {}
  schedules:
    shutdown:
      cron: 0 19 * * *
    startup:
      cron: 30 4 * * 1-5
  targetNamespaceFilter: ^default$

```

This schedule will perform cleanup for default namespace every day at 19:00 (UTC) and startup at 4:30 (UTC) in working days.

Also, available kubernetes plugin to perform startup-shutdown actions on demand.
You can find the latest release on repository release page.

With plugin, to perform shutdown for existing policy spec, run:

```bash
kubectl stand shutdown test-policy-name
```

Optionally, with wait flag:

```bash
kubectl stand shutdown test-policy-name --wait
```

Plugin will wait until specified policy will be in Completed/Failed status

## How it works

* For shutdown action, controller will:
  * Scale down all deployments and statefulsets to zero replicas 
  * Create resource quota with zero pods spec
  * Deletes all existing pods
  * Stops all matching external resources
* For startup action, controller will:
  * Starts all matching external resources
  * Deletes resource quota
  * Scale up all deployments and statefulsets to previous value

## Development

To run all kinds of checks and generators please use:

```bash
make prepare
```

### Prerequisites

* golang
* docker
* kind

### Testing

#### Unit tests

```bash
make test-unit
```

#### Integration tests

```bash
make test-integration-setup
make test-integration
```

To perform cleanup:

```bash
make test-inregration-cleanup
```
