//go:build !android

package main

func readSharedFromCache() (string, error) { return "", nil }

