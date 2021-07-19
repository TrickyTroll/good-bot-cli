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

### Manual

General instructions on how to use `good-bot-cli`. Some more
tourough documentation is [coming soon](https://github.com/TrickyTroll/good-bot-cli/issues/10).

#### Available commands

##### `configure`

`configure` allows you to interactively set up the variables required
by Good Bot to use the TTS engine and to use passwords. After entering
the command, an interactive prompt will ask you to provide the path
towards your text to speech key and password file.

If Good Bot had already been configured, this command will overwrite
the previous parameters.

When using another Good Bot command, the `configure` command will run
automatically if no configuration file is found.

##### `echoConfig`

The `echoConfig` simply outputs the configuration file. Can be used
to know which credential and password files will be used.

##### `record`

The record command uses a project directory previously created by
the [`setup`](#setup) command to record a video. This command takes
care of typing the command while recording and sending the text files
to Google Text to Speech to generate audio. When executed with no
flags, this command also merges the videos produced for each scene
in a final video.

There are two additional options when using the `record` command:

* `--no-render`: Only produces Asciinema recordings and `mp3` narration.
  Does not convert the recordings to the gif format and does not create
  any `mp4` file.

* `--gifs-only`: Only convert the Asciinema recordings to the `gif`
  format. The final product will contain the converted gifs, the
  original `asciicasts`, and the audio narration. No `mp4` file will
  be created.

##### `render`

`render` uses a project that has been recorded but not rendered yet
and renders a video from it. It is essentially the missing part of the
`record --no-render` command. This command converts the `asciicasts`
to the `gif` format and merges the audio and video files to create 
`mp4` files from your project.

##### `setup`

This command uses your script (the YAML instruction file you wrote)
to create a directory that contains the information required for the
`record` command. It essentially splits your config file into many
smaller files that will be used by Good Bot's modules
(the [runner](https://github.com/TrickyTroll/good-bot-runner), Google
TTS, [ezvi](https://github.com/TrickyTroll/ezvi)).

You need to provide the path towards the script as an argument.

For more information on writing scripts, see
[Writing scripts](#writing-scripts).

#### Writing scripts

## Motivation

Before writing `good-bot-cli`, Good Bot was only distributed as a
Docker image. This allows the program to always start recording in a 
new environment with nothing installed. You won't need to manually 
uninstall every program that was installed for your demo each time
you want to record a new take.

The main problem with containerized application is that it makes file
and environment sharing between the host and the container quite
tedious.

When using Google Text to Speech and passwords at the same time, the
`record` command had to look something like this:

```shell
docker run -it -v $PWD:/project -v /home/tricky/Documents/credentials:/credentials --env GOOGLE_APPLICATION_CREDENTIALS=/credentials/good-bot-tts.json --env-file /home/tricky/Documents/credentials/good-bot.env gb record project
```

Two hundred and sixty-six characters for a single command wasn't very reasonable.

`good-bot-cli` allows a user to interactively configure each path
towards files that are required by Good Bot. The CLI also takes
care of making the required mounts and starts the container in
interactive mode when required. Once the configuration process is
completed, `good-bot-cli` would tun the previous command to this:

```shell
good-bot-cli record project
```
