
# :mag_right: Gosearch :mag:

Gosearch is an extremely simple **search engine** written completely in Go. It was developed based on the architecture of [Tiny Search Engine](https://github.com/allenmqcymp/tse/),
a search engine written in C, developed for Dartmouth College's [Engineering Sciences 50](https://thayer.github.io/engs50/) course.

Gosearch intends to provide the basic functionality of a search engine, and includes a three-part modular structure:
- a concurrent crawler
- a concurrent indexer
- a querier and page ranker that supports fuzzy searching as well as basic boolean operators

The goal of Gosearch is not to provide a comprehensive, cutting-edge, open-source search engine. Rather, Gosearch is a simple 
implementation of the very foundations of a search engine, and aims to take advantage of Go's concurrency primitives in the form of 
channels and goroutines to make the code simple and intuitive, yet performant.

## Todo

- write a fuzzy checker based on some metric, such as Edit distance
- write a go backend and a front-end (probably in React) and deploy to AWS/GCP (dockerized)
- write more tests, unit and integration, and a fuzzer
- godoc, Travis CI integration


