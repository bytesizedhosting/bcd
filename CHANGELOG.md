## Changelog

### Version 0.12.0

- Added (lazy) auto-https support.

Just visit any of the created routes over https and bcd-proxy will automatically attempt to request a certifcate through let's encrypt. This will only work for domains other then the free bysh.io domain since this will hit the rate limits very quick. There is also no support yet for manual requesting/renewing of certificates.

- Ensure config folders exist

BCD now creates the various folders that are used by the Docker container. This reduces the chance that folders are inside of Docker with the wrong permissions.
