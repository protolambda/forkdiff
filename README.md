# forkdiff

**experimental, WORK IN PROGRESS**

Forkdiff is a simple CLI tool to diff a git fork and its base,
structure and enhance it with descriptions based on a `fork.yaml`,
and output a `index.html` to show the diff.

Usage:

```
-repo string
    path to local git repository (default ".")

-base string
    base reference to diff against (default "master")
-target string
    target reference to retrieve diff for (default "HEAD")

-fork string
    fork page definition (default "fork.yaml")
-out string
    output (default "index.html")
```

The `fork.yaml` defines the page structure, to organize and document the diff of the fork.

Example:

```yaml
title: "Example Fork"  # Define the HTML page title
def:
    title: "Example Fork diff"
    description: | # description in markdown
      These are some **really** important `code` modifications in the fork.
      The original can be found at [`github.com/example/greeter`](https://github.com/example/greeter).
      And the fork at [`github.com/protolambda/greeter`](https://github.com/protolambda/greeter).
    globs:
      - "hello/world/greeter.go"  # list files of which the patches should be included
      - "hello/util/*"  # use file globs to include multiple files
    sub:
      - title: "fork testing"  # titles are optional
        description: "This fork tests the modifications to `greeter.go`"
        globs:
          - "hello/world/greeter.go"
      - title: "modifications to hello/printer"
        description: "The `printer` package prints greetings"
        globs:
          - "hello/printer/*"
      - title: "MOTD"
        description: "New package that generates a message of the day (MOTD) to add to the greeting"
        globs:
          - "motd/*"
```

## License

MIT, see [`LICENSE` file](./LICENSE).
