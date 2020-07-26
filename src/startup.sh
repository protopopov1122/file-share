#!/usr/bin/env sh
/bin/chown -R www:www /file-share/storage
/usr/bin/supervisord >/dev/null 2>&1
/usr/bin/supervisorctl status all
/usr/bin/supervisorctl tail -f file-share
