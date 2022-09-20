#!/usr/local/bin/bash

set -e
set -x

# push origin
git push origin master

# overwrite nginx config
scp lakehouse.sirodoht.com.conf root@lakehouse.wiki:/etc/nginx/sites-available/

# pull and reload on server
ssh root@lakehouse.wiki 'cd /opt/apps/lakehouse \
    && git pull \
    && systemctl reload nginx'
