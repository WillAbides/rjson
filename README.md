# rjson

[![godoc](https://pkg.go.dev/badge/github.com/willabides/rjson.svg)](https://pkg.go.dev/github.com/willabides/rjson)
[![ci](https://github.com/WillAbides/rjson/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/WillAbides/rjson/actions?query=workflow%3Aci+branch%3Amain+event%3Apush)

rjson is a json parser that relies on [Ragel-generated](http://www.colm.net/open-source/ragel/) state machines for most
parsing. rjson's api is minimal and focussed on efficient parsing.
