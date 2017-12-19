# Release Notes

Generates markdown-style release notes from the pull requests merged between
two given branches or tags

### Usage

```sh
go get -u github.com/sagansystems/release-notes
release-notes -since release-20171212014000Z # or whatever the last release was
```

Note: ensure `GITHUB_USER` and `GITHUB_TOKEN` are set in your environment and
that your `GITHUB_TOKEN` has access to `repos`.

![](http://take.ms/k8mfl)

At https://github.com/settings/tokens
