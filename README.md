# lakehouse

Fast docs with real-time collaboration.

## Development

Source code on [sr.ht](https://git.sr.ht/~sirodoht/lakehouse),
mirrored on [GitHub](https://github.com/sirodoht/lakehouse).

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

### Websocket client

This is the websocket client / frontend part of our real-time collaboration editor:

```sh
cd websocket-client/
npm install
npm run build  # one-off bundle.js generation
npm run watch  # develop mood, watch for changes and rebuild
```

## Dependencies

To upgrade dependencies for each service:

```sh
go get -u all
```

```sh
cd websocket-server/
npx ncu -u
npm install
```

```sh
cd websocket-client/
npx ncu -u
npm install
```

## Deployment

```sh
make deploy
```

## License

MIT
