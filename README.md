# vitess-releaser

## What is vitess-releaser?

vitess-releaser is a tool designed to automate the process of releasing new versions of Vitess. 
It streamlines various tasks, such as version bumping, changelog generation, issue creation and everything else that is involved in creating a Vitess release.

## Install

```bash
# Clone the repo
> gh repo clone vitessio/vitess-releaser

# Install dependencies
> gh extension install valeriobelli/gh-milestone

# Install the binary
> cd vitess-releaser
> go install ./go/vitess-releaser.go
```

## Usage
```
Tooling used to release new versions of Vitess

Usage:
  vitess-releaser [flags]

Flags:
  -d, --date string           Date of the release with the format: YYYY-MM-DD. Required when initiating a release.
  -h, --help                  Displays this help.
      --live                  If live is true, will run against vitessio/vitess and planetscale/vitess-operator. Otherwise everything is done against your own forks.
      --rc int                Define the release as an RC release, value is used to determine the number of the RC.
  -r, --release string        Number of the major release on which we want to create a new release.
      --vtop-release string   Number of the major and minor release on which we want to create a new release, i.e. '2.11', leave empty for no vtop release.
```

## Examples on how to run a release

### RC release

In this example we are releasing `v19.0.0-rc1` of vitess and `v2.12.0-rc1` of vitess-operator. If you want to do an RC-2, or above, you can just change the value of the `--rc` flag.

```bash
vitess-releaser --date="2024-02-07" --live --rc=1 --release=19 --vtop-release=2.12
```

> [!IMPORTANT]
> RC releases are, in most cases, shipped with an equivalent vitess-operator release. We must set the `--vtop-release` flag in this case.

### Patch release

In this example we are releasing `v18.0.3` of vitess, and there are no vitess-operator release.

```bash
vitess-releaser --date="2024-02-07" --live --release=18
```

## Documentation

The goal of vitess-releaser is to be more and more self-documenting.
In the meantime you can find the full and original documentation used to run releases [in the vitess repository](https://github.com/vitessio/vitess/tree/main/doc/internal/release).

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
