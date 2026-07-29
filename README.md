# Zswapctl 
## What is this?
Zswapctl is a port of [zswap-cli](https://github.com/xvitaly/zswap-cli) written in go,
it has similar feature and command line usage as zswap-cli.
rewritten in go for ease of deployment and code simplicity.

## Install
You can download the pre build binaries from the github release page [here](https://github.com/Dlcuy22/zswapctl/releases)
rename it to ```zswapctl``` and move it into your system's PATH
or you can compile and install from source 
make sure you have the go toolchain installed in your system
```
go install https://github.com/Dlcuy22/zswapctl 
``` 
it will install zswapctl into your ~/go/bin/ make sure its in your PATH 
> [!NOTE]
> like zswap-cli, zswapctl need root acces in order to work properly.

## License
This project is licensed under the terms of the MIT license.

## Documentation
since its a port this [documentation](https://github.com/xvitaly/zswap-cli/blob/master/docs/README.md) is still relevant.
