# proxy-fix

This simple app simply spawns another HTTP server with `PORT` env injected with another number, then proxying it with [a clean header](https://github.com/domcloud/proxy-fix/blob/cb40ba29e5ac592438848b3071f2137ea9b3e0b6/main.go#L132-L140) request.

Built primarily for fixing https://github.com/phusion/passenger/issues/2521 temporarily. 

Yes, this proxy doesn't support Websocket. Pull requests are welcome.

## Install

Download from releases or build it and place it to `~/.local/bin/bunfix`

## Usage

```sh
env PORT=4000 bunfix bun app.ts
```
