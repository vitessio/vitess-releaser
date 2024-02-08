# vitess-releaser

## What is vitess-releaser?

vitess-releaser is a tool designed to automate the process of releasing new versions of Vitess. 
It streamlines various tasks, such as version bumping, changelog generation, issue creation and everything else that is involved in creating a Vitess release.

### Install

```bash
# Clone the repo
> gh repo clone vitessio/vitess-releaser

# Install dependencies
> gh extension install valeriobelli/gh-milestone

# Install the binary
> cd vitess-releaser
> go install ./go/vitess-releaser.go
```

### Usage
```
Usage:
  vitess-releaser [command]

Available Commands:
  interactive Runs the releaser in interactive mode

Flags:
  -d, --date string           Date of the release with the format: YYYY-MM-DD. Required when initiating a release.
      --live                  If live is true, will run against vitessio/vitess. Otherwise everything is done against your personal repository
      --rc int                Define the release as an RC release, the integer value is used to determine the number of the RC.
  -r, --release string        Number of the major release on which we want to create a new release.
      --vtop-release string   Number of the major and minor release on which we want to create a new release, i.e. '2.11', leave empty for no vtop release.

Use "vitess-releaser [command] --help" for more information about a command.
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
