# Sitekit

> An opinionated server-side rendering framework for Go.

Building traditional web applications in Go often means assembling the same pieces over and over again: routing, templates, rendering, middleware, configuration, static assets, SEO, and error handling.

Sitekit provides those pieces with a consistent structure while staying close to the tools Go already gives you.


---

## Features

* Standard library first
* Server-side rendering with `html/template`
* Layouts and page rendering
* Routing built on `http.ServeMux`
* HTTP and handler middleware
* Environment-based configuration
* Static file serving
* SEO helpers
* Structured HTTP errors
* Minimal, explicit APIs

---

## Installation

```bash
go get github.com/ikorby/sitekit
```

---

## Non-goals

Sitekit intentionally does **not** provide:

* an ORM
* dependency injection
* a custom template language
* a frontend framework
* a replacement for `net/http`

The goal is to build maintainable server-rendered applications.

---

## Status

Sitekit is under active development. While the project is already usable, APIs may continue to evolve before the first stable release.

---

![tests](https://github.com/ikorby/sitekit/actions/workflows/test.yml/badge.svg)