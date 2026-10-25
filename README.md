<p align="center"><img src="docs/logo.svg" width="96" alt="" /></p>

# skillplumb

skillplumb is an offline linter and indexer for folders of agent skills: small directories containing a `SKILL.md` file that coding agents load. It is for skill authors and maintainers who want to catch invalid metadata, missing local links, and repeated descriptions before sharing a skill.

[![CI](https://github.com/Gulliby-development/skillplumb/actions/workflows/ci.yml/badge.svg)](https://github.com/Gulliby-development/skillplumb/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

The program reads the directory you name and files needed to inspect its skills and relative links. It does not modify files or contact a network service.

## Install

Use the Go toolchain version declared in `go.mod`. Clone the repository and install the command from its root:

```sh
git clone https://github.com/Gulliby-development/skillplumb.git
cd skillplumb
go install .
```

Go installs the `skillplumb` executable into `GOBIN`, or otherwise `GOPATH/bin`; add that directory to your `PATH` if needed. skillplumb uses only the Go standard library and has no runtime service or credential requirements.

## Usage

Run `skillplumb check DIR` to lint the immediate visible skill folders under `DIR`. Each skill is expected to contain a regular `SKILL.md`. Findings are printed to standard output, followed by a summary. The command exits with status 1 when it reports errors, or when it reports warnings in strict mode. Command, directory, or argument errors are written to standard error and exit with status 2.

Run `skillplumb index DIR` to print a table of each skill's body line count, local link count, and description code-point count. It prints no findings. Use `skillplumb --help` to print the command synopsis.

The `check` command accepts `--max-desc N` (default 200 code points), `--max-lines N` (default 80 body lines), repeatable `--ignore CODE` to omit findings with a diagnostic code, and `--strict` to make warnings produce exit status 1.

Run all three commands in order from the repository root: the first two create the fixture that the third checks. The `go run .` invocation stands in for running the installed `skillplumb` command; the lines below it are the program's output.

```sh
$ mkdir -p demo/hello
$ printf -- '---\nname: hello\ndescription: Says hello.\n---\nSee [guide](guide.md).\n' > demo/hello/SKILL.md
$ go run . check demo
hello/SKILL.md:5: E09 link guide.md does not exist
skillplumb: 1 skills, 1 errors, 0 warnings
```

## Checks

| Code | Meaning |
| --- | --- |
| E01 | `SKILL.md` is missing or cannot be read. |
| E02 | `SKILL.md` contains NUL or is not valid UTF-8. |
| E03 | Front matter is missing or is not closed by a later `---` line. |
| E04 | A front-matter line is not in the accepted key/value form. |
| E05 | A key is given more than once; the first value is retained. |
| E06 | The name is missing, empty, invalid, or does not match the skill folder. |
| E07 | The description is missing or empty. |
| E08 | The description exceeds the configured code-point limit. |
| E09 | A checked local link does not exist, or a non-final path part is not a folder. |
| E10 | A relative link would leave the skill folder. |
| W01 | A front-matter key is not one of the recognized keys. |
| W02 | A nonblank body has more lines than the configured limit. |
| W03 | The body is empty or contains only blank lines. |
| W04 | The description duplicates an earlier skill's description after ASCII case and blank-space normalization. |

## Development

Run checks from the repository root:

```sh
go test ./...
go vet ./...
gofmt -l .
```

The project has no third-party dependencies. Most tests exercise the command and helpers through the in-memory `MemIo`; the CLI tests also cover the real process boundary.

## License

MIT, see [LICENSE](LICENSE).

---

Built with BAKED
