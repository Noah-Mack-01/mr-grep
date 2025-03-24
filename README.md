# mr-grep
Application for the concurrent/distributed grepping of text-content.

## Disclaimer
Portions of this codebase were not written by myself.
`mrcoordinator.go`, `mrworker.go`, as well as some signatures in the `mr` file are part of the MIT 6.5840 assignments. However, implementation of these signatures was done by myself and most of the copied code is now overwritten or refactored. 

In addition, the `Map` and `Reduce` Grep methods were produced by Copilot. 
Implementations of other methods were sometimes debugged with Copilot. 

## How to run
`bash build.sh`
`bash run.sh`
### Cleanup:
`rm -rf ./output/mr-*`
`rm -rf ./worker/mr-*`
noerkrieg.com/mrgrep