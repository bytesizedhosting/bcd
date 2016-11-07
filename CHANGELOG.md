## Changelog

### Version 0.14.1

- Changed rTorrent app to expose all ports by default and no longer use Dockerfile port bindings
- Added filebot.sh to the /config folder instead of /app so users can make any modifications they want to the filebot cli arguments

### Version 0.14.0

- New Appplication added: Filebot.

Filebot uses inotify to watch a folder for filesystem events, runs on the input and outputs to the given outfolder folder.

### Version 0.13.0

- New Application added: Murmur, the server for the Mumble voice client.
- Enable auto-https by default

You can disable the support by using the --disable-auto-https flag.

### Version 0.12.0

- Added (lazy) auto-https support.

Just visit any of the created routes over https and bcd-proxy will automatically attempt to request a certifcate through let's encrypt. This will only work for domains other then the free bysh.io domain since this will hit the rate limits very quick. There is also no support yet for manual requesting/renewing of certificates.

This feature is disabled by default until it is tested more you can enable it using bcd-proxy with the `--enable-auto-https` flag.

- Ensure config folders exist

BCD now creates the various folders that are used by the Docker container. This reduces the chance that folders are inside of Docker with the wrong permissions.
