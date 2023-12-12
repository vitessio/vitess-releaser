# vitess-releaser

## What is vitess-releaser?

vitess-releaser is a tool designed to automate the process of releasing new versions of Vitess. 
It streamlines various tasks, such as version bumping, changelog generation, issue creation and everything else that is involved in creating a Vitess release.

### Getting started

```bash
# Install dependencies

> gh extension install valeriobelli/gh-milestone

# Install the binary

> go install ./go/vitess-releaser.go

# Move to the vitess git folder

> cd ~/dev/vitess

# Get going

vitess-releaser -h
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
