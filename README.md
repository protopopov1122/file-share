### Simple file share service

The goal of this project is providing file share which can be deployed as a single Docker container. File uploaders are restricted through SSH keys, downloaders can be anonymous. The project is primarily intended for personal use.

Components:
* File storage service - serves saved files, receives new uploads, collects garbage.
* File uploader - processes input parameters and pipes stdin into the storage.
* SSH service - exposes uploader as a user shell with key-based authentication.
* Nginx configuration - limits access to uploading functionality outside the container.
* Dockerfile

#### Author & License
Author: Jevgēnijs Protopopovs

License: MIT