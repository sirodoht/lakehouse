# Server Playbook

## Base updates and user setup

* apt update & upgrade
* reboot
* add deploy user & add to groups

```sh
apt update
apt upgrade
reboot # possible reboot due to kernel upgrade
adduser deploy
adduser deploy www-data
sudo -i -u deploy
```

## PostgreSQL

```sh
apt install postgresql
sudo -i -u postgres
createdb lakehouse
createuser lakehouse
psql
postgres=# ALTER DATABASE lakehouse OWNER TO lakehouse;
postgres=# ALTER USER lakehouse WITH PASSWORD 'xxx';
```

## Caddy

```sh
apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
apt update
apt install caddy
adduser deploy caddy
```

## Golang

```sh
curl -OL https://go.dev/dl/go1.19.5.linux-amd64.tar.gz
sha256sum go1.19.5.linux-amd64.tar.gz
tar -C /usr/local -xvf go1.19.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
sudo -i -u deploy
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

## Node.js

* install node.js lts
* https://github.com/nodesource/distributions#using-debian-as-root-1

```sh
curl -fsSL https://deb.nodesource.com/setup_18.x | bash - && apt-get install -y nodejs
```

## Application

```sh
apt install -y make
cd /var/
mkdir www
chown -R deploy:www-data www/

sudo -i -u deploy
cd /var/www/
git clone https://git.sr.ht/~sirodoht/lakehouse
cd lakehouse
go build
cp .envrc.example .envrc
vim .envrc
exit

cp /var/www/lakehouse/Caddyfile /etc/caddy/
systemctl reload caddy
```

## systemd

```sh
cp /var/www/lakehouse/lakehouse-web.service /lib/systemd/system/
ln -s /lib/systemd/system/lakehouse-web.service /etc/systemd/system/multi-user.target.wants/
systemctl daemon-reload
systemctl enable lakehouse-web.service
systemctl start lakehouse-web.service
systemctl status lakehouse-web.service --no-pager -l
journalctl -u lakehouse-web -f

visudo
# append the following:

# # Allow deploy user to restart apps
# %deploy ALL=NOPASSWD: /usr/bin/systemctl restart lakehouse-web.service
```
