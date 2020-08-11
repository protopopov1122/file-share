#!/usr/bin/env sh

# Paths
FS_ROOT="/file-share"
FS_CONFIG="$FS_ROOT/config"
FS_STORAGE="$FS_ROOT/storage"
FS_USER="$FS_ROOT/user"
FS_CONFIG_PERSISTENT="$FS_CONFIG/persistent"
FS_CONFIG_RUNTIME="$FS_CONFIG/runtime"

# Wipe previous configuration
/bin/rm -rf "$FS_USER/.ssh"
/bin/rm -rf "$FS_CONFIG/api.json"

# Configure permissions and directory structure
/bin/chown -R www:www "$FS_STORAGE"
/bin/mkdir "$FS_USER/.ssh"
/bin/touch "$FS_USER/.ssh/authorized_keys"
/bin/chown upload:upload -R "$FS_USER"
/bin/chmod 700 "$FS_USER/.ssh"
/bin/chmod 600 "$FS_USER/.ssh/authorized_keys"

# Copy configuration
if [ -f "$FS_CONFIG_RUNTIME/api.json" ]; then
    /bin/ln -sf "$FS_CONFIG_RUNTIME/api.json" "$FS_CONFIG/api.json"
else
    /bin/ln -sf "$FS_CONFIG_PERSISTENT/api.json" "$FS_CONFIG/api.json"
fi

function add_keys {
    if [ -d "$1" ]; then
        find "$1" -type f -maxdepth 1 -exec cat {} \; >> "$FS_USER/.ssh/authorized_keys"
    fi
}

echo > "$FS_USER/.ssh/authorized_keys"
add_keys "$FS_CONFIG_PERSISTENT/keys"
add_keys "$FS_CONFIG_RUNTIME/keys"

function stopService {
    /usr/bin/pkill -SIGINT -f supervisord
    exit 0
}

trap stopService SIGINT

# Start applications
/usr/bin/supervisord >/dev/null 2>&1
/usr/bin/supervisorctl status all
/usr/bin/supervisorctl tail -f file-share &
wait
