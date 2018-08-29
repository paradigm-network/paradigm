<h1 align="center">Paradigm </h1>
<h4 align="center">Version 0.1 </h4>

[![GoDoc](https://godoc.org/github.com/paradigm-network/paradigm?status.svg)](https://godoc.org/github.com/paradigm-network/paradigm)
[![Go Report Card](https://goreportcard.com/badge/github.com/paradigm-network/paradigm)](https://goreportcard.com/report/github.com/paradigm-network/paradigm)
[![Travis](https://travis-ci.org/paradigm/paradigm.svg?branch=master)](https://travis-ci.org/paradigm/paradigm)
[![Discord](https://img.shields.io/discord/102860784329052160.svg)](https://discord.gg/kU3ewZQ)

Welcome to Paradigm Network !

The code is currently alpha quality, but is in the process of rapid development. The master code may be unstable; stable versions can be downloaded in the release page.

## Build development environment
The requirements to build Paradigm are:

- Golang version 1.9 or later
- Glide (a third party package management tool)
- Properly configured Go language environment
- Golang supported operating system

## Get Paradigm
### Get from source code

Clone the Paradigm repository into the appropriate $GOPATH/src/github.com/paradigm-network directory.

```
$ git clone --recursive https://github.com/paradigm-network/paradigm.git
```
or
```
$ go get github.com/paradigm-network/paradigm
```
Fetch the dependent third party packages with glide.

```
$ cd $GOPATH/src/github.com/paradigm-network/paradigm
$ make get_vendor_deps
$ make install
```

Build the source code with make.

```
$ make all
```

After building the source code sucessfully, you should see two executable programs:

- `paradigm`: the node program/command line program for paradigm node control

### get from release
You can download at [release page](https://github.com/paradigm-network/paradigm/releases).

## Contributions

Please open a pull request with a signed commit. We appreciate your help! You can also send your code as emails to the developer mailing list. You're welcome to join the Paradigm mailing list or developer forum.

Please provide detailed submission information when you want to contribute code for this project. The format is as follows:

Header line: explain the commit in one line (use the imperative).

Body of commit message is a few lines of text, explaining things  in more detail, possibly giving some background about the issue  being fixed, etc.

The body of the commit message can be several paragraphs. Please do proper word-wrap and keep columns shorter than 74 characters or so. That way "git log" will show things  nicely even when it is indented.

Make sure you explain your solution and why you are doing what you are  doing, as opposed to describing what you are doing. Reviewers and your  future self can read the patch, but might not understand why a  particular solution was implemented.

Reported-by: whoever-reported-it &
Signed-off-by: Your Name [youremail@yourhost.com](mailto:youremail@yourhost.com)

## Open source community
### Site

- <https://paradigm.fund/>

### Developer Discord Group

- <https://discord.gg/kU3ewZQ/>

## License

The Paradigm library is licensed under the GNU Lesser General Public License v3.0, read the LICENSE file in the root directory of the project for details.
