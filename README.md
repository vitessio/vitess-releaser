# vitess-releaser

## What is vitess-releaser?

vitess-releaser is a tool designed to automate the process of releasing new versions of Vitess. 
It streamlines various tasks, such as version bumping, changelog generation, issue creation and everything else that is involved in creating a Vitess release.

## Install

```bash
# Install dependencies
> gh extension install valeriobelli/gh-milestone

# Install the binary
> go install github.com/vitessio/vitess-releaser@latest
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

## Authenticate with GH

Each Pull Request either on `vitessio/vitess` or `planetscale/vitess-operator` require at least two approvals before getting merged.
When doing a release, the release team is either composed of two or three people depending on availability.
If the team is composed of two people, there can be a bottleneck around PR reviews. While one of the team member is the PR author, they
cannot review their own PR, and thus the flow is slowed down.

To counter this problem, one can set the environment variable `VITESS_RELEASER_GH_TOKEN` to a GitHub Personal Access Token with the `repo` and `org:read` permissions.
We usually use the `@vitess-bot` account to author our PRs, but any other account can be used.

## Examples on how to run a release

### RC release

In this example we are releasing `v19.0.0-rc1` of vitess and `v2.12.0-rc1` of vitess-operator. If you want to do an RC-2, or above, you can just change the value of the `--rc` flag.

```bash
vitess-releaser --date="2024-02-07" --live --rc=1 --release=19 --vtop-release=2.12
```

> [!NOTE]
> RC releases are, in most cases, shipped with an equivalent vitess-operator release. We must set the `--vtop-release` flag in this case.

----
### GA release

In this example we are releasing `v19.0.0` of vitess and `v2.12.0` of vitess-operator. We had previously done the RC on `release-19.0` and `release-2.12`.

```bash
vitess-releaser --date="2024-02-07" --live --release=19 --vtop-release=2.12
```

----
### Patch release

In this example we are releasing `v19.0.1` of vitess, and there are no vitess-operator release.

The version to release is defined based on the `SNAPSHOT` version of the release branch.
In this case, the `release-19.0` branch is on `19.0.1-SNAPSHOT` as we have just released the GA, which is why `v19.0.1` will be released with the following snippet.

```bash
vitess-releaser --date="2024-02-07" --live --release=19
```

> [!NOTE]
> We often do not include a patch release of vitess-operator during patches. But it depends on how much was changed on vtop's release branch.

## Documentation

The goal of vitess-releaser is to be more and more self-documenting.
In the meantime you can find the full and original documentation used to run releases [in the vitess repository](https://github.com/vitessio/vitess/tree/main/doc/internal/release).

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Versioning

When one or more new PRs are merged on this repository, we should cut a new release to ensure people keep using the latest version.
New releases can be published the following way:

1. Merge all opened PRs
2. Create a new tag, push it, and create an associated release on the GitHub UI
3. Create another PR that bumps the `VERSION` constant to the next release ([example PR here](https://github.com/vitessio/vitess-releaser/pull/124))