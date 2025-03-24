# mr-grep
Application for the concurrent/distributed grepping of text-content.

## Disclaimer
Portions of this codebase were not written by myself.
`mrcoordinator.go`, `mrworker.go`, as well as some signatures in the `mr` file are part of the MIT 6.5840 assignments. However, implementation of these signatures was done by myself. 

In addition, the `Map` and `Reduce` Grep methods were produced by Copilot. 
Implementations of other methods were sometimes debugged with Copilot. 

## How to run
`cd ./main`
`go build -buildmode=plugin ./grep.go`
`go run ./mrcoordinator.go pg-*.txt`
in X other terminals,
`go run ./mrworker.gnoerkrieg.com/mrgrepo grep.so <regex-pattern>`
### Cleanup:
`rm -rf mr-*`
noerkrieg.com/mrgrep