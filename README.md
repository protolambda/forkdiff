# forkdiff

Forkdiff is a simple CLI tool to diff a git fork and its base,
structure and enhance it with descriptions based on a `fork.yaml`,
and output a `index.html` to show the diff.

Usage:

```
-repo string
    path to local git repository (default ".")
-fork string
    fork page definition (default "fork.yaml")
-out string
    output (default "index.html")
```

The `fork.yaml` defines the page structure, to organize and document the diff of the fork.

Example:

```yaml
title: "protolambda's Greeter fork"  # Define the HTML page title
footer: |  # define the footer with markdown
  [Greeter](https://github.com/protolambda/greeter) fork overview &middot created with [Forkdiff](https://github.com/protolambda/forkdiff)
base:
  name: example/greeter
  url: https://github.com/example/greeter
  ref: refs/heads/master
fork:
  name: protolambda/greeter
  url: https://github.com/protolambda/greeter
  ref: refs/heads/optimism-history
def:
    title: "Example Fork diff"
    description: | # description in markdown
      These are some **really** important `code` modifications in the fork.
      The original can be found at [`github.com/example/greeter`](https://github.com/example/greeter).
      And the fork at [`github.com/protolambda/greeter`](https://github.com/protolambda/greeter).
    globs:
      - "hello/world/greeter.go"  # list files of which the patches should be included
      - "hello/util/*"  # use file globs to include multiple files
      - "hello/util/*[!_test].go"  # you can ignore things with globs too
    sub:
      - title: ""  # titles are optional
        description: "This fork tests the modifications to `greeter.go` and utils."
        globs:
          - "hello/world/greeter.go"
          - "hello/util/*_test.go"
      - title: "modifications to hello/printer"
        description: "The `printer` package prints greetings"
        globs:
          - "hello/printer/*"
      - title: "MOTD"
        description: "New package that generates a message of the day (MOTD) to add to the greeting"
        globs:
          - "motd/*"
# files can be ignored globally, these will be listed in a separate grayed-out section,
# and do not count towards the total line count.
ignore:
  - "*.sum"
```

## License

MIT, see [`LICENSE` file](./LICENSE).
