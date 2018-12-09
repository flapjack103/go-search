# Go-Search

## Description
TODO

## Examples

```
COMP10771:go-search alexandra.grant$ ./go-search
Constructing index on 8 files...
Building prefix tree...
Initializing search
Please use `exit` or `Ctrl-D` to exit this program.
>>> help
Commands: search, list

For more information try 'man search' etc.
>>> man search
usage: search <word> [options]

    Description: The search command searches go source files for the given word. Works
                 on partial matches (eg. 'err' will return instances of 'err' and also
                  'error') making it equivalent to <word>*.

    The options are as follows:

    -l    Number of results to display (limit). Default is 20.

    -s    Sort type which determines how results are ranked. Available sorting types
          are 'lex' (lexicographically), 'count', and 'rel' (relevance). Default is
          'lex.'

    -f    Filter on a subset of files to search within. Paths must be absolute or
          relative to current directory. Pass as a comma seperated list. No filter
          by default.

    Example: search err -l 10 -s rel -f main.go

>>> search err
+---+------+-------------+-------+
| # | WORD | FILE        | COUNT |
+---+------+-------------+-------+
| 1 | err  | index.go    |     1 |
| 2 | err  | executor.go |     3 |
+---+------+-------------+-------+
>>> list err
+---+-------------+------+--------+
| # | FILE        | LINE | COLUMN |
+---+-------------+------+--------+
| 1 | executor.go |  104 |      5 |
| 2 | executor.go |  133 |     71 |
| 3 | executor.go |  178 |      8 |
| 4 | index.go    |   31 |      6 |
+---+-------------+------+--------+
>>> search t
+---+--------+--------------+-------+
| # | WORD   | FILE         | COUNT |
+---+--------+--------------+-------+
| 1 | t      | index.go     |     1 |
| 2 | t      | reference.go |     1 |
| 3 | t      | results.go   |     1 |
| 4 | t      | trie_test.go |     2 |
| 5 | t      | trie.go      |     5 |
| 6 | tri    | trie_test.go |     2 |
| 7 | trie   | querier.go   |     1 |
| 8 | trie   | main.go      |     1 |
| 9 | twords | trie_test.go |     2 |
+---+--------+--------------+-------+
>>> exit
Bye!
```
