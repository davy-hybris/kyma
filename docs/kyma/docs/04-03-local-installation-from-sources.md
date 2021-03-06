---
title: Install Kyma locally from sources
type: Installation
---

This Installation guide shows developers how to quickly deploy Kyma on a Mac or Linux from local sources. Follow it if you want to use Kyma for development purposes.

Kyma installs locally using a proprietary installer based on a [Kubernetes operator](https://coreos.com/operators/). The document describes only the installation part. For prerequisites, certificates setup, deployment validation, and troubleshooting steps, see [this](#installation-install-kyma-locally-from-the-release) document.

## Install Kyma

To run Kyma locally, clone this Git repository to your machine.

To start the local installation, run the following command:

```
./installation/cmd/run.sh
```

This script sets up default parameters, starts Minikube, builds the Kyma Installer, generates local configuration, creates the Installation custom resource, and sets up the Installer.

> **NOTE:** See [this](#installation-local-installation-scripts-deep-dive) document for a detailed explanation of the `run.sh` script and the subscripts it triggers.

You can execute the `installation/cmd/run.sh` script with the following parameters:

- `--password {YOUR_PASSWORD}` which allows you to set a password for the **admin@kyma.cx** user.
- `--skip-minikube-start` which skips the execution of the `installation/scripts/minikube.sh` script.
- `--vm-driver` which points to either `virtualbox` or `hyperkit`, depending on your operating system.

Read [this](#installation-reinstall-kyma) document to learn how to reinstall Kyma without deleting the cluster from Minikube.

To learn how to test Kyma, see [this](#details-testing-kyma) document.
