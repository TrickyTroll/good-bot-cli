# good-bot-cli

Good-bot's command line interface for Docker containers.

## Installation

### Binary download

Coming  soon.

### Go install

> **Note**: You will need to have the Go Toolchain installed to
> use this method.

> **Note**: With this installation method, `good-bot-cli` will
> be installed in your `$GOPATH`. If your `$GOPATH` has not
> been added to your `PATH`, the executable won’t be available
> from everywhere on your machine.

#### Quick install

```shell
go install github.com/TrickyTroll/good-bot-cli.
```

#### Go toolchain

To install `good-bot-cli` in your `GOPATH`, you should run
the following command.

```shell
go install github.com/TrickyTroll/good-bot-cli
```

#### Cloning the repo

You can install `good-bot-cli` by cloning this repository.

```shell
git clone https://github.com/TrickyTroll/good-bot-cli.git
```

You can then:

1. Navigate to the root directory of the repository you just
   cloned.
2. Install the app with the Go toolchain.
   ```shell
   go install .
   ```

#### Downloading the repository

You can also [download](https://github.com/TrickyTroll/good-bot-cli/archive/refs/heads/main.zip)
the repository. This will download a `zip` archive. You should then
be able to:

1. Unzip the archive.
2. Navigate to the root of the unzipped archive.
3. Install the app with the Go toolchain.
   ```shell
   go install .
   ```

## Usage

### TLDR

Good Bot has two main commands, `setup` and `record`.
You **always** need to run `setup` before the `record` command
if you are working on a new project.

#### Using `setup`

First, navigate to the directory where your video script is
saved. Then, to use setup:

##### On a Unix-like system

```shell
good-bot-cli setup [script-name.yaml]
```

This should prompt you for a project name. Good Bot will then
create a new directory that contains everything required by the
`record` command.

#### Using `record`

Make sure that you are in the same directory as the directory
that you created with the `setup` command. Then, you should
be able to record your project using the following instructions.

##### On a Unix-like system

```shell
good-bot-cli record [project-name]
```

This should start recording right away. Please note that the
`[project-name]` is the name of the directory created with
`setup`, not the name of your script.

## Motivation

