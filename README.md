# Wollemi
Please build file generator and formatter capable of generating go_binary,
go_library and go_test build rules from existing go code.

## Requirements
- [please](https://please.build)

### Install
In your please project create a file (`tools/wollemi/BUILD.plz`) with following content:

```
build_rule(
    name = "wollemi",
    binary = True,
    srcs = [remote_file(
        name = "wollemi",
        _tag = "download",
        url = f"https://github.com/tcncloud/wollemi/releases/download/v0.0.3/wollemi-v0.0.3-{CONFIG.HOSTOS}-{CONFIG.HOSTARCH}.tar.gz"
    )],
    cmd = " && ".join([
        "tar xf $SRCS",
    ]),
    outs = ["wollemi"],
    visibility = ["PUBLIC"],
)
```
Then you can run it by invoking please: `plz run //tools/wollemi`

Or, just install it in your `GOPATH`:

```
GO111MODULE=on go get github.com/tcncloud/wollemi
```

Wollemi can also be installed by running the following install script from the
root of the repository which builds the binary using please.

```
./install.sh
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
Wollemi can be setup to auto gofmt whenever writing go files with the following
addition to your vimrc. With this addition saving a go file will cause wollemi
gofmt to be automatically run on the package which contains the file.

```
autocmd BufWritePost *.go silent exec '!wollemi gofmt' shellescape(expand('%:h'), 1)
```

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
    "go.opencensus.io": "//third_party/go/go.opencensus.io:all_libs",
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
["project/service/routes/..."] to any new go build rules generated.

```
wollemi gofmt project/service/routes/...
```

When gofmt is run on an individual package the default visiblity applied is
["PUBLIC"] for any new go build rules generated.

Alternatively the default visiblity can be explicitly provided through
a .wollemi.json config file which will override both implicit cases above.

Sometimes a third party dependency is required even though the go code
doesn't directly require it. To force gofmt to keep these dependencies you
must decorate the dependency with the following comment.

```
"//third_party/go/cloud.google.com/go:container",  # wollemi:keep
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
