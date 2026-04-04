# ViBrowsing

A from-scratch web browser written in Go, with no external browser engine libraries. Because sometimes the best way to understand how browsers work is to build one badly and watch it render things that are mostly... approximately... right.

> "Vi" — from "view" / "visibility" / "vi" (the editor you either love or tolerate)
> "Browsing" — the thing browsers do

## What

A toy browser project that aims to render web pages using its own HTML/CSS engine, built entirely from scratch in Go:
- HTTP fetch via stdlib
- HTML5 tokenizer (state machine, no regex)
- DOM tree builder
- CSS parser + cascaded style computation
- CSS box model layout engine (block layout to start)
- Software rasterization to pixel buffer
- Ebitengine for window chrome and input (planned)

JavaScript is out of scope. Flexbox/Grid is out of scope. Production use is... also out of scope.

## Why

Building a browser from scratch is the best way to understand every layer of the web platform. Also, it's fun. Also, we had some spare tokens and a mid-tier AMD laptop.

## Architecture

```
URL Input
    ↓
net/http fetch
    ↓
HTML5 Tokenizer → DOM Tree
    ↓
CSS Parser → Cascaded Styles
    ↓
Box Layout Engine → Render Tree
    ↓
Pixel Buffer → PNG (CLI) or Ebitengine window (GUI)
```

## Name

"ViBrowsing" — part browser, part eternal debate about whether to use vi or not.

## Status

Planning phase. See [PLAN.md](PLAN.md) for the full roadmap.

## Building

```bash
go mod tidy
go run ./cmd/browser [url]
```

Output is a PNG file called `output.png`.
