# go-self-delete
Go implementation of the self-deletion of an running executable from disk.

> **DISCLAIMER.** All information contained in this repository is provided for educational and research purposes only. The owner is not responsible for any illegal use of included code snippets.

## Usage

```go
package main

import "github.com/secur30nly/go-self-delete"

func main() {
	// some code ...

	selfdelete.SelfDeleteExe()

	// some code ...
}

```

## References

- https://github.com/LloydLabs/delete-self-poc
- https://twitter.com/jonasLyk/status/1350401461985955840
