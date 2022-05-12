# ordered [![Test Status](https://github.com/mdlayher/ordered/workflows/Test/badge.svg)](https://github.com/mdlayher/ordered/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/mdlayher/ordered.svg)](https://pkg.go.dev/github.com/mdlayher/ordered)  [![Go Report Card](https://goreportcard.com/badge/github.com/mdlayher/ordered)](https://goreportcard.com/report/github.com/mdlayher/ordered)

Package `ordered` implements data structures which maintain consistent ordering
of inserted elements.

## Note

This package is an experiment to handle use cases where regular Go maps are
beneficial, but deterministic iteration order is also desired. No guarantees are
made about the stability or performance characteristics of this package.
