# lakehousewiki

Not a wiki.

## Development

Source code on [sr.ht](https://git.sr.ht/~sirodoht/lakehousewiki),
mirrored on [GitHub](https://github.com/sirodoht/lakehousewiki).

### Database

We use PostgreSQL and Nix:

```
cd postgresql/
make init  # only the first time
make start
```

### Webserver

We use [modd](https://github.com/cortesi/modd) to autoreload:

```sh
make serve
```

### Websocket server

We need this for real-time collaboration:

```sh
cd websocket-server/
npm install
npm start
```

### Frontend editor

This is the frontend part of our real-time collaboration editor:

```sh
cd editor/
npm install
npm run build  # one-off bundle.js generation
npm run watch  # develop mood, watch for changes and rebuild
```

## Deployment

```sh
make deploy
```

## License

MIT
