# Wollemi
Please build file generator and formatter capable of generating `go_binary`,
`go_library` and `go_test` build rules from existing go code while also
ensuring that unused dependencies are stripped from go build rules as the
underlying go code changes over time.

Wollemi currently does not generate third party `go_get` rules but may do so
in a future release. When the `gofmt` command is unable to find a third party
dependency to satisfy a go import it will issue an error with the message
`"could not resolve go import"`.  This can be fixed by defining a `go_get`
rule for the go import anywhere inside of the `third_party/go` directory.

## Demo
See [Vim](#vim) setup.

[![asciicast](https://asciinema.org/a/342181.svg)](https://asciinema.org/a/342181)

## Requirements
- [please](https://please.build)

### Install
```
GO111MODULE=on go get github.com/tcncloud/wollemi
```

Wollemi can also be installed by running the following install script from the
root of the repository which will build the binary using please.

```
./install.sh
```

Last, but not least you can download Wollemi with Please itself.

In your project create a file (eg. `tools/BUILD.plz`) with following content:

```starlark
WOLLEMI_VERSION = "v0.2.1"
remote_file(
    name = "wollemi",
    url = f"https://github.com/tcncloud/wollemi/releases/download/{WOLLEMI_VERSION}/wollemi-{WOLLEMI_VERSION}-{CONFIG.HOSTOS}-{CONFIG.HOSTARCH}.tar.gz",
    extract = True,
    binary = True,
)

```

**Note:** commands like `gofmt` require `GOROOT` to be set, so running Wollemi with `plz run` requires wrapper that sets `GOROOT` as an environment variable:

```starlark
build_rule(
    name = "wollemi-wrapper",
    outs = ["wollemi.sh"],
    binary = True,
    cmd = "cat > \"$OUT\" << EOF\n#!/bin/sh\nexport GOROOT=$(\"$TOOLS_GO\" env GOROOT)\n$(out_exe :wollemi) \\\\$@\nEOF\n",
    tools = {"go": [CONFIG.GO_TOOL]},
    visibility = ["PUBLIC"],
    deps = [":wollemi"],
)
```

### Install Bash Completion
The wollemi completion script for Bash can be generated with the command wollemi
completion bash. Sourcing the completion script in your shell enables wollemi
command auto-completion.

To do so in all your shell sessions, add the following to your ~/.bash_profile file:

```
source <(wollemi completion bash)
```

After reloading your shell, wollemi autocompletion should be working.

Wollemi completions require bash version 4.1 or higher and bash-completion@2. You
can check your version by running echo $BASH_VERSION. If your version is too old
and you are using macOS you can install or upgrade it using Homebrew.

```
brew install bash
brew install bash-completion@2
```

### Install ZSH Completion
The wollemi completion script for Zsh can be generated with the command wollemi
completion zsh. Sourcing the completion script in your shell enables wollemi
command auto-completion.

To do so in all your shell sessions, add the following to your ~/.zshrc file:

```
autoload -Uz compinit && compinit -C
source <(wollemi completion zsh)
compdef _wollemi wollemi
```

After reloading your shell, wollemi autocompletion should be working.

### Vim
Vim can be setup to automatically run wollemi gofmt on file changes by adding
the following line to your vimrc. With this addition, whenever a go file is
written, wollemi gofmt will be automatically run on the package containing the
modified file.

```
autocmd BufWritePost *.go silent exec '!wollemi --log fatal gofmt' shellescape(expand('%:h'), 1)
```

### VS Code 

Thanks to @jamesjarvis wollemi have a VS Code plugin now https://github.com/jamesjarvis/vs-wollemi
VSCode marketplace link: https://marketplace.visualstudio.com/items?itemName=jamesjarvis.vs-wollemi


---

## Commands

### Format
Formats please build files. Formatting modifications include:
  - Double quoted strings instead of single quoted strings.
  - Deduplication of attribute list entries.
  - Ordering of attribute list entries.
  - Ordering of rule attributes.
  - Deletion of empty build files.
  - Consistent build identifiers.
  - Text alignment.

```
Format a specific build file.
    $ wollemi fmt project/service/routes

Recursively format all build files under the routes directory.
    $ wollemi fmt project/service/routes/...

Recursively format all build files under the working directory.
    $ wollemi fmt
```

### Go Format
Rewrites and generates go_binary, go_library and go_test rules according to
existing go code. It also applies all formatting modifications from the
wollemi fmt command.

Wollemi is currently unable to parse build files which contain python
string interpolation. These build files will not be formatted because of
this issue. Also, when the unparseable build file contains go get rules
gofmt will be unable to resolve go dependencies to targets contained in
this build file. To get around the unresolved go dependency issue you can
write a .wollemi.json config file which contains a known dependency mapping
from the unresolvable go package to the correct build target.

```
# project/.wollemi.json
{
  "known_dependency": {
    "go.opencensus.io": "//third_party/go/go.opencensus.io:all_libs"
  }
}
```

Occasionally a go dependency will be able to be resolved to multiple go
get rules and wollemi may choose the wrong target for your needs. These
cases can be resolved using a config file which sets a known dependency
mapping from the go package to the desired target.

```
# project/service/routes/.wollemi.json
{
  "default_visibility": "//project/service/routes/...",
  "known_dependency": {
    "github.com/olivere/elastic": "//third_party/go/github.com/olivere/elastic:v7"
  }
}
```

Config files can be placed in any directory. Every build file gets
formatted using a config which is the result of merging together all
config files discovered between the build file directory and the
directory gofmt was invoked from.

The config file can also define a default visibility. When wollemi gofmt
is invoked recursively on a directory it will use a default visibility
equal to the path it was given on any new go build rules generated. The
visibility of existing go build rules is never modified.

For example, the following gofmt would apply a default visibility of
`["project/service/routes/..."]` to any new go build rules generated.

```
wollemi gofmt project/service/routes/...
```

When gofmt is run on an individual package the default visiblity applied is
`["PUBLIC"]` for any new go build rules generated.

Alternatively the default visiblity can be explicitly provided through
a `.wollemi.json` config file which will override both implicit cases above.
The format and usage of these config files is described in more detail below
in [Configuration](#configuration).

Sometimes a third party dependency is required even though the go code
doesn't directly require it. To force gofmt to keep these dependencies you
must decorate the dependency with the following comment.

```
"//third_party/go/cloud.google.com/go:container", # wollemi:keep
```

The keep comment can also be placed above go build rules you don't want gofmt
to modify. These cases should be rare and this feature should be used only when
absolutely necessary.

```
Go format a specific build file.
    $ wollemi gofmt project/service/routes

Recursively go format all build files under the routes directory.
    $ wollemi gofmt project/service/routes/...

Recursively go format all build files under the working directory.
    $ wollemi gofmt
```

### Rules Unused
Lists potentially unused build rules. Unused in this context simply means no
other build files depend on this rule. User discretion is needed to make the
final call whether an unused build rule listed here should be pruned.

```
List all unused go_get rules.
    $ wollemi rules unused --kind go_get

List all unused rules under the routes directory.
    $ wollemi rules unused project/service/routes/...

List all unused rules except those under k8s and third_party.
    $ wollemi rules unused --exclude k8s,third_party

Prune unused third_party go_get rules.
    $ wollemi rules unused --prune --kind go_get third_party/go/...
```

### Symlink List
Lists and optionally prunes project symlinks. Listed symlinks can be filtered
with --broken in which case only broken symlinks are shown, --name in which
case only symlinks with a name matching the provided pattern will be shown.
For information on what the --name pattern can contain see go doc
path/filepath.Match. Listed symlinks can also be filtered using --exclude
in which case only symlinks which do not have the excluded prefix will be
shown. The --prune flag can be added to any list in which case listed symlinks
are deleted.

```
List all symlinks under the routes directory.
    $ wollemi symlink list project/service/routes/...

List only symlinks in a specific directory.
    $ wollemi symlink list project/service/routes

List all symlinks in the GOPATH. (excludes working directory)
    $ wollemi symlink list --go-path

List all broken symlinks.
    $ wollemi symlink list --broken

List all go_mock symlinks under the routes directory.
    $ wollemi symlink list --name *.mg.go project/service/routes/...

Prune all go_mock symlinks under the routes directory.
    $ wollemi symlink list --prune --name *.mg.go project/service/routes/...
```

### Symlink Go Path
Symlinks third party dependencies into the go path. Symlinks will not be created
when there are existing files in the symlink path. Instead the command will
issue warnings where symlink creation was not possible. When deletion of these
files is acceptable the --force flag can be used to force symlink creation by
first removing the existing files preventing the symlink creation.

```
Symlink all imported third party deps under the routes directory into the go path.
    $ wollemi symlink go-path project/service/routes/...

Symlink all third party deps for specific package into the go path.
    $ wollemi symlink go-path project/service/routes
```

---

### Configuration

The following is an example of a valid `.wollemi.json` config file.
```
{
  "default_visibility": "//project/service/routes/...",
  "allow_unresolved_dependency": true,
  "explicit_sources": true,
  "known_dependency": {
    "github.com/olivere/elastic": "//third_party/go/github.com/olivere/elastic:v7"
  },
  "gofmt": {
    "rewrite": true,
    "create": ["go_binary", "go_library", "go_test"],
    "manage": ["default", "go_custom_binary"]
    "mapped": {
      "go_test": "go_custom_test"
    }
  }
}
```

These config files may be placed in any directory and the settings defined apply
to the package that contains the config file as well as to all sub directory
packages.

```
foo
├── .plzconfig
├── .wollemi.json
├── BUILD.plz
└── bar
    ├── .wollemi.json
    ├── BUILD.plz
    ├── bar.go
    └── baz
        ├── BUILD.plz
        └── baz.go
```

For example, given the package layout above, the config defined in `foo` would
apply to package `foo` as well as packages `bar` and `baz`. In addition, a
config file defined in `bar` inherits settings from the config in `foo`. This
enables the config in `bar` to either override or extend the inherited config.
If the config in bar sets `default_visibility` then it overrides the previous
`default_visibility` whereas each `known_dependency` key value defined either
overrides or adds to the inherited `known_dependency` mapping. This inheritance
continues up the directory chain and stops at the please root directory which
is identified by the existence of a `.plzconfig` file.

The `wollemi gofmt` `--create`, `--manage` and `--mapped` flags, when explicitly
set will override any configuration found on disk.

##### `default_visibility`
  When set all rules created by `wollemi gofmt` will be created using this
  visibility. This does not effect the visibility of any existing rules.

##### `allow_unresolved_dependency`
  Whenever `wollemi gofmt` encounters a dependency it's unable to resolve it logs
  a warning and skips rewriting the rules which required the dependency. When
  this flag is set it unresolved depencies are ignored.

##### `explicit_sources`
  Whenever `wollemi gofmt` creates a new rule it uses glob patterns to define
  a rules srcs. When this flag is set all new rules, as well as all existing
  rules, will explicitly list every src file instead of using glob patterns.

##### `known_dependency`
  Whenever `wollemi gofmt` encounters a dependency it cannot resolve it checks
  to see if a `known_dependency` was manually defined. If a manually defined
  mapping can be found it recovers using the target defined.

##### `gofmt.rewrite`
  Allows `wollemi gofmt` to create new rules and or managing the src files and
  dependencies of existing rules. This is enabled by default but could be
  disabled on a per package basis.

##### `gofmt.create`
  Whitelist of rule kinds allowed to be created by `wollemi gofmt`. By default
  this is `["go_binary", "go_library", "go_test"]`. This can be completely
  disabled by setting it to `[]` or `"off"`. Alternatively this can be
  re-enabled in a child package by setting it to `"on"`, `"default"` or some
  other subset of the default.

##### `gofmt.manage`
  Whitelist of rule kinds allowed to be managed by `wollemi gofmt`. Manage in
  this context means updating an existing rules srcs and/or dependencies
  according to the golang source files. By default this is set to
  `["go_binary", "go_library", "go_test"]`. This can be completely disabled by
  setting it to `[]` or `"off"`. Alternatively this can be re-enabled in a
  child package by setting it to `"on"`, `"default"`, some other list. The
  keyword `"default"` within a list will expand to the original default managed
  rules. Therefore the list `["default", "my_custom_rule"]` is shorthand for
  `["go_binary", "go_library", "go_test", "my_custom_rule"]`.

##### `gofmt.mapped`
  This setting maps standard go rules such as `go_binary`, `go_library` and
  `go_test` to custom go rules. For example, you might define a mapping from
  `go_test` to `go_custom_test`. Defining a mapping like this has two effects.
  First, whenever `wollemi gofmt` determines a package contains go test files
  but lacks a test rule it will create one using `go_custom_test` instead of
  `go_test`. Second, `wollemi gofmt` will manage existing `go_custom_test` rules
  as if they were `go_test` rules instead.
