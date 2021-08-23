# fuzzy

[![Build Status](https://img.shields.io/circleci/build/github/isacikgoz/fuzzy/master)](https://app.circleci.com/pipelines/github/isacikgoz/fuzzy)
[![Documentation](https://godoc.org/github.com/isacikgoz/fuzzy?status.svg)](https://pkg.go.dev/github.com/isacikgoz/fuzzyy)

Originally developed by [Sahil Muthoo](https://github.com/sahilm), `fuzzy` is a fuzzy search library that provides extensive searching for strings. It is optimized for filenames and code symbols. This library is external dependency-free. It only depends on the Go standard library. This fork is an parallel version of its original repository.

## Features

- Intuitive matching. Results are returned in descending order of match quality. Quality is determined by:
  - The first character in the pattern matches the first character in the match string.
  - The matched character is camel cased.
  - The matched character follows a separator such as an underscore character.
  - The matched character is adjacent to a previous match.

- Speed. Matches are returned in milliseconds. It's perfect for interactive search boxes.

- The positions of matches are returned. Allows you to highlight matching characters.

- Unicode aware.

- Works parallel.

## Usage

The following example prints out matches with the matched chars in bold.

```go
package main

import (
	"fmt"
	"sort"

	"github.com/isacikgoz/fuzzy"
)

func main() {
	const bold = "\033[1m%s\033[0m"
	pattern := "mnr"
	data := []string{"game.cpp", "moduleNameResolver.ts", "my name is_Ramsey"}

	results := Find(context.Background(),pattern, data)

	matches := make([]Match, 0)
	for result := range results {
		matches = append(matches, result)
	}

	sort.Stable(fuzzy.Sortable(matches))

	for _, match := range matches {
		for i := 0; i < len(match.Str); i++ {
			if contains(i, match.MatchedIndexes) {
				fmt.Print(fmt.Sprintf(bold, string(match.Str[i])))
			} else {
				fmt.Print(string(match.Str[i]))
			}
		}
		fmt.Println()
	}
}

func contains(needle int, haystack []int) bool {
	for _, i := range haystack {
		if needle == i {
			return true
		}
	}
	return false
}
```

Check out the [godoc](https://godoc.org/github.com/isacikgoz/fuzzy) for detailed documentation.

## Installation

`go get github.com/isacikgoz/fuzzy`

## Speed

Here are a few benchmark results on a normal laptop.

```
BenchmarkFind/with_unreal_4_(~16K_files)-12         	     171	   6941671 ns/op	   21341 B/op	     886 allocs/op
BenchmarkFind/with_linux_kernel_(~60K_files)-12     	      76	  15827954 ns/op	    6286 B/op	     195 allocs/op
```

## Contributing

Everyone is welcome to contribute. Please send me a pull request or file an issue.

## Credits

* [@sahilm](https://github.com/sahilm) for the original project.

* [@ericpauley](https://github.com/ericpauley) & [@lunixbochs](https://github.com/lunixbochs) contributed Unicode awareness and various performance optimisations.

* The algorithm is based of the awesome work of [forrestthewoods](https://github.com/forrestthewoods/lib_fts/blob/master/code/fts_fuzzy_match.js). 
See [this](https://blog.forrestthewoods.com/reverse-engineering-sublime-text-s-fuzzy-match-4cffeed33fdb#.d05n81yjy)
blog post for details of the algorithm.
