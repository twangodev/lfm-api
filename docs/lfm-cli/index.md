---
hide:
    - navigation
---

# lfm-cli
![GitHub Release](https://img.shields.io/github/v/release/twangodev/lfm-cli)
![Build Status](https://img.shields.io/github/actions/workflow/status/twangodev/lfm-cli/build.yml?branch=master)
![Supported Platforms](https://img.shields.io/badge/Platforms-Windows%2C%20MacOS%2C%20Linux-orange)
![GitHub License](https://img.shields.io/github/license/twangodev/lfm-cli)

Show your fellow gamers and friends what you're listening to on Last.FM without touching a single API Key!

lfm-cli is a command-line interface implementing `lfm-api` to have the active scrobble displayed on [Discord Rich Presence](https://discord.com/rich-presence).

## Installation

You can install `lfm-cli` by downloading prebuilt binaries from the [releases page](https://github.com/twangodev/lfm-cli/releases) or by building from source.

??? question "Building from Source"

    To build from source, you will need to have Go installed on your system. To install [Go](https://golang.org/), follow the instructions on the [official website](https://golang.org/doc/install).

    Once you have Go installed, you can clone the repository onto your system to proceed with the build.

    ```bash
    git clone https://github.com/twangodev/lfm-cli.git
    ```

    After cloning the repository, navigate to the directory and build using the following command:

    ```bash
    go build
    ```

    This will create an executable file in the current directory that you can run to use to run `lfm-cli`.

## Usage

`lfm-cli` works right out of the box - no configuration needed. **You just need to have Discord open and running on your system.**

To run `lfm-cli`, simply execute the binary file you downloaded or built. 

```bash
./lfm-cli -u YOUR_LASTFM_USERNAME
```

The active scrobble will be displayed on your Discord profile.

!!! warning "Known Issues"
    - If you are using the Discord web client, the rich presence will not be displayed. This is a limitation of the Discord web client and not `lfm-cli`.
    - If you are using the Discord desktop client, ensure that you have the "Display Spotify as your status" option disabled in the Discord settings. This option can interfere with the rich presence display.
    - If Discord is closed or not running, the rich presence will not be displayed. Ensure that Discord is running in the background for the rich presence to work. Otherwise, restart `lfm-cli` when Discord is running.

### Available Flags

- `--help`, `-h`: Shows a list of commands or help for one command
- `--user USERNAME`, `-u USERNAME`: Display Last.FM scrobbles from `USERNAME`
- `--refresh X`, `-r X`: Checks Last.FM every X seconds for new scrobbles (default: `10`)
- `--hide-profile`: Removes buttons to the specified Last.FM profile (default: `false`)
- `--show-loved, -l`: Replaces the default smallImage key with a heart for loved songs. (default: `false`)
- `--rm-covers`: Does not show album cover images. (default: `false`)
- `--rm-time`: Does not show time elapsed for the scrobble. (default: `false`)
- `--keep-status`: Shows status even when there is no active scrobble. (default: `false`)
- `--debug`, `-d`: Enable verbose and debug logging (default: `false`)
- `--version`, `-v`: Print the version


