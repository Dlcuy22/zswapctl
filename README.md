# Zswapctl

## What is this?

Zswapctl is a port of [zswap-cli](https://github.com/xvitaly/zswap-cli) written in Go. It has similar features and command-line usage to zswap-cli, rewritten in Go for ease of deployment and code simplicity.

## Installation

You can run the one-line installer script:

```bash
curl -sSL https://raw.githubusercontent.com/Dlcuy22/zswapctl/main/scripts/install.sh | bash
```

Alternatively, you can download pre-built binaries from the GitHub release page [here](https://github.com/Dlcuy22/zswapctl/releases), or compile and install from source:

```bash
go install github.com/Dlcuy22/zswapctl@latest
```

This will install `zswapctl` into your `~/go/bin/`. Make sure it is in your `PATH`.

> [!NOTE]
> Like zswap-cli, zswapctl needs root access in order to work properly.

## Extensions

Additional flags in `zswapctl` compared to the original `zswap-cli`:

- `--install`: Automatically creates the default config `/etc/zswapctl/zswapctl.conf` and systemd service `/etc/systemd/system/zswapctl.service`, then reloads the systemd daemon, enables, and starts `zswapctl.service`.

## Documentation

Since it is a port, this [documentation](https://github.com/xvitaly/zswap-cli/blob/master/docs/README.md) is still relevant.

## License

This project is licensed under the terms of the MIT license.
