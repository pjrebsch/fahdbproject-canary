#!/bin/bash

# Since there are multiple main package files, we would need to list them
# all when using `go run`. We can work around that by building the package
# and then running it.

go build github.com/pjrebsch/fahdbproject-canary

./fahdbproject-canary
